package mysql

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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
		return nil, errors.Wrapf(err, "[Account] ItemID: %s AccountID: %s", itemID, accountID)
	}

	var account = new(ledger.Account)
	err = r.db.GetContext(ctx, account, query, args...)

	return account, errors.Wrapf(err, "[Account] ItemID: %s AccountID: %s", itemID, accountID)

}

func (r *accountRepository) Accounts(ctx context.Context, itemID string) ([]*ledger.Account, error) {

	query, args, err := sq.Select(accountColumns...).From(accountTable).Where(sq.Eq{"item_id": itemID}).ToSql()
	if err != nil {
		return nil, errors.Wrapf(err, "[Accounts] ItemID: %s", itemID)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "[Accounts] ItemID: %s", itemID)
	}

	defer rows.Close()
	accounts, err := scanAccountFromRows(rows)
	if err != nil {
		return nil, errors.Wrapf(err, "[Accounts] ItemID: %s", itemID)
	}

	return accounts, nil

}

func (r *accountRepository) AccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*ledger.Account, error) {

	var columns = make([]string, 0, len(accountColumns))
	for _, column := range accountColumns {
		columns = append(columns, fmt.Sprintf("a.%s", column))
	}

	query, args, err := sq.Select(columns...).
		From(fmt.Sprintf("%s a", accountTable)).
		Where(sq.Eq{"ui.user_id": userID}).
		Join("user_items ui on ui.item_id = a.item_id").
		ToSql()
	if err != nil {
		return nil, errors.Wrapf(err, "[AccountsByUserID] UserID: %s", userID)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "[AccountsByUserID] UserID: %s", userID)
	}

	defer rows.Close()
	accounts, err := scanAccountFromRows(rows)

	return accounts, errors.Wrapf(err, "[AccountsByUserID] UserID: %s", userID)

}

func (r *accountRepository) AccountsByItemID(ctx context.Context, itemID string) ([]*ledger.Account, error) {

	query, args, err := sq.Select(accountColumns...).
		From(fmt.Sprintf("%s a", accountTable)).
		Where(sq.Eq{"item_id": itemID}).
		ToSql()
	if err != nil {
		return nil, errors.Wrapf(err, "[AccountsByItemID] UserID: %s", itemID)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "[AccountsByItemID] UserID: %s", itemID)
	}

	defer rows.Close()
	accounts, err := scanAccountFromRows(rows)

	return accounts, errors.Wrapf(err, "[AccountsByItemID] UserID: %s", itemID)

}

func scanAccountFromRows(rows *sqlx.Rows) ([]*ledger.Account, error) {

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

		err := rows.Scan(
			&item_id, &account_id, &mask, &name,
			&official_name, &balance_available, &balance_current, &balance_limit,
			&balance_last_updated, &iso_currency_code, &unofficial_currency_code, &subtype,
			&accountType, &created_at, &updated_at,
		)
		if err != nil {
			return nil, errors.Wrap(err, "[scanAccountFromRows]")
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

	mapColValues := mapAccount(account)
	mapColValues["created_at"] = sq.Expr(`NOW()`)
	mapColValues["updated_at"] = sq.Expr(`NOW()`)

	query, args, err := sq.Insert(accountTable).SetMap(mapColValues).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[CreateAccount]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[CreateAccount]")
	}

	return r.Account(ctx, account.ItemID, account.AccountID)

}

func (r *accountRepository) UpdateAccount(ctx context.Context, itemID, accountID string, account *ledger.Account) (*ledger.Account, error) {

	mapColValues := mapAccount(account)
	mapColValues["updated_at"] = sq.Expr(`NOW()`)

	query, args, err := sq.Update(accountTable).SetMap(mapColValues).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[UpdateAccount]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[UpdateAccount]")
	}

	return r.Account(ctx, account.ItemID, account.AccountID)

}

func (r *accountRepository) DeleteAccount(ctx context.Context, itemID, accountID string) error {

	query, args, err := sq.Delete(accountTable).Where(sq.Eq{"item_id": itemID, "account_id": accountID}).ToSql()
	if err != nil {
		return errors.Wrap(err, "[DeleteAccount]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "[DeleteAccount]")
	}

	return nil
}

func mapAccount(account *ledger.Account) map[string]interface{} {

	mapColValues := map[string]interface{}{
		"item_id":       account.ItemID,
		"account_id":    account.AccountID,
		"mask":          account.Mask,
		"name":          account.Name,
		"official_name": account.OfficialName,
		"subtype":       account.Subtype,
		"type":          account.Type,
	}

	if account.Balance != nil {
		mapColValues["balance_available"] = account.Balance.Available
		mapColValues["balance_current"] = account.Balance.Current
		mapColValues["balance_limit"] = account.Balance.Limit
		mapColValues["balance_last_updated"] = account.Balance.LastUpdated
		mapColValues["unofficial_currency_code"] = account.Balance.UnofficialCurrencyCode
		mapColValues["iso_currency_code"] = account.Balance.ISOCurrencyCode
	}

	return mapColValues

}
