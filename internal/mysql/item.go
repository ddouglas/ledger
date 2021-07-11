package mysql

import (
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
)

type userItemRepository struct {
	db *sqlx.DB
}

func NewItemRepository(db *sqlx.DB) ledger.ItemRepository {
	return &userItemRepository{db: db}
}
