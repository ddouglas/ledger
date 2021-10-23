// package transaction provides service access to account logic and repositories
package transaction

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/cache"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/r3labs/diff"
	"github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"
)

type Service interface {
	ProcessTransactions(ctx context.Context, item *ledger.Item, newTrans []*ledger.Transaction) error
	TransactionReceiptPresignedURL(ctx context.Context, itemID, transactionID string) (string, error)
	AddReceiptToTransaction(ctx context.Context, itemID, transactionID string, buffer *bytes.Buffer) error
	RemoveReceiptFromTransaction(ctx context.Context, itemID, transactionID string) error
	ledger.TransactionRepository
}

type service struct {
	logger  *logrus.Logger
	cache   cache.Service
	s3      *s3.Client
	gateway gateway.Service
	bucket  string

	ledger.TransactionRepository
	ledger.MerchantRepository
}

var allowedFileTypes = []string{
	"application/pdf", "image/jpeg",
}

func New(
	s3 *s3.Client,
	logger *logrus.Logger,
	gateway gateway.Service,
	cache cache.Service,
	bucket string,
	transaction ledger.TransactionRepository,
	merchants ledger.MerchantRepository,
) Service {
	return &service{
		gateway:               gateway,
		cache:                 cache,
		s3:                    s3,
		bucket:                bucket,
		TransactionRepository: transaction,
		MerchantRepository:    merchants,
		logger:                logger,
	}

}

func (s *service) ProcessTransactions2(ctx context.Context, item *ledger.Item, newTrans []*ledger.Transaction) error {

	for _, plaidTransaction := range newTrans {

		entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
			"id":   plaidTransaction.TransactionID,
			"date": plaidTransaction.Date.Format("2006-01-02"),
		})

		transaction, err := s.Transaction(ctx, item.ItemID, plaidTransaction.TransactionID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			entry.WithError(err).Error()
			return errors.New("failed to fetch transactions from DB")
		}

		if errors.Is(err, sql.ErrNoRows) {

			entry.Info("new transaction detected, fetching records for date")
			filters := &ledger.TransactionFilter{
				OnDate: null.TimeFrom(plaidTransaction.Date),
			}
			transactions, err := s.TransactionsPaginated(ctx, item.ItemID, plaidTransaction.AccountID, filters)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				entry.WithError(err).Error()
				return errors.New("failed to fetch transactions from DB")
			}

			entry = entry.WithField("count", len(transactions)).WithError(err)
			plaidTransaction.ItemID = item.ItemID

			if err != nil && errors.Is(err, sql.ErrNoRows) || len(transactions) == 0 {
				entry.WithFields(logrus.Fields{
					"dateTime":       plaidTransaction.Date,
					"transaction_id": plaidTransaction.TransactionID,
				}).Info("no records exist for date, set dateTime to date")
				date := plaidTransaction.Date
				if date.IsZero() {
					now := time.Now()
					date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
				}
				plaidTransaction.DateTime.SetValid(date)
			}

			if err == nil && len(transactions) > 0 {
				entry.Info("found transactions, determining next timestamp")
				sort.SliceStable(transactions, func(i, j int) bool {
					return transactions[i].DateTime.Time.Unix() > transactions[j].DateTime.Time.Unix()
				})

				firstTransForDate := transactions[0]
				var nextTransDatetime time.Time
				if firstTransForDate.DateTime.Valid && !firstTransForDate.DateTime.Time.IsZero() {
					nextTransDatetime = firstTransForDate.DateTime.Time.Add(time.Second)
				} else {
					now := time.Now()
					nextTransDatetime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
				}
				plaidTransaction.DateTime.SetValid(nextTransDatetime)
				entry.WithFields(logrus.Fields{
					"dateTime":       nextTransDatetime,
					"transaction_id": plaidTransaction.TransactionID,
				}).Info("setting transaction datetime")
			}

			_, err = s.CreateTransaction(ctx, plaidTransaction)
			if err != nil {
				entry.WithError(err).Error()
				return errors.Errorf("failed to insert transaction %s into DB", plaidTransaction.TransactionID)
			}
			entry.Info("transaction created successfully")

			if plaidTransaction.PendingTransactionID.Valid {
				entry = entry.WithField("pending_transaction_id", plaidTransaction.PendingTransactionID.String)
				entry.Info("transaction has pending transaction id associated with it")

				pendingTransaction, err := s.Transaction(ctx, plaidTransaction.ItemID, plaidTransaction.PendingTransactionID.String)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					entry.WithError(err).Error()
					return errors.New("unexpected error encountered querying for transactions, please check logs")
				}

				if err != nil && errors.Is(err, sql.ErrNoRows) {
					entry.Info("No Transactions found for provided pending transaction id. Skipping")
					continue
				}

				pendingTransaction.HiddenAt.SetValid(time.Now())
				_, err = s.UpdateTransaction(ctx, pendingTransaction.TransactionID, pendingTransaction)
				if err != nil {
					entry.WithError(err).Error()
					return errors.Errorf("failed to update transaction %s", pendingTransaction.TransactionID)
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
			return errors.New("unable to determine updated attributes of transaction")
		}

		if len(changelog) == 0 {
			entry.Info("diff between plaidTransaction and transaction is 0, skipping update")
			continue
		}

		entry = entry.WithField("changelog", changelog)

		err = deepcopier.Copy(plaidTransaction).To(transaction)
		if err != nil {
			entry.WithError(err).Error()
			return errors.Errorf("failed to copy plaidTransaction to ledgerTransaction")
		}

		_, err = s.UpdateTransaction(ctx, transaction.TransactionID, transaction)
		if err != nil {
			entry.WithError(err).Error()
			return errors.Errorf("failed to update transaction %s", transaction.TransactionID)
		}
	}

	return nil

}

