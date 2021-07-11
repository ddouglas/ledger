package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type authService interface {
	JWKS(ctx context.Context) ([]byte, error)
	SaveJWKS(ctx context.Context, jwks []byte) error
}

const keyAuthJWKS = "ledger::auth::jwks"

func (s *service) JWKS(ctx context.Context) ([]byte, error) {

	result, err := s.client.Get(ctx, keyAuthJWKS).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil

}

func (s *service) SaveJWKS(ctx context.Context, jwks []byte) error {

	_, err := s.client.Set(ctx, keyAuthJWKS, jwks, time.Hour*6).Result()

	return err

}
