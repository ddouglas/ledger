package importer

import (
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type configOption func(s *service)

type service struct {
	redis    *redis.Client
	gateway  gateway.Service
	logger   *logrus.Logger
	newrelic *newrelic.Application
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
