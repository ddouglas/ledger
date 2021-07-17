package server

import (
	"net/http"

	"github.com/ddouglas/ledger/internal"
)

func (s *server) handleGetUserAccounts(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	accounts, err := s.account.AccountsByUserID(ctx, user.ID)
	if err != nil {
		s.writeError(ctx, w, http.StatusForbidden, err)
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, accounts)

}
