package server

import (
	"errors"
	"net/http"

	"github.com/ddouglas/ledger/internal"
	"github.com/go-chi/chi/v5"
)

func (s *server) handleGetAccountTransactions(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("accountID is required"))
		return
	}

	account, err := s.account.Account(ctx, accountID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch account"))
		return
	}

	_, err = s.item.ItemByUserID(ctx, user.ID, account.ItemID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify ownership"))
		return
	}

	transactions, err := s.transaction.TransactionsByAccountID(ctx, account.ItemID, accountID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to fetch transactions"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, transactions)

}
