package importer

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"
)

type Service interface {
	Run()
	VerifyWebhookMessage(ctx context.Context, header http.Header, message []byte) error
	PublishWebhookMessage(ctx context.Context, webhook *WebhookMessage) error
	PublishCustomWebhookMessage(ctx context.Context, webhook *WebhookMessage) error
}

func New(optFucs ...configOption) Service {
	s := &service{}
	for _, optFunc := range optFucs {
		optFunc(s)
	}
	return s
}

func (s *service) Run() {

	entry := s.logger.WithFields(logrus.Fields{
		"service": "Importer",
		"channel": gateway.PubSubPlaidWebhook,
	})
	entry.Info("Subscibed to Redis Pubsub")
	for {
		txn := s.newrelic.StartTransaction("check-plaid-message-queue")
		ctx := newrelic.NewContext(context.Background(), txn)
		entry := s.logger.WithContext(ctx)
		entry.Debug("checking message queue")

		data, err := s.redis.LPop(ctx, gateway.PubSubPlaidWebhook).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			entry.WithError(err).Error("failed to fetch messages from queue")
			txn.NoticeError(err)
			sleep()
			continue
		}

		if err != nil && errors.Is(err, redis.Nil) {
			entry.Debug("received nil, going to sleep")
			txn.Ignore()
			sleep()
			continue
		}

		entry.WithField("message", data).Info("webhook received, dispatching processor")
		var message = new(WebhookMessage)
		err = json.Unmarshal([]byte(data), message)
		if err != nil {
			entry.WithError(err).Error("failed to decode message")
			txn.NoticeError(err)
			continue
		}

		s.processMessage(ctx, message)
		s.logger.Info("message processed successfully")
		txn.End()
		sleep()

	}

}

func sleep() {
	time.Sleep(time.Second * 1)
}

func (s *service) processMessage(ctx context.Context, message *WebhookMessage) {

	switch message.WebhookType {
	case "TRANSACTIONS":
		s.processTransactionUpdate(ctx, message)
	default:
		s.logger.WithContext(ctx).WithField("message", message).Error("recieved message with unhandled webhook type")
	}

}

func (s *service) processTransactionUpdate(ctx context.Context, message *WebhookMessage) {

	txn := newrelic.FromContext(ctx)
	entry := s.logger.WithContext(ctx)

	seg := txn.StartSegment("checking for existing item")
	existingItem, err := s.item.Item(ctx, message.ItemID)
	if err != nil {
		entry.WithError(err).Error("failed to fetch item with itemID provided by message")
		return
	}
	seg.End()

	seg = txn.StartSegment("fetching updated item from plaid")
	item, err := s.gateway.Item(ctx, existingItem.AccessToken)
	if err != nil {
		entry.WithError(err).Error("failed to fetch plaid item with accessToken")
		return
	}
	seg.End()

	err = deepcopier.Copy(item).To(existingItem)
	if err != nil {
		entry.WithError(err).Error("failed to copy plaid item to ledger item")
		return
	}

	seg = txn.StartSegment("updating item")
	_, err = s.item.UpdateItem(ctx, existingItem.ItemID, existingItem)
	if err != nil {
		entry.WithError(err).Error("failed to update item")
		return
	}
	seg.End()

	seg = txn.StartSegment("updating accounts")
	accounts, err := s.gateway.Accounts(ctx, existingItem.AccessToken)
	if err != nil {
		entry.WithError(err).Error("failed to update item")
		return
	}

	for _, account := range accounts {
		account.ItemID = existingItem.ItemID
		_, err = s.account.UpdateAccount(ctx, existingItem.ItemID, account.AccountID, account)
		if err != nil {
			entry.WithError(err).WithField("account_id", account.AccountID).Error("failed to update account")
			return
		}
	}
	seg.End()

	seg = txn.StartSegment("evaluating webhook code")
	var start, end time.Time
	var accountIDs []string
	switch message.WebhookCode {
	case "INITIAL_UPDATE":
		start = time.Now().AddDate(0, 0, -30)
		end = time.Now()
	case "HISTORICAL_UPDATE":
		start = time.Now().AddDate(-2, 0, 0)
		end = time.Now().AddDate(0, 0, -30)
	case "DEFAULT_UPDATE":
		start = time.Now().AddDate(0, 0, 0)
		end = time.Now()
	case "CUSTOM_UPDATE":
		start = message.StartDate
		end = message.EndDate
		if message.Options != nil && len(message.Options.AccountIDs) > 0 {
			accountIDs = message.Options.AccountIDs
		}
	case "TRANSACTIONS_REMOVED":
		// How to handle this, thinking about calling a seperate func
		// and then returning here instead of allowing the  func to continue processing
	default:
		// unhandled webhook code received
	}
	seg.AddAttribute("startDate", start.Format("2006-01-02"))
	seg.AddAttribute("endDate", end.Format("2006-01-02"))
	seg.End()

	seg = txn.StartSegment("fetching transactions from plaid")
	transactions, err := s.gateway.Transactions(ctx, item.AccessToken, start, end, accountIDs)
	if err != nil {
		entry.WithError(err).Error("failed to fetch transactions")
		return
	}
	seg.AddAttribute("transactionCount", len(transactions))
	seg.End()

	seg = txn.StartSegment("processing transactions")
	err = s.transaction.ProcessTransactions(ctx, item, transactions)
	if err != nil {
		entry.WithError(err).Error("failed to process transactions")
		return
	}
	seg.End()

	if existingItem.IsRefreshing {
		entry.Info("Item is in refreshing state, updating to false")
		existingItem.IsRefreshing = false
		_, err = s.item.UpdateItem(ctx, existingItem.ItemID, existingItem)
		if err != nil {
			entry.WithError(err).Error("failed to toggle isRefreshing flag on item")
			return
		}
	}

	entry.Info("transactions processed successfully")

}
