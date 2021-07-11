package importer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/go-redis/redis/v8"
)

type Service interface {
	Run(ctx context.Context)
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

	pubsub := s.redis.Subscribe(ctx, gateway.PubSubPlaidWebhook)

	ch := pubsub.Channel(
		redis.WithChannelHealthCheckInterval(time.Second*15),
		redis.WithChannelSendTimeout(time.Second*30),
		redis.WithChannelSize(10),
	)

	entry := s.logger.WithField("service", "Gateway")

	for m := range ch {
		data := []byte(m.String())
		var hook = new(WebhookMessage)
		err := json.Unmarshal(data, hook)
		if err != nil {

		}
	}

}
