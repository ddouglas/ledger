package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/go-chi/chi/v5"
)

func (s *server) handleGetUserItems(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	items, err := s.item.ItemsByUserID(ctx, user.ID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch items by user: %w", err))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, items)

}

func (s *server) handlePostUserItems(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var body = new(ledger.RegisterItemRequest)
	defer closeRequestBody(ctx, r)
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch items by user: %w", err))
		return
	}

	item, err := s.item.RegisterItem(ctx, body)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, item)

}

func (s *server) handleDeleteUserItem(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("item id is required to delete an item"))
		return
	}

	err := s.item.DeleteItem(ctx, user.ID, itemID)
	if err != nil {
		s.writeError(ctx, w, http.StatusInternalServerError, fmt.Errorf("failed to delete item: %w", err))
		return
	}

	s.writeResponse(ctx, w, http.StatusNoContent, nil)

}