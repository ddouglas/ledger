package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type transactionService interface {
	FetchPresignedURL(ctx context.Context, transactionID string) (string, error)
	CachePresignedURL(ctx context.Context, transactionID, url string, duration time.Duration) error
}

func presignedURLKeyFunc(transactionID string) string {
	return fmt.Sprintf("ledger::preseignedReceiptURLs::%s", transactionID)
}

func (s *service) FetchPresignedURL(ctx context.Context, transactionID string) (string, error) {

	result, err := s.client.Get(ctx, presignedURLKeyFunc(transactionID)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", errors.Wrap(err, "[cache.FetchPresignedURL]")
	}

	return result, nil

}

func (s *service) CachePresignedURL(ctx context.Context, transactionID, url string, duration time.Duration) error {

	_, err := s.client.Set(ctx, presignedURLKeyFunc(transactionID), url, duration).Result()
	return errors.Wrap(err, "[cache.CachePresignedURL]")

}
