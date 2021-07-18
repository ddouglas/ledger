package mysql

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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

const tableName = "transactions"

func NewTransactionRepository(db *sqlx.DB) ledger.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Transaction(ctx context.Context, itemID, transactionID string) (*ledger.Transaction, error) {

	query, args, err := sq.Select(transactionColumns...).From(tableName).Where(sq.Eq{
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

func (r *transactionRepository) TransactionsByAccountID(ctx context.Context, itemID, accountID string, filters *ledger.TransactionFilter) ([]*ledger.Transaction, error) {

	stmt := sq.Select(transactionColumns...).
		From(tableName).
		Where(sq.Eq{
			"item_id":    itemID,
			"account_id": accountID,
		}).
		OrderBy("date desc", "pending desc")
	if filters != nil {
		if filters.FromIterator.Number > 0 {
			stmt = stmt.Where(sq.And)
		}
		if filters.Count > 0 {
			stmt = stmt.Limit(filters.Count)
		}
	}

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.TransactionsByAccountID]")
	}

	var transactions = make([]*ledger.Transaction, 0)
	err = r.db.SelectContext(ctx, &transactions, query, args...)

	return transactions, errors.Wrap(err, "[mysql.TransactionsByAccountID]")

}

func (r *transactionRepository) TransactionsByTransactionIDs(ctx context.Context, itemID string, transactionIDs []string) ([]*ledger.Transaction, error) {

	query, args, err := sq.Select(transactionColumns...).From(tableName).Where(sq.Eq{"item_id": itemID, "transaction_id": transactionIDs}).ToSql()
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

	query, args, err := sq.Update(tableName).
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
