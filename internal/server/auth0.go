package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ddouglas/ledger"
)

func (s *server) handleAuth0PostEmailExchange(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	// Validate Server Token that should've been transmitted with the request
	// token := r.Header.Get("X-Auth0-ServerToken")
	// if token == "" || token != s.auth0ServerToken {
	// 	s.writeError(ctx, w, http.StatusForbidden, fmt.Errorf("missing or invalid token provided"))
	// 	return
	// }

	defer closeRequestBody(ctx, r)
	var user = new(ledger.User)
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		s.writeError(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to decode request body: %w", err))
		return
	}

	user, err = s.user.FetchOrCreateUser(ctx, user)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch user: %w", err))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, user)

}

func (s *server) handleAuth0PostCodeExchange(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	code := r.URL.Query().Get("code")
	if code == "" {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("expected non empty query param code, got empty param"))
		return
	}

	token, _, err := s.auth.ExchangeCode(ctx, code)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch user: %w", err))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})

}
