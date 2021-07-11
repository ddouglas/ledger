package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ddouglas/ledger/internal"
)

func (s *server) handlePlaidPostV1Webhook(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) handlePlaidGetLinkToken(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	token, err := s.gateway.LinkToken(ctx, user)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch link token from plaid: %w", err))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})

}
