package user

import "github.com/ddouglas/ledger"

type configOption func(s *service)

type service struct {
	ledger.UserRepository
}

func WithUserRepository(user ledger.UserRepository) configOption {
	return func(s *service) {
		s.UserRepository = user
	}
}
