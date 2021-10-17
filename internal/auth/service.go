package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ddouglas/ledger/internal/cache"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/oauth2"
)

type Service interface {
	ExchangeCode(ctx context.Context, code string) (string, jwt.Token, error)
	ValidateToken(ctx context.Context, t string) (jwt.Token, error)
}

type service struct {
	client *http.Client
	cache  cache.Service
	oauth  *oauth2.Config
	config *config
}

type config struct {
	JWKSURI  string
	Audience string
	Issuer   string
}

func New(cache cache.Service, client *http.Client, oauth *oauth2.Config, jwksURI, audience, issuer string) Service {
	return &service{
		client: client,
		cache:  cache,
		oauth:  oauth,
		config: &config{
			JWKSURI:  jwksURI,
			Audience: audience,
			Issuer:   issuer,
		},
	}
}

// ExchangeCode exchanges an Auth0 callback code for a JWT Token and then validates that the
// token is valid and no MITM injection happened resulting in an invalid token
func (s *service) ExchangeCode(ctx context.Context, code string) (string, jwt.Token, error) {

	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.client)
	bearer, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	parsed, err := s.ValidateToken(ctx, bearer.AccessToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse and verify token: %w", err)
	}

	return bearer.AccessToken, parsed, nil

}

// ValidateToken validates that t is a valid JWT. It fetch a JWKS from cache and checks
// to see if various config options are set before calling the lestrrat-go JWT lib
// to validate T
func (s *service) ValidateToken(ctx context.Context, t string) (jwt.Token, error) {

	set, err := s.getKeySet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWK Set: %w", err)
	}

	token, err := jwt.ParseString(
		t,
		jwt.WithKeySet(set),
		jwt.WithAudience(s.config.Audience),
		jwt.WithIssuer(s.config.Issuer),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return token, nil

}

// getKeySet fetches the JWKS from cache. If the cache does not have
// a JWKS to return, we attempt to fetch it from Auth0 and cache
// it again. The timeout is handled by the cache package, but is generally
// set to 6 hours
func (s *service) getKeySet(ctx context.Context) (jwk.Set, error) {

	b, err := s.cache.JWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("unexpected error occured querying redis for jwks: %w", err)
	}

	if b == nil {
		res, err := s.client.Get(s.config.JWKSURI)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve jwks from sso: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code recieved while fetching jwks. %d", res.StatusCode)
		}

		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read jwk response body: %w", err)
		}

		err = s.cache.SaveJWKS(ctx, buf)
		if err != nil {
			return nil, fmt.Errorf("failed to save jwks to cache layer: %w", err)
		}

		b = buf
	}

	return jwk.ParseString(string(b))

}
