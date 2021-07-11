package gateway

type configOption func(c *config)

type config struct {
	products     []string
	language     *string
	webhook      *string
	countryCodes []string
}

func WithProducts(products ...string) configOption {
	return func(c *config) {
		c.products = products
	}
}

func WithLanguage(lang string) configOption {
	return func(c *config) {
		c.language = toStringPointer(lang)
	}
}

func WithWebhook(hook string) configOption {
	return func(c *config) {
		c.webhook = toStringPointer(hook)
	}
}

func WithCountryCode(codes ...string) configOption {
	return func(c *config) {
		c.countryCodes = codes
	}
}

func toStringPointer(s string) *string {
	return &s
}
