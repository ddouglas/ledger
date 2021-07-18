package account

import (
	"github.com/ddouglas/ledger"
)

type configOption func(s *service)

type service struct {
	// cache cache.Service

	// gateway gateway.Service
	ledger.AccountRepository
}

func WithAccountRepository(account ledger.AccountRepository) configOption {
	return func(s *service) {
		s.AccountRepository = account
	}
}
