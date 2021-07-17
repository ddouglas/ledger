package ledger

import (
	"context"
	"time"
)

type InstitutionRepository interface {
	Institution(ctx context.Context, id string) (*Institution, error)
	Institutions(ctx context.Context) ([]*Institution, error)
	CreateInstitution(ctx context.Context, institution *Institution) (*Institution, error)
}

type Institution struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
}
