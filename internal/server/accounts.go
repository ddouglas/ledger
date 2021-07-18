package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/importer"
	"github.com/go-chi/chi/v5"
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

	transactions, err := s.transaction.TransactionsByAccountID(ctx, itemID, accountID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transactions"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, transactions)

}

func (s *server) handleUpdateTransactions(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var message = new(importer.WebhookMessage)
	err := json.NewDecoder(r.Body).Decode(message)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to decode request body"))
		return
	}

	err = s.importer.PublishWebhookMessage(ctx, message)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to process refresh request"))
		return
	}

	s.writeResponse(ctx, w, http.StatusNoContent, nil)
}
