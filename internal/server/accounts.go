package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

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

	contentLengthStr := r.Header.Get("Content-Length")
	if contentLengthStr == "" {
		err := errors.New("headers missing content length, unable to read file")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		err := errors.New("unable to determine content length from header")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	// ensure the request body gets closed
	defer closeRequestBody(ctx, r)
	// Since the request body is base64 encoded, we first need to decode the base64 encoded string to a byte slice
	buf := new(bytes.Buffer)
	// content length tells us how long the string is in byte, so lets grow the bytes slice to that length
	buf.Grow(contentLength)
	// Now that the memory has been allocated, lets read from the Request Body
	_, _ = buf.ReadFrom(r.Body)
	// Now lets allocate a second buffer. This buffer will hold the base64 decoded data. This should either be an image
	// or a PDF
	buf2 := new(bytes.Buffer)
	buf2.Grow(contentLength)

	// Read the first buffer into the base 64 decoder
	b64Decoder := base64.NewDecoder(base64.StdEncoding, buf)
	// Read from the decoder now into our second buffer
	_, _ = buf2.ReadFrom(b64Decoder)

	err = s.transaction.AddReceiptToTransaction(ctx, itemID, transactionID, buf2)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusInternalServerError, errors.New("failed to add file to transaction"))
		return
	}
	s.writeResponse(ctx, w, http.StatusNoContent, nil)

}

func (s *server) handleDeleteAccountTransactionReceipt(w http.ResponseWriter, r *http.Request) {

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

	err = s.transaction.RemoveReceiptFromTransaction(ctx, itemID, transactionID)
	if err != nil {
		err := errors.New("failed to remove receipt from transaction")
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(ctx, w, http.StatusNoContent, nil)

}
