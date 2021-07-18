package importer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run()
	VerifyWebhookMessage(ctx context.Context, header http.Header, message []byte) error
	PublishWebhookMessage(ctx context.Context, webhook *WebhookMessage) error
}

func New(optFucs ...configOption) Service {
	s := &service{}
	for _, optFunc := range optFucs {
		optFunc(s)
	}
	return s
}

func (s *service) Run() {

	// pubsub := s.redis.Subscribe(ctx, gateway.PubSubPlaidWebhook)

	// ch := pubsub.Channel(
	// 	redis.WithChannelHealthCheckInterval(time.Second*15),
	// 	redis.WithChannelSendTimeout(time.Second*30),
	// 	redis.WithChannelSize(10),
	// )

	entry := s.logger.WithFields(logrus.Fields{
		"service": "Importer",
		"channel": gateway.PubSubPlaidWebhook,
	})
	entry.Info("Subscibed to Redis Pubsub")
	for {
		txn := s.newrelic.StartTransaction("check-plaid-message-queue")
		ctx := newrelic.NewContext(context.Background(), txn)
		entry := s.logger.WithContext(ctx)
		entry.Info("checking message queue")

		data, err := s.redis.LPop(ctx, gateway.PubSubPlaidWebhook).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			entry.WithError(err).Error("failed to fetch messages from queue")
			txn.NoticeError(err)
			sleep()
			continue
		}

		if err != nil && errors.Is(err, redis.Nil) {
			entry.Info("received nil, going to sleep")
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
	time.Sleep(time.Second * 10)
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

	seg = txn.StartSegment("updating item")
	item.UserID = existingItem.UserID
	_, err = s.item.UpdateItem(ctx, item.ItemID, item)
	if err != nil {
		entry.WithError(err).Error("failed to update item")
		return
	}
	seg.End()

	seg = txn.StartSegment("evaluating webhook code")
	var start, end time.Time
	switch message.WebhookCode {
	case "INITIAL_UPDATE", "DEFAULT_UPDATE":
		start = time.Now().AddDate(0, 0, -30)
		end = time.Now()
	case "HISTORICAL_UPDATE":
		start = time.Now().AddDate(-2, 0, 0)
		end = time.Now().AddDate(0, 0, -30)
	// case "DEFAULT_UPDATE":
	// 	start = time.Now().AddDate(0, 0, 0)
	// 	end = time.Now()
	case "TRANSACTIONS_REMOVED":
		// How to handle this, thinking about calling a seperate func
		// and then returning here instead of allowing the func to continue processing
	default:
		// unhandled webhook code received
	}
	seg.AddAttribute("startDate", start.Format("2006-01-02"))
	seg.AddAttribute("endDate", start.Format("2006-01-02"))
	seg.End()

	seg = txn.StartSegment("fetching transactions from plaid")
	transactions, err := s.gateway.Transactions(ctx, item.AccessToken, start, end)
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

	entry.Info("transactions processed successfully")

}
