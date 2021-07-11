package gateway

import "github.com/plaid/plaid-go/plaid"

const (
	PubSubPlaidWebhook = "plaid-webhook"
)

type configOption func(s *service)

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
		s.webhook = toStringPointer(hook)
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
