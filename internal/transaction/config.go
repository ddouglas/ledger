package transaction

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/cache"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/sirupsen/logrus"
)

type configOption func(s *Service)

type Service struct {
	logger  *logrus.Logger
	cache   cache.Service
	s3      *s3.Client
	gateway gateway.Service
	bucket  string

	ledger.TransactionRepository
}

func WithGateway(gateway gateway.Service) configOption {
	return func(s *Service) {
		s.gateway = gateway
	}
}

func WithCache(cache cache.Service) configOption {
	return func(s *Service) {
		s.cache = cache
	}
}

func WithS3(s3 *s3.Client, bucket string) configOption {
	return func(s *Service) {
		s.s3 = s3
		s.bucket = bucket
	}
}

func WithTransactionRepository(transaction ledger.TransactionRepository) configOption {
	return func(s *Service) {
		s.TransactionRepository = transaction
	}
}

func WithLogger(logger *logrus.Logger) configOption {
	return func(s *Service) {
		s.logger = logger
	}
}
