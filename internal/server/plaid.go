package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ddouglas/ledger"
)

func (s *server) handlePlaidPostV1Webhook(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var message = new(ledger.WebhookMessage)

	defer closeRequestBody(ctx, r)
	err := json.NewDecoder(r.Body).Decode(message)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to decode request body: %w", err))
		return
	}

	// publish message to pubsub via importer service
	err = s.importer.PublishWebhookMessage(ctx, message)
	if err != nil {
		GetLogEntry(r).WithError(err).Error()
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to publish message: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) handlePlaidPostLinkToken(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var body = new(ledger.RegisterItemRequest)
	defer closeRequestBody(ctx, r)
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to decode request body: %w", err))
		return
	}

	item, err := s.item.RegisterItem(ctx, body)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, item)

}