func (s *service) ProcessTransactions(ctx context.Context, item *ledger.Item, newTrans []*ledger.Transaction) error {

	for _, plaidTransaction := range newTrans {
		err := s.processTransaction(ctx, item, plaidTransaction)
		if err != nil {
			s.logger.WithError(err).Error("failed to process transaction")
		}
	}

	return nil

}

func (s *service) processTransaction(ctx context.Context, item *ledger.Item, plaidTransaction *ledger.Transaction) error {
	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"id":   plaidTransaction.TransactionID,
		"date": plaidTransaction.Date.Format("2006-01-02"),
	})

	transaction, err := s.Transaction(ctx, item.ItemID, plaidTransaction.TransactionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		entry.WithError(err).Error()
		return errors.New("failed to fetch transactions from DB")
	}

	if errors.Is(err, sql.ErrNoRows) {

		plaidTransaction.ItemID = item.ItemID

		err = s.handleTransactionMerchant(ctx, plaidTransaction)
		if err != nil {
			entry.WithError(err).Error()
			return errors.Wrap(err, "failed to process merchant")
		}

		_, err = s.CreateTransaction(ctx, plaidTransaction)
		if err != nil {
			entry.WithError(err).Error()
			return errors.Errorf("failed to insert transaction %s into DB", plaidTransaction.TransactionID)
		}

		if plaidTransaction.PendingTransactionID.Valid {
			entry = entry.WithField("pending_transaction_id", plaidTransaction.PendingTransactionID.String)

			pendingTransaction, err := s.Transaction(ctx, plaidTransaction.ItemID, plaidTransaction.PendingTransactionID.String)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				entry.WithError(err).Error()
				return errors.New("unexpected error encountered querying for transactions, please check logs")
			}

			if err != nil && errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			pendingTransaction.HiddenAt.SetValid(time.Now())
			_, err = s.UpdateTransaction(ctx, pendingTransaction.TransactionID, pendingTransaction)
			if err != nil {
				entry.WithError(err).Error()
				return errors.Errorf("failed to update transaction %s", pendingTransaction.TransactionID)
			}

			entry.Info("pending transaction updated successfully")

		}

		entry.Info("transaction created successfully")

		return nil

	}

	if !transaction.Pending {
		return nil
	}

	entry.Info("existing transaction discovered, updating record")

	changelog, err := diff.Diff(transaction, plaidTransaction)
	if err != nil {
		entry.WithError(err).Error()
		return errors.New("unable to determine updated attributes of transaction")
	}

	if len(changelog) == 0 {
		return nil
	}

	err = deepcopier.Copy(plaidTransaction).To(transaction)
	if err != nil {
		entry.WithError(err).Error()
		return errors.Errorf("failed to copy plaidTransaction to ledgerTransaction")
	}

	_, err = s.UpdateTransaction(ctx, transaction.TransactionID, transaction)
	if err != nil {
		entry.WithError(err).Error()
		return errors.Errorf("failed to update transaction %s", transaction.TransactionID)
	}

	entry.Info("transaction updated successfully")
	return nil
}

