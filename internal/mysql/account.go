package mysql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
)

type accountRepository struct {
	db *sqlx.DB
}

var accountColumns := []string{
	"account_id",
	"mask",
	"name",
	"official_name",
	"subtype",
	"type",
	"available",
	"current",
	"iso_currency_code",
	"limit",
	"unofficial_currency_code",
	"last_updated_datetime",
}

func NewAccountRepository(db *sqlx.DB) ledger.AccountRepository {
	return &accountRepository{db: db}
}



func (r *accountRepository) Account(ctx context.Context, itemID string, accountID string) (*ledger.Account, error) {
	
	query, args, err := sq.Select()

}

func (r *accountRepository) Accounts(ctx context.Context, itemID string) ([]*ledger.Account, error) {
	return nil, nil
}

func (r *accountRepository) CreateAccount(ctx context.Context, account *ledger.Account) (*ledger.Account, error) {
	return nil, nil
}

func (r *accountRepository) UpdateAccount(ctx context.Context, itemID, accountID string, account *ledger.Account) (*ledger.Account, error) {
	return nil, nil
}

func (r *accountRepository) DeleteAccount(ctx context.Context, itemID, accountID string) error {
	return nil
}
