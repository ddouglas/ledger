package server

import (
	"fmt"
	"net/http"
)

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
