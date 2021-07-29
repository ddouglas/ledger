package mysql

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type plaidRepository struct {
	db *sqlx.DB
}

var (
	plaidInstitutionColumns = []string{
		"id", "name", "created_at", "updated_at",
	}
	plaidCategoryColumns = []string{
		"id", "name", "`group`", "hierarchy", "created_at", "updated_at",
	}
	plaidInstitutionsTable = "plaid_institutions"
	plaidCategoriesTable   = "plaid_categories"
)

func NewPlaidRepository(db *sqlx.DB) ledger.PlaidRepository {
	return &plaidRepository{db}
}

func (r *plaidRepository) PlaidCategory(ctx context.Context, id string) (*ledger.PlaidCategory, error) {

	query, args, err := sq.Select(plaidCategoryColumns...).From(plaidCategoriesTable).Where(sq.Eq{
		"id": id,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var category = new(ledger.PlaidCategory)
	err = r.db.GetContext(ctx, category, query, args...)

	return category, errors.Wrap(err, "[PlaidCategory]")

}
func (r *plaidRepository) CreatePlaidCategory(ctx context.Context, category *ledger.PlaidCategory) (*ledger.PlaidCategory, error) {

	query, args, err := sq.Insert(plaidCategoriesTable).Columns(plaidCategoryColumns...).Values(
		category.ID, category.Name, category.Group, category.Hierarchy, sq.Expr(`NOW()`), sq.Expr(`NOW()`),
	).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[CreatePlaidCategory]")
	}

	return r.PlaidCategory(ctx, category.ID)

}

func (r *plaidRepository) PlaidInstitution(ctx context.Context, id string) (*ledger.PlaidInstitution, error) {

	query, args, err := sq.Select(plaidInstitutionColumns...).From(plaidInstitutionsTable).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var institution = new(ledger.PlaidInstitution)
	err = r.db.GetContext(ctx, institution, query, args...)

	return institution, errors.Wrap(err, "[Institution]")

}

func (r *plaidRepository) PlaidInstitutions(ctx context.Context) ([]*ledger.PlaidInstitution, error) {

	query, args, err := sq.Select(plaidInstitutionColumns...).From(plaidInstitutionsTable).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var institutions = make([]*ledger.PlaidInstitution, 0)
	err = r.db.SelectContext(ctx, &institutions, query, args...)

	return institutions, errors.Wrap(err, "[Institutions]")

}

func (r *plaidRepository) CreatePlaidInstitution(ctx context.Context, institution *ledger.PlaidInstitution) (*ledger.PlaidInstitution, error) {

	query, args, err := sq.Insert(plaidInstitutionsTable).Columns(
		plaidInstitutionColumns...,
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

	return r.PlaidInstitution(ctx, institution.ID)

}
