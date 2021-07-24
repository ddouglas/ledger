// package transaction provides service access to account logic and repositories
package transaction

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"

	"github.com/ddouglas/ledger"
	"github.com/r3labs/diff"
	"github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"
)

type Service interface {
	TransactionReceiptPresignedURL(ctx context.Context, itemID, transactionID string) (string, error)
	AddReceiptToTransaction(ctx context.Context, itemID, transactionID, contentType string, receiptData []byte) error
	ProcessTransactions(ctx context.Context, item *ledger.Item, newTrans []*ledger.Transaction) error
	ledger.TransactionRepository
}

var allowedFileTypes = []string{
	"application/pdf", "image/jpeg",
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

		entry := s.logger.WithContext(ctx)
		entry = entry.WithFields(logrus.Fields{
			"id":   plaidTransaction.TransactionID,
			"date": plaidTransaction.Date.Format("2006-01-02"),
		})
		entry.Info("processing transaction")

		transaction, err := s.Transaction(ctx, item.ItemID, plaidTransaction.TransactionID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			entry.WithError(err).Error()
			return fmt.Errorf("failed to fetch transactions from DB")
		}

		if errors.Is(err, sql.ErrNoRows) {

			entry.Info("new transaction detected, fetching records for date")
			filters := &ledger.TransactionFilter{
				OnDate: null.NewTime(plaidTransaction.Date, true),
			}
			transactions, err := s.TransactionsPaginated(ctx, item.ItemID, plaidTransaction.AccountID, filters)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				entry.WithError(err).Error()
				return fmt.Errorf("failed to fetch transactions from DB")
			}

			entry = entry.WithField("count", len(transactions))
			entry = entry.WithError(err)
			plaidTransaction.ItemID = item.ItemID

			if err != nil && errors.Is(err, sql.ErrNoRows) || len(transactions) == 0 {
				entry.WithFields(logrus.Fields{
					"dateTime":       plaidTransaction.Date,
					"transaction_id": plaidTransaction.TransactionID,
				}).Info("no records exist for date, set dateTime to date")
				plaidTransaction.DateTime.SetValid(plaidTransaction.Date)
			}

			if err == nil && len(transactions) > 0 {
				entry.Info("found transactions, determining next timestamp")
				sort.SliceStable(transactions, func(i, j int) bool {
					return transactions[i].DateTime.Time.Unix() > transactions[j].DateTime.Time.Unix()
				})

				firstTransForDate := transactions[0]
				nextTransDatetime := firstTransForDate.DateTime.Time.Add(time.Second)
				plaidTransaction.DateTime.SetValid(nextTransDatetime)
				entry.WithFields(logrus.Fields{
					"dateTime":       nextTransDatetime,
					"transaction_id": plaidTransaction.TransactionID,
				}).Info("setting transaction datetime")
			}

			_, err = s.CreateTransaction(ctx, plaidTransaction)
			if err != nil {
				entry.WithError(err).Error()
				return fmt.Errorf("failed to insert transaction %s into DB", plaidTransaction.TransactionID)
			}
			entry.Info("transaction created successfully")

			if plaidTransaction.PendingTransactionID.Valid {
				entry = entry.WithField("pending_transaction_id", plaidTransaction.PendingTransactionID.String)
				entry.Info("transaction has pending transaction id associated with it")

				pendingTransaction, err := s.Transaction(ctx, plaidTransaction.ItemID, plaidTransaction.PendingTransactionID.String)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					entry.WithError(err).Error()
					return fmt.Errorf("unexpected error encountered querying for transactions, please check logs")
				}

				if err != nil && errors.Is(err, sql.ErrNoRows) {
					entry.Info("No Transactions found for provided pending transaction id. Skipping")
					continue
				}

				pendingTransaction.HiddenAt.SetValid(time.Now())
				_, err = s.UpdateTransaction(ctx, pendingTransaction.TransactionID, pendingTransaction)
				if err != nil {
					entry.WithError(err).Error()
					return fmt.Errorf("failed to update transaction %s", pendingTransaction.TransactionID)
				}

			}

			continue

		}

		entry.Info("existing transaction discovered, updating record")

		if !transaction.Pending {
			entry.Info("transactions is not pending, skipping")
			continue
		}

		changelog, err := diff.Diff(transaction, plaidTransaction)
		if err != nil {
			entry.WithError(err).Error()
			return fmt.Errorf("unable to determine updated attributes of transaction")
		}

		if len(changelog) == 0 {
			entry.Info("diff between plaidTransaction and transaction is 0, skipping update")
			continue
		}

		entry = entry.WithField("changelog", changelog)

		err = deepcopier.Copy(plaidTransaction).To(transaction)
		if err != nil {
			entry.WithError(err).Error()
			return fmt.Errorf("failed to copy plaidTransaction to ledgerTransaction")
		}

		_, err = s.UpdateTransaction(ctx, transaction.TransactionID, transaction)
		if err != nil {
			entry.WithError(err).Error()
			return fmt.Errorf("failed to update transaction %s", transaction.TransactionID)
		}
	}

	return nil

}

