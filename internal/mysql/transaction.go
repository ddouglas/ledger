package mysql

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
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
	"has_receipt",
	"receipt_type",
	"payment_channel",
	"merchant_id",
	"unofficial_currency_code",
	"iso_currency_code",
	"amount",
	"transaction_code",
	"authorized_date",
	"authorized_datetime",
	"date",
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

func (r *transactionRepository) TransactionsCount(ctx context.Context, itemID, accountID string, filters *ledger.TransactionFilter) (uint64, error) {

	var count uint64
	stmt := sq.Select(`COUNT(*)`).From(transactionsTableName).Where(sq.Eq{
		"item_id":    itemID,
		"account_id": accountID,
	})
	xfilters := *filters
	if xfilters.Limit.Valid {
		xfilters.Limit = null.NewUint64(0, false)
	}
	stmt = transactionsQueryBuilder(stmt, &xfilters)

	query, args, err := stmt.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "[mysql.TransactionsCount]")
	}

	err = r.db.GetContext(ctx, &count, query, args...)
	return count, errors.Wrap(err, "[mysql.TransactionsCount]")

}

func (r *transactionRepository) TransactionsPaginated(ctx context.Context, itemID, accountID string, filters *ledger.TransactionFilter) ([]*ledger.Transaction, error) {

	stmt := sq.Select(transactionColumns...).
		From(transactionsTableName).
		Where(sq.Eq{
			"item_id":    itemID,
			"account_id": accountID,
		}).
		OrderBy("date desc, amount asc")
	stmt = transactionsQueryBuilder(stmt, filters)
	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.TransactionsPaginated]")
	}

	var transactions = make([]*ledger.Transaction, 0)
	err = r.db.SelectContext(ctx, &transactions, query, args...)

	return transactions, errors.Wrap(err, "[mysql.TransactionsPaginated]")

}

func transactionsQueryBuilder(stmt sq.SelectBuilder, filters *ledger.TransactionFilter) sq.SelectBuilder {
	if filters != nil {
		if filters.FromTransactionID.Valid {
			// https://github.com/Masterminds/squirrel/issues/258#issuecomment-673315028
			stmt = stmt.Where(transactionIDSubQuery(filters.FromTransactionID.String))
		}
		if filters.CategoryID.Valid {
			stmt = stmt.Where(sq.Eq{"category_id": filters.CategoryID.String})
		}
		if filters.MerchantID.Valid {
			stmt = stmt.Where(sq.Eq{"merchant_id": filters.MerchantID.String})
		}
		if filters.Limit.Valid {
			stmt = stmt.Limit(filters.Limit.Uint64)
		}
		if filters.EndDate.Valid {
			endDate := map[string]interface{}{"date": filters.EndDate.Time.Format("2006-01-02")}
			var op sq.Sqlizer = sq.Lt(endDate)
			if filters.DateInclusive.Valid && filters.DateInclusive.Bool {
				op = sq.LtOrEq(endDate)
			}
			stmt = stmt.Where(op)
		}
		if filters.StartDate.Valid {
			endDate := map[string]interface{}{"date": filters.StartDate.Time.Format("2006-01-02")}
			var op sq.Sqlizer = sq.Gt(endDate)
			if filters.DateInclusive.Valid && filters.DateInclusive.Bool {
				op = sq.GtOrEq(endDate)
			}
			stmt = stmt.Where(op)
		}
		if filters.OnDate.Valid {
			stmt = stmt.Where(sq.Eq{"date": filters.OnDate.Time})
		}
		if filters.AmountDir.Valid {
			if filters.AmountDir.Float64 > 0 {
				stmt = stmt.Where(sq.Gt{"amount": 0})
			}
			if filters.AmountDir.Float64 < 0 {
				stmt = stmt.Where(sq.Lt{"amount": 0})
			}
		}
	}

	// Never fetch hidden transactions
	stmt = stmt.Where(sq.Eq{"hidden_at": nil}).Where(sq.Eq{"deleted_at": nil})

	return stmt
}

func transactionIDSubQuery(transactionID string) squirrel.Sqlizer {
	sql, args, _ := sq.Select("datetime").From(transactionsTableName).Where(sq.Eq{"transaction_id": transactionID}).ToSql()
	return sq.Expr(fmt.Sprintf("datetime < (%s)", sql), args...)
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
			transaction.HasReceipt,
			transaction.ReceiptType,
			transaction.PaymentChannel,
			transaction.MerchantID,
			transaction.UnofficialCurrencyCode,
			transaction.ISOCurrencyCode,
			transaction.Amount,
			transaction.TransactionCode,
			transaction.AuthorizedDate,
			transaction.AuthorizedDateTime,
			transaction.Date,
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

	query, args, err := sq.Update(transactionsTableName).SetMap(map[string]interface{}{
		"pending_transaction_id":   transaction.PendingTransactionID,
		"category_id":              transaction.CategoryID,
		"name":                     transaction.Name,
		"pending":                  transaction.Pending,
		"has_receipt":              transaction.HasReceipt,
		"receipt_type":             transaction.ReceiptType,
		"payment_channel":          transaction.PaymentChannel,
		"merchant_id":              transaction.MerchantID,
		"unofficial_currency_code": transaction.UnofficialCurrencyCode,
		"iso_currency_code":        transaction.ISOCurrencyCode,
		"amount":                   transaction.Amount,
		"transaction_code":         transaction.TransactionCode,
		"authorized_date":          transaction.AuthorizedDate,
		"authorized_datetime":      transaction.AuthorizedDateTime,
		"date":                     transaction.Date,
		"updated_at":               sq.Expr(`NOW()`),
	}).Where(sq.Eq{
		"transaction_id": transaction.TransactionID,
	}).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.UpdateTransaction]")
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "[mysql.UpdateTransaction]")
	}

	return r.Transaction(ctx, transaction.ItemID, transaction.TransactionID)
}

func (r *transactionRepository) UpdateTransactionMerchantTx(ctx context.Context, tx ledger.Transactioner, byMerchantID, toMerchantID string) error {

	txn, ok := tx.(*transaction)
	if !ok {
		return ErrInvalidTransaction
	}

	query, args, err := sq.Update(transactionsTableName).SetMap(map[string]interface{}{
		"merchant_id": toMerchantID,
	}).Where(sq.Eq{
		"merchant_id": byMerchantID,
	}).ToSql()
	if err != nil {
		return errors.Wrap(err, "[mysql.UpdateTransactionMerchant]")
	}

	_, err = txn.ExecContext(ctx, query, args...)

	return errors.Wrap(err, "[mysql.UpdateTransactionMerchant]")

}
