package mysql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var merchantsTable = "merchants"
var merchantsColumns = []string{
	"id",
	"name",
	"created_at",
	"updated_at",
}

var merchantAliasesTable = "merchant_aliases"
var merchantAliasesColumns = []string{
	"alias_id",
	"merchant_id",
	"alias",
	"created_at",
	"updated_at",
}

type merchantRepository struct {
	db *sqlx.DB
}

func NewMerchantRepository(db *sqlx.DB) ledger.MerchantRepository {
	return &merchantRepository{
		db: db,
	}
}

func (r *merchantRepository) Merchant(ctx context.Context, id string) (*ledger.Merchant, error) {

	query, args, err := sq.Select(merchantsColumns...).From(merchantsTable).Where(sq.Eq{
		"id": id,
	}).ToSql()
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	var merchant = new(ledger.Merchant)
	err = r.db.GetContext(ctx, merchant, query, args...)
	return merchant, err

}

func (r *merchantRepository) MerchantByAlias(ctx context.Context, alias string) (*ledger.Merchant, error) {

	query := `
		SELECT
			m.id,
			m.name,
			m.created_at,
			m.updated_at
		FROM merchant_aliases ma
		LEFT JOIN merchants m ON (m.id = ma.merchant_id)
		WHERE ma.Alias = ?
	`

	var merchant = new(ledger.Merchant)
	err := r.db.GetContext(ctx, merchant, query, alias)
	return merchant, err

}

func (r *merchantRepository) Merchants(ctx context.Context) ([]*ledger.Merchant, error) {

	query, args, err := sq.Select(merchantsColumns...).From(merchantsTable).OrderBy("name asc").ToSql()
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	var merchants = make([]*ledger.Merchant, 0)
	err = r.db.SelectContext(ctx, &merchants, query, args...)
	return merchants, err

}

func (r *merchantRepository) createMerchantQuery(merchant *ledger.Merchant) (string, []interface{}, error) {
	return sq.Insert(merchantsTable).Columns(merchantsColumns...).Values(
		merchant.ID,
		merchant.Name,
		sq.Expr(`NOW()`),
		sq.Expr(`NOW()`),
	).ToSql()
}

func (r *merchantRepository) CreateMerchant(ctx context.Context, merchant *ledger.Merchant) (*ledger.Merchant, error) {

	query, args, err := r.createMerchantQuery(merchant)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return merchant, err

}

func (r *merchantRepository) CreateMerchantTx(ctx context.Context, tx ledger.Transactioner, merchant *ledger.Merchant) (*ledger.Merchant, error) {

	txn, ok := tx.(*transaction)
	if !ok {
		return nil, ErrInvalidTransaction
	}

	query, args, err := r.createMerchantQuery(merchant)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = txn.ExecContext(ctx, query, args...)
	return merchant, err

}

func (r *merchantRepository) updateMerchantQuery(id string, merchant *ledger.Merchant) (string, []interface{}, error) {
	return sq.Update(merchantsTable).SetMap(map[string]interface{}{
		"name":       merchant.Name,
		"updated_at": sq.Expr(`NOW()`),
	}).Where(sq.Eq{"id": id}).ToSql()
}

func (r *merchantRepository) UpdateMerchant(ctx context.Context, id string, merchant *ledger.Merchant) (*ledger.Merchant, error) {

	query, args, err := r.updateMerchantQuery(id, merchant)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return merchant, err

}

func (r *merchantRepository) UpdateMerchantTx(ctx context.Context, tx ledger.Transactioner, id string, merchant *ledger.Merchant) (*ledger.Merchant, error) {

	txn, ok := tx.(*transaction)
	if !ok {
		return nil, ErrInvalidTransaction
	}

	query, args, err := r.updateMerchantQuery(id, merchant)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = txn.ExecContext(ctx, query, args...)
	return merchant, err

}

func (r *merchantRepository) deleteMerchantQuery(id string) (string, []interface{}, error) {

	return sq.Delete(merchantsTable).Where(sq.Eq{"id": id}).ToSql()

}

func (r *merchantRepository) DeleteMerchant(ctx context.Context, id string) error {

	query, args, err := r.deleteMerchantQuery(id)
	if err != nil {
		return errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return err

}

func (r *merchantRepository) DeleteMerchantTx(ctx context.Context, tx ledger.Transactioner, id string) error {

	txn, ok := tx.(*transaction)
	if !ok {
		return ErrInvalidTransaction
	}

	query, args, err := r.deleteMerchantQuery(id)
	if err != nil {
		return errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = txn.ExecContext(ctx, query, args...)
	return err

}

func (r *merchantRepository) MerchantAliasesByMerchantID(ctx context.Context, merchantID string) ([]*ledger.MerchantAlias, error) {

	query, args, err := sq.Select(merchantAliasesColumns...).From(merchantAliasesTable).Where(sq.Eq{
		"merchant_id": merchantID,
	}).ToSql()
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	var aliases = make([]*ledger.MerchantAlias, 0)
	err = r.db.SelectContext(ctx, &aliases, query, args...)
	return aliases, err

}

func (r *merchantRepository) createMerchantAliasQuery(alias *ledger.MerchantAlias) (string, []interface{}, error) {
	return sq.Insert(merchantAliasesTable).SetMap(map[string]interface{}{
		"alias_id":    alias.AliasID,
		"merchant_id": alias.MerchantID,
		"alias":       alias.Alias,
		"created_at":  sq.Expr(`NOW()`),
		"updated_at":  sq.Expr(`NOW()`),
	}).ToSql()
}

func (r *merchantRepository) CreateMerchantAlias(ctx context.Context, alias *ledger.MerchantAlias) (*ledger.MerchantAlias, error) {

	query, args, err := r.createMerchantAliasQuery(alias)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return alias, err

}

func (r *merchantRepository) CreateMerchantAliasTx(ctx context.Context, tx ledger.Transactioner, alias *ledger.MerchantAlias) (*ledger.MerchantAlias, error) {

	txn, ok := tx.(*transaction)
	if !ok {
		return nil, ErrInvalidTransaction
	}

	query, args, err := r.createMerchantAliasQuery(alias)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = txn.ExecContext(ctx, query, args...)
	return alias, err

}

func (r *merchantRepository) updateMerchantAliasQuery(aliasID string, alias *ledger.MerchantAlias) (string, []interface{}, error) {
	return sq.Update(merchantsTable).SetMap(map[string]interface{}{
		"merchant_id": alias.MerchantID,
		"alias":       alias.Alias,
		"updated_at":  sq.Expr(`NOW()`),
	}).Where(sq.Eq{"alias_id": aliasID}).ToSql()
}

func (r *merchantRepository) UpdateMerchantAlias(ctx context.Context, aliasID string, alias *ledger.MerchantAlias) (*ledger.MerchantAlias, error) {

	query, args, err := r.updateMerchantAliasQuery(aliasID, alias)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	return alias, err

}

func (r *merchantRepository) UpdateMerchantAliasTx(ctx context.Context, tx ledger.Transactioner, aliasID string, alias *ledger.MerchantAlias) (*ledger.MerchantAlias, error) {

	txn, ok := tx.(*transaction)
	if !ok {
		return nil, ErrInvalidTransaction
	}

	query, args, err := r.updateMerchantAliasQuery(aliasID, alias)
	if err != nil {
		return nil, errors.Errorf("failed to generate sql stmt: %s", err)
	}

	_, err = txn.ExecContext(ctx, query, args...)
	return alias, err

}
