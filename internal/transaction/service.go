// package transaction provides service access to account logic and repositories
package transaction

import (
	"context"
	"fmt"
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

	// sort.SliceStable(newTrans, func(i, j int) bool {

	// 	// var next = newTrans[j]
	// 	newTrans[i].DateTime.SetValid(newTrans[i].Date)
	// 	newTrans[j].DateTime.SetValid(newTrans[j].Date)

	// 	if newTrans[i].DateTime.Time == newTrans[j].DateTime.Time {
	// 		newTrans[j].DateTime.SetValid(newTrans[j].DateTime.Time.Add(time.Second))
	// 	}

	// 	return newTrans[i].DateTime.Time.Unix() > newTrans[j].DateTime.Time.Unix()

	// })

	transactionMap := make(map[string][]*ledger.Transaction)
	const dateFmt = "2006-01-02"
	for _, transaction := range newTrans {

		if _, ok := transactionMap[transaction.Date.Format(dateFmt)]; !ok {
			transactionMap[transaction.Date.Format(dateFmt)] = make([]*ledger.Transaction, 0, 10)
		}

		transactionMap[transaction.Date.Format(dateFmt)] = append(transactionMap[transaction.Date.Format(dateFmt)], transaction)

	}
	modifiedTransactions := make([]*ledger.Transaction, 0, len(newTrans))
	for _, transactions := range transactionMap {
		numTransactions := len(transactions)
		fmt.Println(numTransactions)
		for i, transaction := range transactions {
			if i == 0 {
				transactions[i].DateTime.SetValid(transactions[i].Date)
				continue
			}

			prevTransaction := transactions[i-1]
			transactions[i].DateTime.SetValid(prevTransaction.DateTime.Time.Add(time.Second))
			fmt.Printf("Index: %d Date: %s DateTime: %s\n", i, transaction.Date.Format("2006-01-02"), transaction.DateTime.Time.Format("2006-01-02 15:04:05"))
		}

	}

	// for i, tran := range newTrans {
	// 	fmt.Printf("Index: %d Date: %s DateTime: %s\n", i, tran.Date.Format("2006-01-02"), tran.DateTime.Time.Format("2006-01-02 15:04:05"))
	// }

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
