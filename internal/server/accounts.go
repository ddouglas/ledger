package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/importer"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
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
	fromTransactionID := r.URL.Query().Get("fromTransactionID")
	if fromTransactionID != "" {
		filters.FromTransactionID = null.NewString(fromTransactionID, true)
	}

	limit := r.URL.Query().Get("limit")
	if limit != "" {
		parsedLimit, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			GetLogEntry(r).WithError(err).Error()
			s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to parse value in limit query param to valid uint64"))
			return
		}

		filters.Limit = null.Uint64From(parsedLimit)
	}

	fromDate := r.URL.Query().Get("fromDate")
	if fromDate != "" {

		parsedDate, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			GetLogEntry(r).WithError(err).Error()
			s.writeError(ctx, w, http.StatusBadRequest, errors.Wrap(err, "failed to parse value in fromDate query param to valid time"))
			return
		}

		filters.FromDate = null.NewTime(parsedDate, true)

	}

	var results = new(ledger.PaginatedTransactions)

	transactions, err := s.transaction.TransactionsPaginated(ctx, itemID, accountID, filters)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transactions"))
		return
	}

	results.Transactions = transactions

	count, err := s.transaction.TransactionsCount(ctx, itemID, accountID)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transactions"))
		return
	}

	results.Total = count

	s.writeResponse(ctx, w, http.StatusOK, results)

}

// handleUpdateTransactions handles triggering importer to refresh transactions within a specific date range for a specific account
func (s *server) handleUpdateTransactions(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var message = new(importer.WebhookMessage)
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
	message.Options = &importer.WebhookMessageOptions{
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
