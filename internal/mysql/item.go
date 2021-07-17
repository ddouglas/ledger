package mysql

import (
	"context"

	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type userItemRepository struct {
	db *sqlx.DB
}

func NewItemRepository(db *sqlx.DB) ledger.ItemRepository {
	return &userItemRepository{db: db}
}

func (r *userItemRepository) Item(ctx context.Context, itemID string) (*ledger.Item, error) {
	return nil, nil
}

func (r *userItemRepository) ItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*ledger.Item, error) {
	return nil, nil
}

func (r *userItemRepository) CreateItem(ctx context.Context, item *ledger.Item) (*ledger.Item, error) {
	return nil, nil
}

func (r *userItemRepository) UpdateItem(ctx context.Context, itemID string, item *ledger.Item) (*ledger.Item, error) {
	return nil, nil
}

func (r *userItemRepository) DeleteItem(ctx context.Context, userID uuid.UUID, itemID string) error {
	return nil
}
