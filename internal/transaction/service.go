// package transaction provides service access to account logic and repositories
package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
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
	ConvertMerchantToAlias(ctx context.Context, parentMerchantID, childMerchantID string) (*ledger.Merchant, error)
	ProcessTransactions(ctx context.Context, item *ledger.Item, newTrans []*ledger.Transaction) error
	TransactionReceiptPresignedURL(ctx context.Context, itemID, transactionID string) (*ledger.TransactionReceipt, error)
	AddReceiptToTransaction(ctx context.Context, itemID, transactionID string, file graphql.Upload) error
	RemoveReceiptFromTransaction(ctx context.Context, itemID, transactionID string) error
	ledger.TransactionRepository
	ledger.MerchantRepository
}

type service struct {
	logger  *logrus.Logger
	cache   cache.Service
	s3      *s3.Client
	gateway gateway.Service
	bucket  string
	starter ledger.Starter

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
	starter ledger.Starter,
	transaction ledger.TransactionRepository,
	merchants ledger.MerchantRepository,
) Service {
	return &service{
		gateway:               gateway,
		cache:                 cache,
		s3:                    s3,
		bucket:                bucket,
		starter:               starter,
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
			plaidTransaction.ItemID = item.ItemID

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

func (s *service) TransactionReceiptPresignedURL(ctx context.Context, itemID, transactionID string) (*ledger.TransactionReceipt, error) {

	// urlStr, err := s.cache.FetchPresignedURL(ctx, transactionID)
	// if err != nil {
	// 	s.logger.WithError(err).Error("failed to fetch url from cache")
	// }

	// if urlStr != "" {
	// 	return urlStr, nil
	// }

	transaction, err := s.Transaction(ctx, itemID, transactionID)
	if err != nil {
		return nil, errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to fetch transaction")
	}

	psClient := s3.NewPresignClient(s.s3)
	expireDuration := time.Minute * 10

	receipt := new(ledger.TransactionReceipt)

	_, err = s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(transaction.Filename()),
	})
	if err == nil {
		get, err := psClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(transaction.Filename()),
		}, s3.WithPresignExpires(expireDuration))
		if err != nil {
			return nil, errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to generate presigned get url for object")
		}

		receipt.Get = null.NewString(get.URL, get.URL != "")
	}

	put, err := psClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(transaction.Filename()),
	}, s3.WithPresignExpires(expireDuration))
	if err != nil {
		return nil, errors.Wrap(err, "[transaction.TransactionReceiptPresignedURL] failed to generate presigned put url for object")
	}

	receipt.Put = null.NewString(put.URL, put.URL != "")
	return receipt, nil

}

func (s *service) AddReceiptToTransaction(ctx context.Context, itemID, transactionID string, file graphql.Upload) error {

	transaction, err := s.Transaction(ctx, itemID, transactionID)
	if err != nil {
		return errors.Wrap(err, "[transaction.AddReceiptToTransaction] failed to fetch transaction")
	}

	err = validateContentType(file.ContentType)
	if err != nil {
		return err
	}

	obj := s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(file.Filename),
		Body:        file.File,
		ContentType: aws.String(file.ContentType),
	}

	_, err = s.s3.PutObject(ctx, &obj)
	if err != nil {
		return errors.Wrap(err, "[transaction.AddReceiptToTransaction] failed to write file to s3")
	}

	var ext string
	switch file.ContentType {
	case "application/pdf":
		ext = "pdf"
	case "image/jpeg":
		ext = "jpg"
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

	_, err = s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(transaction.Filename()),
	})
	if err != nil {
		// This should be because the file does not exist.
		s.logger.WithError(err).Error("failed to head object")
		return nil
	}

	_, err = s.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(transaction.Filename()),
	})
	if err != nil {
		return errors.Wrap(err, "[transaction.RemoveReceiptFromTransaction] failed to delete transaction with S3")
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
