package cache

import "github.com/go-redis/redis/v8"

type Service interface {
	authService
	transactionService
}

type service struct {
	client *redis.Client
}

func New(client *redis.Client) Service {
	return &service{
		client,
	}
}
