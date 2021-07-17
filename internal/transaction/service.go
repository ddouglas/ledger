// package transaction provides service access to account logic and repositories
package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ddouglas/ledger"
	"github.com/ulule/deepcopier"
)

type Service interface {
	ProcessTransactions(ctx context.Context, item *ledger.Item, newTrans []*ledger.Transaction) error
	ledger.TransactionRepository
}

func New(optFuncs ...configOption) Service {
	s := &service{}
	for _, optFunc := range optFuncs {
		optFunc(s)
	}
	return s
}

func (s *service) ProcessTransactions(ctx context.Context, item *ledger.Item, newTrans []*ledger.Transaction) error {

	for _, plaidTransaction := range newTrans {

		transaction, err := s.Transaction(ctx, item.ItemID, plaidTransaction.TransactionID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to fetch transactions from DB")
		}

		if err == sql.ErrNoRows {

			plaidTransaction.ItemID = item.ItemID

			_, err := s.CreateTransaction(ctx, plaidTransaction)
			if err != nil {
				return fmt.Errorf("failed to insert transaction %s into DB: %w", plaidTransaction.TransactionID, err)
			}

			continue

		}

		err = deepcopier.Copy(plaidTransaction).To(transaction)
		if err != nil {
			return fmt.Errorf("failed to copy plaidTransaction to ledgerTransaction:%w", err)
		}

		_, err = s.UpdateTransaction(ctx, transaction.TransactionID, transaction)
		if err != nil {
			return fmt.Errorf("failed to update transaction %s: %w", transaction.TransactionID, err)
		}

		// if !transaction.Location.IsEmpty() {
		// 	_, err = s.
		// }

	}

	return nil

}

func mapTransactionsByTransactionID(trans []*ledger.Transaction) map[string]*ledger.Transaction {
	mapTransactions := make(map[string]*ledger.Transaction)
	for _, tran := range trans {
		mapTransactions[tran.TransactionID] = tran
	}
	return mapTransactions
}
