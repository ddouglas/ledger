package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/go-chi/chi/v5"
)

func (s *server) handlePlaidPostV1Webhook(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var message = new(ledger.WebhookMessage)

	defer closeRequestBody(ctx, r)
	err := json.NewDecoder(r.Body).Decode(message)
	if err != nil {
		s.writeError(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to decode request body: %w", err))
		return
	}

	// publish message to pubsub via importer service
	err = s.importer.PublishWebhookMessage(ctx, message)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to publish message: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) handlePlaidGetV1Category(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	categoryID := chi.URLParam(r, "categoryID")
	if categoryID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("categoryID is required"))
		return
	}

	category, err := s.gateway.PlaidCategory(ctx, categoryID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch category"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, category)

}

func (s *server) handlePlaidGetV1Institution(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	institutionID := chi.URLParam(r, "institutionID")
	if institutionID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("institutionID is required"))
		return
	}

	institution, err := s.gateway.PlaidInstitution(ctx, institutionID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch institution"))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, institution)

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
