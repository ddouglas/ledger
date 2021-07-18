package mysql

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type institutionRepository struct {
	db *sqlx.DB
}

var institutionColumns = []string{
	"id", "name", "created_at", "updated_at",
}

func NewInstitutionRepository(db *sqlx.DB) ledger.InstitutionRepository {
	return &institutionRepository{db}
}

func (r *institutionRepository) Institution(ctx context.Context, id string) (*ledger.Institution, error) {

	query, args, err := sq.Select(institutionColumns...).From("institutions").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var institution = new(ledger.Institution)
	err = r.db.GetContext(ctx, institution, query, args...)

	return institution, errors.Wrap(err, "[Institution]")

}

func (r *institutionRepository) Institutions(ctx context.Context) ([]*ledger.Institution, error) {

	query, args, err := sq.Select(institutionColumns...).From("institutions").ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var institutions = make([]*ledger.Institution, 0)
	err = r.db.SelectContext(ctx, &institutions, query, args...)

	return institutions, errors.Wrap(err, "[Institutions]")

}

func (r *institutionRepository) CreateInstitution(ctx context.Context, institution *ledger.Institution) (*ledger.Institution, error) {

	query, args, err := sq.Insert("institutions").Columns(
		institutionColumns...,
	).Values(
		institution.ID,
		institution.Name,
		sq.Expr(`NOW()`),
		sq.Expr(`NOW()`),
	).Options("IGNORE").ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[CreateInstitution]")
	}

	return r.Institution(ctx, institution.ID)

}
