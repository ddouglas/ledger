package transaction

import (
	"github.com/ddouglas/ledger"
	"github.com/sirupsen/logrus"
)

type configOption func(s *service)

type service struct {
	logger *logrus.Logger

	ledger.TransactionRepository
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
