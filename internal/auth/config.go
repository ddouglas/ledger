package auth

type configOption func(*config)

type config struct {
	JWKSURI  string
	Audience *string
	Issuer   *string
}

func WithJWKSURI(uri string) configOption {
	return func(c *config) {
		c.JWKSURI = uri
	}
}

func WithAudience(aud string) configOption {
	return func(c *config) {
		c.Audience = toStringPointer(aud)
	}
}

func WithIssuer(iss string) configOption {
	return func(c *config) {
		c.Issuer = toStringPointer(iss)
	}
}

func toStringPointer(s string) *string {
	return &s
}
