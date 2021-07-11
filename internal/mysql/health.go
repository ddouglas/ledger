package mysql

import (
	"context"

	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
)

type healthRepository struct {
	db *sqlx.DB
}

func NewHealthRepository(db *sqlx.DB) ledger.HealthRepository {
	return &healthRepository{db}
}

func (r *healthRepository) Cheak(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
