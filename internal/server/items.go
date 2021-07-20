package server

import (
	"encoding/json"
	"errors"
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

func (s *server) handleGetUserItem(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("itemID is required"))
		return
	}

	item, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, fmt.Errorf("failed to fetch items by user: %w", err))
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, item)

}

func (s *server) handleGetItemAccounts(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("itemID is required"))
		return
	}

	accounts, err := s.item.ItemAccountsByUserID(ctx, user.ID, itemID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, accounts)

}

func (s *server) handleGetItemAccount(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	user := internal.UserFromContext(ctx)

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

	_, err := s.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, errors.New("failed to verify item ownership"))
		return
	}

	account, err := s.account.Account(ctx, itemID, accountID)
	if err != nil {
		s.writeError(ctx, w, http.StatusBadRequest, err)
		return
	}

	s.writeResponse(ctx, w, http.StatusOK, account)

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
