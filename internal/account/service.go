// Package account provides service access to account logic and repositories
package account

import (
	"github.com/ddouglas/ledger"
)

type Service interface {
	ledger.AccountRepository
}

type service struct {
	// cache cache.Service

	// gateway gateway.Service
	ledger.AccountRepository
}

func New(account ledger.AccountRepository) Service {
	return &service{
		AccountRepository: account,
	}
}