func (s *service) TransactionReceiptPresignedURL(ctx context.Context, itemID, transactionID string) (string, error) {

	urlStr, err := s.cache.FetchPresignedURL(ctx, transactionID)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch url from cache")
	}

	if urlStr != "" {
		return urlStr, nil
	}

	transaction, err := s.Transaction(ctx, itemID, transactionID)
	if err != nil {
		return "", errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to fetch transaction")
	}

	filename, err := transaction.Filename()
	if err != nil {
		return "", errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] transaction does not have a receipt file associated with it")
	}

	requestObj := s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    filename,
	}

	_, err = s.s3.GetObject(ctx, &requestObj)
	if err != nil {
		return "", errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to fetch receipt from object store")
	}

	psClient := s3.NewPresignClient(s.s3)

	expireDuration := time.Hour

	url, err := psClient.PresignGetObject(ctx, &requestObj, s3.WithPresignExpires(expireDuration))
	if err != nil {
		return "", errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to generate presigned url for object")
	}

	err = s.cache.CachePresignedURL(ctx, transactionID, url.URL, expireDuration)
	if err != nil {
		s.logger.WithError(err).Error("failed to cache url")
	}

	return url.URL, nil

}

func (s *service) AddReceiptToTransaction(ctx context.Context, itemID, transactionID, contentType string, receiptData []byte) error {

	transaction, err := s.Transaction(ctx, itemID, transactionID)
	if err != nil {
		return errors.Wrap(err, "[transaction.AddReceiptToTransaction] failed to fetch transaction")
	}

	err = validateContentType(contentType)
	if err != nil {
		return err
	}

	var ext string
	switch contentType {
	case "application/pdf":
		ext = ".pdf"
	case "image/jpeg":
		ext = "jpg"

	}

	reader := bytes.NewReader(receiptData)

	obj := s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fmt.Sprintf("%s.%s", transaction.TransactionID, ext)),
		Body:        reader,
		ContentType: aws.String(contentType),
	}

	_, err = s.s3.PutObject(ctx, &obj)
	if err != nil {
		return errors.Wrap(err, "[transaction.AddReceiptToTransaction] failed to write file to s3")
	}

	transaction.HasReceipt = true
	transaction.ReceiptType.SetValid(ext)

	_, err = s.UpdateTransaction(ctx, transaction.TransactionID, transaction)
	if err != nil {
		return errors.Wrap(err, "[transaction.AddReceiptToTransaction] failed to update has_receipt flag on transaction")
	}

	return nil

}
func validateContentType(contentType string) error {
	if contentType == "application/octet-stream" {
		return errors.New("unable to correctly determine content type from data format")
	}
	for _, allowedType := range allowedFileTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("%s is not an allowed file type, allowed types are: %s", contentType, strings.Join(allowedFileTypes, ", "))
}

// func sleep() {
// 	time.Sleep(time.Millisecond * 250)
// }

// func mapTransactionsByTransactionID(trans []*ledger.Transaction) map[string]*ledger.Transaction {
// 	mapTransactions := make(map[string]*ledger.Transaction)
// 	for _, tran := range trans {
// 		mapTransactions[tran.TransactionID] = tran
// 	}
// 	return mapTransactions
// }

// for i, tran := range newTrans {
// 	fmt.Printf("Index: %d Date: %s DateTime: %s\n", i, tran.Date.Format("2006-01-02"), tran.DateTime.Time.Format("2006-01-02 15:04:05"))
// }
// transactionMap := make(map[string][]*ledger.Transaction)
// 	const dateFmt = "2006-01-02"
// 	for _, transaction := range newTrans {

// 		if _, ok := transactionMap[transaction.Date.Format(dateFmt)]; !ok {
// 			transactionMap[transaction.Date.Format(dateFmt)] = make([]*ledger.Transaction, 0, 10)
// 		}

// 		transactionMap[transaction.Date.Format(dateFmt)] = append(transactionMap[transaction.Date.Format(dateFmt)], transaction)

// 	}
// 	modifiedTransactions := make([]*ledger.Transaction, 0, len(newTrans))
// 	for _, transactions := range transactionMap {
// 		numTransactions := len(transactions)
// 		fmt.Println(numTransactions)
// 		for i, transaction := range transactions {
// 			if i == 0 {
// 				transactions[i].DateTime.SetValid(transactions[i].Date)
// 				continue
// 			}

// 			prevTransaction := transactions[i-1]
// 			transactions[i].DateTime.SetValid(prevTransaction.DateTime.Time.Add(time.Second))
// 			fmt.Printf("Index: %d Date: %s DateTime: %s\n", i, transaction.Date.Format("2006-01-02"), transaction.DateTime.Time.Format("2006-01-02 15:04:05"))
// 			modifiedTransactions = append(modifiedTransactions, transaction)
// 		}

// 	}

// 	sort.SliceStable(modifiedTransactions, func(i, j int) bool {
// 		return modifiedTransactions[i].DateTime.Time.Unix() < modifiedTransactions[j].DateTime.Time.Unix()
// 	})
