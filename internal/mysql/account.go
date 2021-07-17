package mysql

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/null"
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
	"balance_limit",
	"balance_last_updated",
	"iso_currency_code",
	"unofficial_currency_code",
	"subtype",
	"type",
	"created_at",
	"updated_at",
}

func NewAccountRepository(db *sqlx.DB) ledger.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Account(ctx context.Context, itemID string, accountID string) (*ledger.Account, error) {

	query, args, err := sq.Select(accountColumns...).From(accountTable).Where(sq.Eq{"item_id": itemID, "account_id": accountID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var account = new(ledger.Account)
	err = r.db.GetContext(ctx, account, query, args...)

	return account, err

}

func (r *accountRepository) Accounts(ctx context.Context, itemID string) ([]*ledger.Account, error) {

	query, args, err := sq.Select(accountColumns...).From(accountTable).Where(sq.Eq{"item_id": itemID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}

	defer rows.Close()
	var accounts = make([]*ledger.Account, 0)
	for rows.Next() {

		var (
			item_id                  string
			account_id               string
			mask                     null.String
			name                     null.String
			official_name            null.String
			balance_available        float64
			balance_current          float64
			balance_limit            float64
			balance_last_updated     null.Time
			iso_currency_code        string
			unofficial_currency_code null.String
			subtype                  null.String
			accountType              null.String
			created_at               time.Time
			updated_at               time.Time
		)

		err = rows.Scan(
			item_id, account_id, mask, name,
			official_name, balance_available, balance_current, balance_limit,
			balance_last_updated, iso_currency_code, unofficial_currency_code, subtype,
			accountType, created_at, updated_at,
		)
		if err != nil {
			return nil, fmt.Errorf("faild to scan row: %w", err)
		}

		accounts = append(accounts, &ledger.Account{
			ItemID:       item_id,
			AccountID:    account_id,
			Mask:         mask,
			Name:         name,
			OfficialName: official_name,
			Subtype:      subtype,
			Type:         accountType,
			CreatedAt:    created_at,
			UpdatedAt:    updated_at,
			Balance: &ledger.AccountBalance{
				Available:              balance_available,
				Current:                balance_current,
				Limit:                  balance_limit,
				ISOCurrencyCode:        iso_currency_code,
				UnofficialCurrencyCode: unofficial_currency_code,
				LastUpdated:            balance_last_updated,
			},
		})

	}

	return accounts, nil

}

func (r *accountRepository) AccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*ledger.Account, error) {

	query, args, err := sq.Select(accountColumns...).From(accountTable).Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}

	defer rows.Close()
	var accounts = make([]*ledger.Account, 0)
	for rows.Next() {

		var (
			item_id                  string
			account_id               string
			mask                     null.String
			name                     null.String
			official_name            null.String
			balance_available        float64
			balance_current          float64
			balance_limit            float64
			balance_last_updated     null.Time
			iso_currency_code        string
			unofficial_currency_code null.String
			subtype                  null.String
			accountType              null.String
			created_at               time.Time
			updated_at               time.Time
		)

		err = rows.Scan(
			item_id, account_id, mask, name,
			official_name, balance_available, balance_current, balance_limit,
			balance_last_updated, iso_currency_code, unofficial_currency_code, subtype,
			accountType, created_at, updated_at,
		)
		if err != nil {
			return nil, fmt.Errorf("faild to scan row: %w", err)
		}

		accounts = append(accounts, &ledger.Account{
			ItemID:       item_id,
			AccountID:    account_id,
			Mask:         mask,
			Name:         name,
			OfficialName: official_name,
			Subtype:      subtype,
			Type:         accountType,
			CreatedAt:    created_at,
			UpdatedAt:    updated_at,
			Balance: &ledger.AccountBalance{
				Available:              balance_available,
				Current:                balance_current,
				Limit:                  balance_limit,
				ISOCurrencyCode:        iso_currency_code,
				UnofficialCurrencyCode: unofficial_currency_code,
				LastUpdated:            balance_last_updated,
			},
		})

	}

	return accounts, nil

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
