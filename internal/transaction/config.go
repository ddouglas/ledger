package transaction

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/cache"
	"github.com/sirupsen/logrus"
)

type configOption func(s *service)

type service struct {
	logger *logrus.Logger
	cache  cache.Service
	s3     *s3.Client
	bucket string

	ledger.TransactionRepository
}

func WithCache(cache cache.Service) configOption {
	return func(s *service) {
		s.cache = cache
	}
}

func WithS3(s3 *s3.Client, bucket string) configOption {
	return func(s *service) {
		s.s3 = s3
		s.bucket = bucket
	}
}

func WithTransactionRepository(transaction ledger.TransactionRepository) configOption {
	return func(s *service) {
		s.TransactionRepository = transaction
	}
}

func WithLogger(logger *logrus.Logger) configOption {
	return func(s *service) {
		s.logger = logger
	}
}
