package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type UserRepository interface {
	User(ctx context.Context, id uuid.UUID) (*User, error)
	UserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user *User) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Email        string    `db:"email" json:"email"`
	Auth0Subject string    `db:"auth0_subject" json:"auth0Subject"`
	CreatedAt    time.Time `db:"created_at" json:"-"`
	UpdatedAt    time.Time `db:"updated_at" json:"-"`
}

func (u *User) Validate() error {

	if u.Name == "" {
		return fmt.Errorf("user name must not be empty")
	}

	if u.Email == "" {
		return fmt.Errorf("user email must not be empty")
	}

	if u.Auth0Subject == "" {
		return fmt.Errorf("user auth0 subject must not be empty")
	}

	return nil

}
