package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/r3labs/diff/v2"
)

func (s *server) handleGetAccountTransactions(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		err := errors.New("itemID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify ownership"))
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		err := errors.New("accountID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err = s.account.Account(ctx, itemID, accountID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch account"))
		return
	}

	var filters = new(ledger.TransactionFilter)
	err = filters.BuildFromURLValues(r.URL.Query())
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	var results = new(ledger.PaginatedTransactions)

	results.Transactions, err = s.transaction.TransactionsPaginated(ctx, itemID, accountID, filters)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transactions"))
		return
	}

	results.Total, err = s.transaction.TransactionsCount(ctx, itemID, accountID, filters)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transaction count"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, results)

}

// handleUpdateTransactions handles triggering importer to refresh transactions within a specific date range for a specific account
func (s *server) handleUpdateTransactions(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var message = new(ledger.WebhookMessage)
	err := json.NewDecoder(r.Body).Decode(message)
	if err != nil {
		s.logger.WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to decode request body"))
		return
	}

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("itemID is required"))
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("accountID is required"))
		return
	}

	message.ItemID = itemID
	message.Options = &ledger.WebhookMessageOptions{
		AccountIDs: []string{accountID},
	}

	err = s.importer.PublishCustomWebhookMessage(ctx, message)
	if err != nil {
		s.logger.WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to process refresh request"))
		return
	}

	s.writeResponse(ctx, w, http.StatusNoContent, nil)
}

func (s *server) handleGetAccountTransaction(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		err := errors.New("itemID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify ownership of item"))
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		err := errors.New("accountID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err = s.account.Account(ctx, itemID, accountID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch account"))
		return
	}

	transactionID := chi.URLParam(r, "transactionID")
	if transactionID == "" {
		err := errors.New("transactionID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	transaction, err := s.transaction.Transaction(ctx, itemID, transactionID)
	if err != nil {

		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transaction with provided itemID and transactionID combo"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, transaction)

}

func (s *server) handlePatchAccountTransaction(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		err := errors.New("itemID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify ownership of item"))
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		err := errors.New("accountID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err = s.account.Account(ctx, itemID, accountID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch account"))
		return
	}

	transactionID := chi.URLParam(r, "transactionID")
	if transactionID == "" {
		err := errors.New("transactionID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	transaction, err := s.transaction.Transaction(ctx, itemID, transactionID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transaction with provided itemID and transactionID combo"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to read request body"))
		return
	}

	var requestTransactions = new(ledger.Transaction)
	err = json.Unmarshal(body, requestTransactions)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to parse request body as json"))
		return
	}

	differ, _ := diff.NewDiffer()

	changelog, err := differ.Diff(transaction, requestTransactions)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to properly diff existing and provided items"))
		return
	}

	spew.Dump(changelog)

	s.writeResponse(ctx, w, http.StatusOK, transaction)

}

func (s *server) handleGetAccountTransactionReceiptURL(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()
	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		err := errors.New("itemID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify ownership of item"))
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		err := errors.New("accountID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err = s.account.Account(ctx, itemID, accountID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch account"))
		return
	}

	transactionID := chi.URLParam(r, "transactionID")
	if transactionID == "" {
		err := errors.New("transactionID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	presigned, err := s.transaction.TransactionReceiptPresignedURL(ctx, itemID, transactionID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to generate presigned url for transaction receipt"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, struct {
		URL string `json:"url"`
	}{
		URL: presigned,
	})

}

func (s *server) handlePostAccountTransactionReceipt(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()
	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		err := errors.New("itemID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify ownership of item"))
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		err := errors.New("accountID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	_, err = s.account.Account(ctx, itemID, accountID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch account"))
		return
	}

	transactionID := chi.URLParam(r, "transactionID")
	if transactionID == "" {
		err := errors.New("transactionID is required")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	defer closeRequestBody(ctx, r)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to read request body"))
		return
	}

	err = s.transaction.AddReceiptToTransaction(ctx, itemID, transactionID, http.DetectContentType(data), data)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusInternalServerError, errors.New("failed to add file to transaction"))
		return
	}
	s.writeResponse(ctx, w, http.StatusNoContent, nil)

}
