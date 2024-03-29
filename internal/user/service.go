package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/jwt"
)

type Service interface {
	UserFromToken(ctx context.Context, token jwt.Token) (*ledger.User, error)
	FetchOrCreateUser(ctx context.Context, user *ledger.User) (*ledger.User, error)
	ledger.UserRepository
}

type service struct {
	registrationEnabled bool
	ledger.UserRepository
}

func New(registrationEnabled bool, user ledger.UserRepository) Service {
	return &service{
		registrationEnabled: registrationEnabled,
		UserRepository:      user,
	}

}

// func (s *service) ProcessLinkToken(ctx context.Context, token *ledger.LinkToken) error {

// 	return nil

// }

func (s *service) UserFromToken(ctx context.Context, token jwt.Token) (*ledger.User, error) {

	claims := token.PrivateClaims()

	var user *ledger.User
	var err error

	userIDClaim, userIDok := claims["https://userID"]
	if userIDok {
		id, ok := userIDClaim.(string)
		if !ok {
			return nil, fmt.Errorf("expected https://userID claim to be string, got %T", userIDClaim)
		}

		uuidID, err := uuid.FromString(id)
		if err != nil {
			return nil, fmt.Errorf("failed to parse valid uuid from token userID claim: %w", err)
		}

		user, err = s.User(ctx, uuidID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, "failed to fetch user by userID")
		}

	}

	emailClaim, emailClaimOk := claims["https://email"]
	if !userIDok && emailClaimOk {
		email, ok := emailClaim.(string)
		if !ok {
			return nil, fmt.Errorf("expected https://email claim to be stirng, got %T", emailClaim)
		}

		user, err = s.UserByEmail(ctx, email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, "failed to fetch user by email")
		}
	}

	if user == nil {
		if !s.registrationEnabled {
			return nil, fmt.Errorf("failed to parse userID nor email claim to valid users")
		}
		if emailClaimOk {

			user, err = s.FetchOrCreateUser(ctx, &ledger.User{
				Email:        emailClaim.(string),
				Auth0Subject: token.Subject(),
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to create user ")
			}
		}
	}

	return user, nil

}

func (s *service) FetchOrCreateUser(ctx context.Context, newUser *ledger.User) (*ledger.User, error) {

	if newUser.Email == "" {
		return nil, fmt.Errorf("required identifying attribute missing from input")
	}

	user, err := s.UserRepository.UserByEmail(ctx, newUser.Email)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	err = newUser.Validate()
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate account id: %w", err)
	}

	newUser.ID = id

	return s.UserRepository.CreateUser(ctx, newUser)

}
