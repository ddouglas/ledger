package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	// "github.com/lestrrat-go/jwx/jwk"

	"github.com/ddouglas/ledger/internal/cache"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/oauth2"
)

type Service interface {
	// InitializeState(ctx context.Context) (string, error)
	// CheckState(ctx context.Context, key string) error

	ExchangeCode(ctx context.Context, code string) (string, jwt.Token, error)
	ValidateToken(ctx context.Context, t string) (jwt.Token, error)
}

type service struct {
	client *http.Client
	cache  cache.Service
	oauth  *oauth2.Config
	config *config
}

func New(cache cache.Service, client *http.Client, oauth *oauth2.Config, optFuncs ...configOption) Service {
	c := &config{}
	for _, optFunc := range optFuncs {
		optFunc(c)
	}

	return &service{
		client: client,
		cache:  cache,
		oauth:  oauth,
		config: c,
	}
}

func (s *service) ExchangeCode(ctx context.Context, code string) (string, jwt.Token, error) {

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

func (s *service) ValidateToken(ctx context.Context, t string) (jwt.Token, error) {

	set, err := s.getKeySet(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWK Set: %w", err)
	}

	options := make([]jwt.ParseOption, 0, 3)
	options = append(options, jwt.WithKeySet(set))
	if s.config.Audience != nil {
		options = append(options, jwt.WithAudience(*s.config.Audience))
	}
	if s.config.Issuer != nil {
		options = append(options, jwt.WithIssuer(*s.config.Issuer))
	}

	token, err := jwt.ParseString(t, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return token, nil

}

// func (s *service) InitializeState(ctx context.Context) (string, error) {

// 	state := fmt.Sprintf("%x", sha256.Sum256([]byte(time.Now().Format(time.RFC3339Nano))))
// 	_, err := s.redis.Set(ctx, fmt.Sprintf("state:%s", state), 1, time.Minute*5).Result()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to store state key in redis: %w", err)
// 	}

// 	return state, nil

// }

// func (s *service) CheckState(ctx context.Context, hash string) error {

// 	key := stateKey(hash)
// 	_, err := s.redis.Get(ctx, key).Result()
// 	if err != nil && !errors.Is(err, redis.Nil) {
// 		return fmt.Errorf("failed to check redis for state: %w", err)
// 	}
// 	if err != nil && errors.Is(err, redis.Nil) {
// 		return fmt.Errorf("state key is expired or was never set")
// 	}

// 	_, err = s.redis.Del(ctx, key).Result()
// 	if err != nil {
// 		return fmt.Errorf("failed to remove state key from redis: %w", err)
// 	}

// 	return nil

// }

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
