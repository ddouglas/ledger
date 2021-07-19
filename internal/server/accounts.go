package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/importer"
	"github.com/go-chi/chi/v5"
	"github.com/volatiletech/null"
)

func (s *server) handleGetAccountTransactions(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("itemID is required"))
		return
	}

	_, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify ownership"))
		return
	}

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("accountID is required"))
		return
	}

	_, err = s.account.Account(ctx, itemID, accountID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch account"))
		return
	}

	var filters *ledger.TransactionFilter
	fromTransactionID := r.URL.Query().Get("fromTransactionID")
	count := r.URL.Query().Get("count")
	if fromTransactionID != "" && count != "" {

		parsedCount, err := strconv.ParseUint(count, 10, 64)
		if err != nil {
			s.logger.WithError(err).Error()
			s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to parse value in count query param to valid uint64"))
			return
		}

		filters = &ledger.TransactionFilter{
			FromTransactionID: &ledger.StringFilter{String: fromTransactionID, Operation: ledger.LtOperation},
			Count:             null.Uint64From(parsedCount),
		}
	}

	transactions, err := s.transaction.TransactionsByAccountID(ctx, itemID, accountID, filters)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transactions"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, transactions)

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
