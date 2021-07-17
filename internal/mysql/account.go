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

const accountTable = "accounts"

var accountColumns = []string{
	"item_id",
	"account_id",
	"mask",
	"name",
	"official_name",
	"balance_available",
	"balance_current",
	"iso_currency_code",
	"limit",
	"unofficial_currency_code",
	"last_updated_datetime",
	"subtype",
	"type",
}

func NewAccountRepository(db *sqlx.DB) ledger.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Account(ctx context.Context, itemID string, accountID string) (*ledger.Account, error) {

	query, args, err := sq.Select(accountColumns...).From(accountTable).Where(sq.Eq{"item_id": itemID, "account_id": accountID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query accounts: %w", err)
	}

	for rows.Next() {
		var (
			item_id string
			account_id string
			mask string
			name string
			official_name string
			balance_available float64
			balance_current float64
			iso_currency_code
			limit
			unofficial_currency_code
			last_updated_datetime
			subtype string
			accountType string
			
		)
		

	}


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
