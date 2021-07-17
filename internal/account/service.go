// Package account provides service access to account logic and repositories
package account

import (
	"github.com/ddouglas/ledger"
)

type Service interface {
	ledger.AccountRepository
}

func New(optFuncs ...configOption) Service {
	s := &service{}
	for _, optFunc := range optFuncs {
		optFunc(s)
	}
	return s
}
