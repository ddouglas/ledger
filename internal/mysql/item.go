package mysql

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type userItemRepository struct {
	db *sqlx.DB
}

const userItemTable = "user_items"

var userItemColumns = []string{
	"user_id",
	"item_id",
	"access_token",
	"institution_id",
	"webhook",
	"error",
	"available_products",
	"billed_products",
	"consent_expiration_time",
	"update_type",
	"item_status",
	"created_at",
	"updated_at",
}

func NewItemRepository(db *sqlx.DB) ledger.ItemRepository {
	return &userItemRepository{db: db}
}

func (r *userItemRepository) Item(ctx context.Context, itemID string) (*ledger.Item, error) {

	query, args, err := sq.Select(userColumns...).From(userItemTable).Where(sq.Eq{
		"item_id": itemID,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var item = new(ledger.Item)
	err = r.db.GetContext(ctx, item, query, args...)

	return item, err

}

func (r *userItemRepository) ItemByUserID(ctx context.Context, userID uuid.UUID, itemID string) (*ledger.Item, error) {

	query, args, err := sq.Select(userColumns...).From(userItemTable).Where(sq.Eq{
		"item_id": itemID,
		"user_id": userID,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var item = new(ledger.Item)
	err = r.db.GetContext(ctx, item, query, args...)

	return item, err

}

func (r *userItemRepository) ItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*ledger.Item, error) {

	query, args, err := sq.Select(userColumns...).From(userItemTable).Where(sq.Eq{
		"user_id": userID,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var items = make([]*ledger.Item, 0)
	err = r.db.SelectContext(ctx, &items, query, args...)

	return items, err

}

func (r *userItemRepository) CreateItem(ctx context.Context, item *ledger.Item) (*ledger.Item, error) {

	query, args, err := sq.Insert(userItemTable).Columns(userItemColumns...).Values(
		item.UserID,
		item.ItemID,
		item.AccessToken,
		item.InstitutionID,
		item.Webhook,
		item.Error,
		item.AvailableProducts,
		item.BilledProducts,
		item.ConsentExpirationTime,
		item.UpdateType,
		item.ItemStatus,
		sq.Expr(`NOW()`),
		sq.Expr(`NOW()`),
	).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert item: %w", err)
	}

	return r.ItemByUserID(ctx, item.UserID, item.ItemID)

}

func (r *userItemRepository) UpdateItem(ctx context.Context, itemID string, item *ledger.Item) (*ledger.Item, error) {
	return nil, nil
}

func (r *userItemRepository) DeleteItem(ctx context.Context, userID uuid.UUID, itemID string) error {
	return nil
}
