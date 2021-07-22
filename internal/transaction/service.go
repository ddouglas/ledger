// package transaction provides service access to account logic and repositories
package transaction

import (
	"context"
	"time"

	"github.com/ddouglas/ledger"
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

	// for _, plaidTransaction := range newTrans {

	// 	entry := s.logger.WithContext(ctx)
	// 	entry = entry.WithFields(logrus.Fields{
	// 		"id":   plaidTransaction.TransactionID,
	// 		"date": plaidTransaction.Date.Format("2006-01-02"),
	// 	})
	// 	entry.Info("processing transaction")

	// 	transaction, err := s.Transaction(ctx, item.ItemID, plaidTransaction.TransactionID)
	// 	if err != nil && !errors.Is(err, sql.ErrNoRows) {
	// 		entry.WithError(err).Error()
	// 		return fmt.Errorf("failed to fetch transactions from DB")
	// 	}

	// 	if errors.Is(err, sql.ErrNoRows) {

	// 		entry.Info("new transaction detected, creating record")

	// 		plaidTransaction.ItemID = item.ItemID

	// 		_, err := s.CreateTransaction(ctx, plaidTransaction)
	// 		if err != nil {
	// 			entry.WithError(err).Error()
	// 			return fmt.Errorf("failed to insert transaction %s into DB", plaidTransaction.TransactionID)
	// 		}

	// 		sleep()
	// 		sleep()
	// 		continue

	// 	}

	// 	entry.Info("existing transaction discover, updating record")

	// 	if !transaction.Pending {
	// 		entry.Info("transactions is not pending, skipping")
	// 		// sleep()
	// 		continue
	// 	}

	// 	changelog, err := diff.Diff(transaction, plaidTransaction)
	// 	if err != nil {
	// 		entry.WithError(err).Error()
	// 		return fmt.Errorf("unable to determine updated attributes of transaction")
	// 	}

	// 	if len(changelog) == 0 {
	// 		entry.Info("diff between plaidTransaction and transaction is 0, skipping update")

	// 		// sleep()
	// 		continue
	// 	}

	// 	entry = entry.WithField("changelog", changelog)

	// 	err = deepcopier.Copy(plaidTransaction).To(transaction)
	// 	if err != nil {
	// 		entry.WithError(err).Error()
	// 		return fmt.Errorf("failed to copy plaidTransaction to ledgerTransaction")
	// 	}

	// 	_, err = s.UpdateTransaction(ctx, transaction.TransactionID, transaction)
	// 	if err != nil {
	// 		entry.WithError(err).Error()
	// 		return fmt.Errorf("failed to update transaction %s", transaction.TransactionID)
	// 	}

	// 	// sleep()

	// }

	return nil

}

func sleep() {
	time.Sleep(time.Millisecond * 250)
}

func (s *service) TransactionsByAccountID(ctx context.Context, itemID, accountID string, filters *ledger.TransactionFilter) ([]*ledger.Transaction, error) {

	// if filters != nil && filters.FromTransactionID != nil {
	// 	// transaction, err := s.Transaction(ctx, itemID, filters.FromTransactionID.String)
	// 	// if err != nil {
	// 	// 	s.logger.WithError(err).Errorln()
	// 	// 	return nil, errors.New("unable to filter on unknown transaction")
	// 	// }

	// 	// filters.FromIterator, err = ledger.NewNumberFilter(ledger.LtOperation, int64(transaction.Iterator))
	// 	// if err != nil {
	// 	// 	return nil, err
	// 	// }
	// 	// filters.FromTransactionID = nil
	// }

	return s.TransactionRepository.TransactionsByAccountID(ctx, itemID, accountID, filters)

}

// func mapTransactionsByTransactionID(trans []*ledger.Transaction) map[string]*ledger.Transaction {
// 	mapTransactions := make(map[string]*ledger.Transaction)
// 	for _, tran := range trans {
// 		mapTransactions[tran.TransactionID] = tran
// 	}
// 	return mapTransactions
// }
