package mysql

import (
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
)

type accountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) ledger.AccountRepository {
	return &accountRepository{db: db}
}
