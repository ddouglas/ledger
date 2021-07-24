package importer

import (
	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/account"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/ddouglas/ledger/internal/item"
	"github.com/ddouglas/ledger/internal/transaction"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type configOption func(s *service)

type service struct {
	account     account.Service
	item        item.Service
	transaction transaction.Service

	redis    *redis.Client
	gateway  gateway.Service
	logger   *logrus.Logger
	newrelic *newrelic.Application

	ledger.WebhookRepository
}

func WithWebhookRepository(webhook ledger.WebhookRepository) configOption {
	return func(s *service) {
		s.WebhookRepository = webhook
	}
}

func WithNewrelic(newrelic *newrelic.Application) configOption {
	return func(s *service) {
		s.newrelic = newrelic
	}
}

func WithLogger(logger *logrus.Logger) configOption {
	return func(s *service) {
		s.logger = logger
	}
}

func WithRedis(client *redis.Client) configOption {
	return func(s *service) {
		s.redis = client
	}
}

func WithGateway(gateway gateway.Service) configOption {
	return func(s *service) {
		s.gateway = gateway
	}
}

func WithAccounts(account account.Service) configOption {
	return func(s *service) {
		s.account = account
	}
}

func WithItems(item item.Service) configOption {
	return func(s *service) {
		s.item = item
	}
}

func WithTransactions(transaction transaction.Service) configOption {
	return func(s *service) {
		s.transaction = transaction
	}
}
