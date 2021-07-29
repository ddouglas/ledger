package item

import (
	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/account"
	"github.com/ddouglas/ledger/internal/gateway"
)

type configOption func(s *service)

type service struct {
	gateway gateway.Service
	account account.Service

	ledger.ItemRepository
	ledger.PlaidRepository
}

func WithItemRepository(item ledger.ItemRepository) configOption {
	return func(s *service) {
		s.ItemRepository = item
	}
}

func WithPlaidRepository(plaid ledger.PlaidRepository) configOption {
	return func(s *service) {
		s.PlaidRepository = plaid
	}
}

func WithGateway(gateway gateway.Service) configOption {
	return func(s *service) {
		s.gateway = gateway
	}
}

func WithAccount(account account.Service) configOption {
	return func(s *service) {
		s.account = account
	}
}
