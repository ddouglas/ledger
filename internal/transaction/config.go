package transaction

import "github.com/ddouglas/ledger"

type configOption func(s *service)

type service struct {
	ledger.TransactionRepository
}

func WithTransactionRepository(transaction ledger.TransactionRepository) configOption {
	return func(s *service) {
		s.TransactionRepository = transaction
	}
}
