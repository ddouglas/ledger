package importer

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/go-redis/redis"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run(ctx context.Context)
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

func (s *service) Run(ctx context.Context) {

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

		data, err := s.redis.LPop(ctx, gateway.PubSubPlaidWebhook).Result()
		if err != nil && err.Error() != redis.Nil {
			s.logger.WithError(err).Error("failed to fetch messages from queue")
			return
		}

		if err != nil && err.Error() == redis.Nil {
			s.logger.Info("received nil, going to sleep")
			time.Sleep(time.Second * 10)
			continue
		}

		var message = new(WebhookMessage)
		err := json.Unmarshal([]byte(data), message)
		if err != nil {
			entry.WithError(err).Error("failed to decode message")
			continue
		}

		s.processMessage(ctx, message)

	}

}

func (s *service) processMessage(ctx context.Context, message *WebhookMessage) {

	txn := s.newrelic.StartTransaction("Process Webhook Message")
	defer txn.End()
	ctx = newrelic.NewContext(ctx, s.newrelic.StartTransaction("Process Webhook Message"))
	_ = s.logger.WithContext(ctx).WithField("service", "Importer")

	switch message.WebhookType {
	case "TRANSACTIONS":
		s.processTransactionUpdate(ctx, message)
	default:
		s.logger.WithField("message_type", message.WebhookType).Error("recieved message with unhandled webhook type")
	}

}

func (s *service) processTransactionUpdate(ctx context.Context, message *WebhookMessage) {

	existingItem, err := s.item.Item(ctx, message.ItemID)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch item with itemID provided by message")
		return
	}

	item, err := s.gateway.Item(ctx, existingItem.AccessToken)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch plaid item with accessToken")
		return
	}

	item.UserID = existingItem.UserID
	_, err = s.item.UpdateItem(ctx, item.ItemID, item)
	if err != nil {
		s.logger.WithError(err).Error("failed to update item")
		return
	}

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

	transactions, err := s.gateway.Transactions(ctx, item.AccessToken, start, end)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch transactions")
		return
	}

	err = s.transaction.ProcessTransactions(ctx, item, transactions)
	if err != nil {
		s.logger.WithError(err).Error("failed to process transactions")
	}

	s.logger.Info("transactions processed successfully")

}
