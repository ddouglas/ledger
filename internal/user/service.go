package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	ledger.UserRepository
}

func New(user ledger.UserRepository) Service {
	return &service{
		UserRepository: user,
	}
}

func (s *service) UserFromToken(ctx context.Context, token jwt.Token) (*ledger.User, error) {

	claims := token.PrivateClaims()
	if id, ok := claims["https://accountID"]; !ok || id.(string) == "" {
		return nil, fmt.Errorf("required key accountID is missing  from token")
	}

	accountID, err := uuid.FromString(claims["https://accountID"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse uuid: %w", err)
	}

	user, err := s.User(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user for provided accountID %s: %w", accountID, err)
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