func (s *service) handleTransactionMerchant(ctx context.Context, transaction *ledger.Transaction) error {

	merchantName := transaction.MerchantName.String
	if merchantName == "" {
		merchantName = "Unknown"
	}

	merchant, err := s.MerchantByAlias(ctx, merchantName)
	if err == nil {
		transaction.MerchantID = merchant.ID
		return nil
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Error is sql.ErrNoRows, need to create an merchant and a merchant alias
	merchant = &ledger.Merchant{
		ID:   randString(32),
		Name: merchantName,
	}

	alias := &ledger.MerchantAlias{
		AliasID:    randString(32),
		MerchantID: merchant.ID,
		Alias:      merchantName,
	}

	_, err = s.MerchantRepository.CreateMerchant(ctx, merchant)
	if err != nil {
		return err
	}

	_, err = s.MerchantRepository.CreateMerchantAlias(ctx, alias)
	if err != nil {
		return err
	}

	transaction.MerchantID = merchant.ID
	return nil

}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[src.Int63()%int64(len(letterBytes))]
	}
	return string(b)
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

	if len(strings.Split(filename, ".")) != 2 {
		return "", errors.New("[transaction.TransactionReceiptPresignedURL] unable to determine file name")
	}

	_, err = s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		return "", errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to fetch receipt from object store")
	}

	psClient := s3.NewPresignClient(s.s3)

	expireDuration := time.Hour

	url, err := psClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}, s3.WithPresignExpires(expireDuration))
	if err != nil {
		return "", errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to generate presigned url for object")
	}

	err = s.cache.CachePresignedURL(ctx, transactionID, url.URL, expireDuration)
	if err != nil {
		s.logger.WithError(err).Error("failed to cache url")
	}

	return url.URL, nil

}

func (s *service) AddReceiptToTransaction(ctx context.Context, itemID, transactionID string, buffer *bytes.Buffer) error {

	transaction, err := s.Transaction(ctx, itemID, transactionID)
	if err != nil {
		return errors.Wrap(err, "[transaction.AddReceiptToTransaction] failed to fetch transaction")
	}

	contentType := http.DetectContentType(buffer.Bytes())
	err = validateContentType(contentType)
	if err != nil {
		return err
	}

	var ext string
	switch contentType {
	case "application/pdf":
		ext = "pdf"
	case "image/jpeg":
		ext = "jpg"
	}

	obj := s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fmt.Sprintf("%s.%s", transaction.TransactionID, ext)),
		Body:        bytes.NewReader(buffer.Bytes()),
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

func (s *service) RemoveReceiptFromTransaction(ctx context.Context, itemID, transactionID string) error {

	transaction, err := s.Transaction(ctx, itemID, transactionID)
	if err != nil {
		return errors.Wrap(err, "[transaction.RemoveReceiptFromTransaction] failed to fetch transaction")
	}

	if !transaction.HasReceipt {
		return nil
	}

	input := s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s.%s", transaction.TransactionID, transaction.ReceiptType.String)),
	}

	_, err = s.s3.DeleteObject(ctx, &input)
	if err != nil {
		return errors.Wrap(err, "[transaction.RemoveReceiptFromTransaction] failed to delete transaction with S3")
	}

	err = s.cache.DeletePresignURL(ctx, transactionID)
	if err != nil {
		return errors.Wrap(err, "[transaction.RemoveReceiptFromTransaction] failed to delete transaction in Cache")
	}

	transaction.HasReceipt = false
	transaction.ReceiptType = null.NewString("", false)

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
