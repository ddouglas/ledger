package mysql

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

var userColumns = []string{
	"id", "name", "email", "auth0_subject", "created_at", "updated_at",
}

func NewUserRepository(db *sqlx.DB) ledger.UserRepository {
	return &userRepository{
		db,
	}
}

func (r *userRepository) User(ctx context.Context, id uuid.UUID) (*ledger.User, error) {

	query := sq.Select(userColumns...).
		From("users").
		Where(sq.Eq{"id": id}).
		Limit(1)

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	var user = new(ledger.User)
	err = r.db.GetContext(ctx, user, stmt, args...)

	return user, err

}

func (r *userRepository) UserByEmail(ctx context.Context, email string) (*ledger.User, error) {

	query := sq.Select(userColumns...).
		From("users").
		Where(sq.Eq{"email": email}).
		Limit(1)

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	var user = new(ledger.User)
	err = r.db.GetContext(ctx, user, stmt, args...)

	return user, err

}

func (r *userRepository) CreateUser(ctx context.Context, user *ledger.User) (*ledger.User, error) {

	query := sq.Insert("users").Columns(
		userColumns...,
	).Values(
		user.ID, user.Name, user.Email,
		user.Auth0Subject, time.Now(), time.Now(),
	)

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	_, err = r.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return r.User(ctx, user.ID)

}

func (r *userRepository) UpdateUser(ctx context.Context, id uuid.UUID, user *ledger.User) (*ledger.User, error) {

	query := sq.Update("users").
		Set("name", user.Name).
		Set("email", user.Email).
		Set("auth0_subject", user.Auth0Subject).
		Where(sq.Eq{"id": id})

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	_, err = r.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return r.User(ctx, user.ID)

}

func (r *userRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {

	query := sq.Delete("users").Where(sq.Eq{"id": id})

	stmt, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	_, err = r.db.ExecContext(ctx, stmt, args...)

	return err

}
