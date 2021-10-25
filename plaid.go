package ledger

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type PlaidRepository interface {
	PlaidInstitution(ctx context.Context, id string) (*PlaidInstitution, error)
	PlaidInstitutions(ctx context.Context) ([]*PlaidInstitution, error)
	CreatePlaidInstitution(ctx context.Context, institution *PlaidInstitution) (*PlaidInstitution, error)

	PlaidCategory(ctx context.Context, id string) (*PlaidCategory, error)
	PlaidCategories(ctx context.Context) ([]*PlaidCategory, error)
	CreatePlaidCategory(ctx context.Context, category *PlaidCategory) (*PlaidCategory, error)
}

type PlaidInstitution struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
}

type PlaidCategory struct {
	ID        string      `db:"id" json:"id"`
	Name      string      `db:"name" json:"name"`
	Group     string      `db:"group" json:"group"`
	Hierarchy SliceString `db:"hierarchy" json:"hierarchy"`
	CreatedAt time.Time   `db:"created_at" json:"-"`
	UpdatedAt time.Time   `db:"updated_at" json:"-"`
}

type LinkState struct {
	UserID     uuid.UUID
	State      uuid.UUID
	Token      string
	Expiration time.Time
}

type LinkToken struct {
	State       uuid.UUID `json:"state"`
	AccessToken string    `json:"access_token"`
}
