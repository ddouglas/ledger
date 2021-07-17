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
	ledger.InstitutionRepository
}

func WithItemRepository(item ledger.ItemRepository) configOption {
	return func(s *service) {
		s.ItemRepository = item
	}
}

func WithInstitutionRepository(institution ledger.InstitutionRepository) configOption {
	return func(s *service) {
		s.InstitutionRepository = institution
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
