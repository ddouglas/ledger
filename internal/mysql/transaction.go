package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/davecgh/go-spew/spew"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type transactionRepository struct {
	db *sqlx.DB
}

var transactionColumns = []string{
	"item_id",
	"account_id",
	"transaction_id",
	"pending_transaction_id",
	"category_id",
	"name",
	"pending",
	"payment_channel",
	"merchant_name",
	"categories",
	"unofficial_currency_code",
	"iso_currency_code",
	"amount",
	"transaction_code",
	"authorized_date",
	"authorized_datetime",
	"date",
	"datetime",
	"created_at",
	"updated_at",
}

const transactionsTableName = "transactions"

func NewTransactionRepository(db *sqlx.DB) ledger.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Transaction(ctx context.Context, itemID, transactionID string) (*ledger.Transaction, error) {

	query, args, err := sq.Select(transactionColumns...).From(transactionsTableName).Where(sq.Eq{
		"item_id":        itemID,
		"transaction_id": transactionID,
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	var transaction = new(ledger.Transaction)
	err = r.db.GetContext(ctx, transaction, query, args...)

	return transaction, errors.Wrap(err, "[mysql.Transaction]")

}

func (r *transactionRepository) TransactionsByDate(ctx context.Context, itemID string, date time.Time) ([]*ledger.Transaction, error) {

	query, args, err := sq.Select(transactionColumns...).From(transactionsTableName).Where(sq.Eq{
		"item_id": itemID,
		"date":    date.Format("2006-01-02"),
	}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sql stmt: %w", err)
	}

	var transactions = make([]*ledger.Transaction, 0)
	err = r.db.SelectContext(ctx, &transactions, query, args...)

	return transactions, errors.Wrap(err, "[mysql.TransactionsByDate]")

}

func (r *transactionRepository) TransactionsCount(ctx context.Context, itemID, accountID string) (uint64, error) {

	var count uint64
	query, args, err := sq.Select(`COUNT(*)`).From(transactionsTableName).Where(sq.Eq{
		"item_id":    itemID,
		"account_id": accountID,
	}).ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "[mysql.TransactionsPaginated]")
	}

	err = r.db.GetContext(ctx, &count, query, args...)
	return count, err

}

func (r *transactionRepository) TransactionsPaginated(ctx context.Context, itemID, accountID string, filters *ledger.TransactionFilter) ([]*ledger.Transaction, error) {

	stmt := sq.Select(transactionColumns...).
		From(transactionsTableName).
		Where(sq.Eq{
			"item_id":    itemID,
			"account_id": accountID,
		}).
		OrderBy("datetime desc")
	if filters != nil {
		if filters.FromTransactionID.Valid {
			// https://github.com/Masterminds/squirrel/issues/258#issuecomment-673315028
			stmt = stmt.Where(transactionIDSubQuery(filters.FromTransactionID.String))
		}
		if filters.Limit.Valid {
			stmt = stmt.Limit(filters.Limit.Uint64)
		}
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.TransactionsPaginated]")
	}

	fmt.Println(query)
	spew.Dump(args...)

	var transactions = make([]*ledger.Transaction, 0)
	err = r.db.SelectContext(ctx, &transactions, query, args...)

	return transactions, errors.Wrap(err, "[mysql.TransactionsPaginated]")

}

func transactionIDSubQuery(transactionID string) squirrel.Sqlizer {
	sql, args, _ := sq.Select("datetime").From(transactionsTableName).Where(sq.Eq{"transaction_id": transactionID}).ToSql()
	return sq.Expr(fmt.Sprintf("datetime < (%s)", sql), args...)
}

func (r *transactionRepository) TransactionsByTransactionIDs(ctx context.Context, itemID string, transactionIDs []string) ([]*ledger.Transaction, error) {

	query, args, err := sq.Select(transactionColumns...).From(transactionsTableName).Where(sq.Eq{"item_id": itemID, "transaction_id": transactionIDs}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.TransactionsByTransactionIDs]")
	}

	var transactions = make([]*ledger.Transaction, 0)
	err = r.db.SelectContext(ctx, &transactions, query, args...)

	return transactions, errors.Wrap(err, "[mysql.TransactionsByTransactionIDs]")

}

func (r *transactionRepository) CreateTransaction(ctx context.Context, transaction *ledger.Transaction) (*ledger.Transaction, error) {

	query, args, err := sq.Insert("transactions").Columns(transactionColumns...).
		Values(
			transaction.ItemID,
			transaction.AccountID,
			transaction.TransactionID,
			transaction.PendingTransactionID,
			transaction.CategoryID,
			transaction.Name,
			transaction.Pending,
			transaction.PaymentChannel,
			transaction.MerchantName,
			transaction.Categories,
			transaction.UnofficialCurrencyCode,
			transaction.ISOCurrencyCode,
			transaction.Amount,
			transaction.TransactionCode,
			transaction.AuthorizedDate,
			transaction.AuthorizedDateTime,
			transaction.Date,
			transaction.DateTime,
			sq.Expr(`NOW()`),
			sq.Expr(`NOW()`),
		).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.CreateTransaction]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.CreateTransaction]")
	}

	return r.Transaction(ctx, transaction.ItemID, transaction.TransactionID)

}

func (r *transactionRepository) UpdateTransaction(ctx context.Context, transactionID string, transaction *ledger.Transaction) (*ledger.Transaction, error) {

	query, args, err := sq.Update(transactionsTableName).
		Set("item_id", transaction.ItemID).
		Set("account_id", transaction.AccountID).
		Set("transaction_id", transaction.TransactionID).
		Set("pending_transaction_id", transaction.PendingTransactionID).
		Set("category_id", transaction.CategoryID).
		Set("name", transaction.Name).
		Set("pending", transaction.Pending).
		Set("payment_channel", transaction.PaymentChannel).
		Set("merchant_name", transaction.MerchantName).
		Set("categories", transaction.Categories).
		Set("unofficial_currency_code", transaction.UnofficialCurrencyCode).
		Set("iso_currency_code", transaction.ISOCurrencyCode).
		Set("amount", transaction.Amount).
		Set("transaction_code", transaction.TransactionCode).
		Set("authorized_date", transaction.AuthorizedDate).
		Set("authorized_datetime", transaction.AuthorizedDateTime).
		Set("date", transaction.Date).
		Set("datetime", transaction.DateTime).
		Set("updated_at", sq.Expr(`NOW()`)).
		Where(sq.Eq{"transaction_id": transaction.TransactionID}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.UpdateTransaction]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.UpdateTransaction]")
	}

	return r.Transaction(ctx, transaction.ItemID, transaction.TransactionID)
}
