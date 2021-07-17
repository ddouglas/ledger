package mysql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "[Item]")
	}

	var item = new(ledger.Item)
	err = r.db.GetContext(ctx, item, query, args...)

	return item, errors.Wrap(err, "[Item]")

}

func (r *userItemRepository) ItemByUserID(ctx context.Context, userID uuid.UUID, itemID string) (*ledger.Item, error) {

	query, args, err := sq.Select(userColumns...).From(userItemTable).Where(sq.Eq{
		"item_id": itemID,
		"user_id": userID,
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[ItemByUserID]")
	}

	var item = new(ledger.Item)
	err = r.db.GetContext(ctx, item, query, args...)

	return item, errors.Wrap(err, "[ItemByUserID]")

}

func (r *userItemRepository) ItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*ledger.Item, error) {

	query, args, err := sq.Select(userItemColumns...).From(userItemTable).Where(sq.Eq{
		"user_id": userID,
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[ItemsByUserID]")
	}

	var items = make([]*ledger.Item, 0)
	err = r.db.SelectContext(ctx, &items, query, args...)

	return items, errors.Wrap(err, "[ItemsByUserID]")

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
		return nil, errors.Wrap(err, "[CreateItem]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[CreateItem]")
	}

	return r.ItemByUserID(ctx, item.UserID, item.ItemID)

}

func (r *userItemRepository) UpdateItem(ctx context.Context, itemID string, item *ledger.Item) (*ledger.Item, error) {

	query, args, err := sq.Update(userItemTable).
		Set("user_id", item.UserID).
		Set("item_id", item.ItemID).
		Set("access_token", item.AccessToken).
		Set("institution_id", item.InstitutionID).
		Set("webhook", item.Webhook).
		Set("error", item.Error).
		Set("available_products", item.AvailableProducts).
		Set("billed_products", item.BilledProducts).
		Set("consent_expiration_time", item.ConsentExpirationTime).
		Set("update_type", item.UpdateType).
		Set("item_status", item.ItemStatus).
		Set("updated_at", sq.Expr(`NOW()`)).
		Where(sq.Eq{"item_id": item.ItemID, "user_id": item.UserID}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[UpdateItem]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[UpdateItem]")
	}

	return r.ItemByUserID(ctx, item.UserID, item.ItemID)

}

func (r *userItemRepository) DeleteItem(ctx context.Context, userID uuid.UUID, itemID string) error {

	query, args, err := sq.Delete(userItemTable).Where(sq.Eq{"item_id": itemID, "user_id": userID}).ToSql()

	if err != nil {
		return errors.Wrap(err, "[DeleteItem]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "[DeleteItem]")
	}

	return errors.Wrap(err, "[DeleteItem]")

}
