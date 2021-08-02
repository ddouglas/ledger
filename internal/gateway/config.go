package gateway

import (
	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/cache"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
)

const (
	PubSubPlaidWebhook = "plaid-webhook"
)

type service struct {
	newrelic     *newrelic.Application
	cache        cache.Service
	client       *plaid.Client
	logger       *logrus.Logger
	products     []string
	language     *string
	webhook      *string
	countryCodes []string

	ledger.PlaidRepository
}

type configOption func(s *service)

func WithNewrelicApplication(newrelic *newrelic.Application) configOption {
	return func(s *service) {
		s.newrelic = newrelic
	}
}

func WithPlaidRepository(plaid ledger.PlaidRepository) configOption {
	return func(s *service) {
		s.PlaidRepository = plaid
	}

}

func WithCache(cache cache.Service) configOption {
	return func(s *service) {
		s.cache = cache
	}
}

func WithLogger(logger *logrus.Logger) configOption {
	return func(s *service) {
		s.logger = logger
	}
}

func WithProducts(products ...string) configOption {
	return func(s *service) {
		s.products = products
	}
}

func WithPlaidClient(client *plaid.Client) configOption {
	return func(s *service) {
		s.client = client
	}
}

func WithLanguage(lang string) configOption {
	return func(s *service) {
		s.language = toStringPointer(lang)
	}
}

func WithWebhook(hook string) configOption {
	return func(s *service) {
		if hook != "" {
			s.webhook = toStringPointer(hook)
		}
	}
}

func WithCountryCode(codes ...string) configOption {
	return func(s *service) {
		s.countryCodes = codes
	}
}

func toStringPointer(s string) *string {
	return &s
}
