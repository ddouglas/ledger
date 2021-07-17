// package item provides service access to account logic and repositories
package item

import (
	"context"
	"fmt"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type Service interface {
	ItemAccountsByUserID(ctx context.Context, userID uuid.UUID, itemID string) ([]*ledger.Account, error)
	RegisterItem(ctx context.Context, request *ledger.RegisterItemRequest) (*ledger.Item, error)
	ledger.ItemRepository
}

func New(optFuncs ...configOption) Service {
	s := &service{}
	for _, optFunc := range optFuncs {
		optFunc(s)
	}
	return s
}

func (s *service) ItemAccountsByUserID(ctx context.Context, userID uuid.UUID, itemID string) ([]*ledger.Account, error) {

	// Ensure Item exists
	item, err := s.ItemByUserID(ctx, userID, itemID)
	if err != nil {
		return nil, errors.Wrap(err, "[ItemAccountsByUserID]")
	}

	accounts, err := s.account.AccountsByItemID(ctx, item.ItemID)
	if err != nil {
		return nil, errors.Wrap(err, "[ItemAccountsByUserID]")
	}

	return accounts, nil

}

func (s *service) ItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*ledger.Item, error) {

	items, err := s.ItemRepository.ItemsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	institutions, err := s.Institutions(ctx)
	if err != nil {
		return nil, err
	}

	mapInstitutions := make(map[string]*ledger.Institution)
	for _, institution := range institutions {
		mapInstitutions[institution.ID] = institution
	}

	for _, item := range items {
		if !item.InstitutionID.Valid {
			continue
		}
		if institution, ok := mapInstitutions[item.InstitutionID.String]; ok {
			item.Institution = institution
		}
	}

	return items, nil

}

func (s *service) RegisterItem(ctx context.Context, request *ledger.RegisterItemRequest) (*ledger.Item, error) {

	user := internal.UserFromContext(ctx)

	accounts, err := s.account.AccountsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	var shouldRegister bool = true
	if len(accounts) > 0 {

		mapRequestAccounts := make(map[string]*ledger.RegisterItemRequestAccount)
		for _, account := range request.Accounts {
			mapRequestAccounts[account.ID] = account
		}

		for _, account := range accounts {
			if knownAccount, ok := mapRequestAccounts[account.AccountID]; ok {
				if knownAccount.Mask == account.Mask.String && account.Name.String == knownAccount.Name {
					shouldRegister = false
					break
				}
			}
		}

	}

	if !shouldRegister {
		return nil, nil
	}

	_, err = s.CreateInstitution(ctx, &ledger.Institution{
		ID:   request.Institution.InstitutionID,
		Name: request.Institution.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create institution: %w", err)
	}

	// Exchange token for item
	_, accessToken, err := s.gateway.ExchangePublicToken(ctx, request.PublicToken)
	if err != nil {
		return nil, err
	}

	item, err := s.gateway.Item(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item for access token: %w", err)
	}

	// Use the Account Token to fetch Accounts
	accounts, err = s.gateway.Accounts(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	item.UserID = user.ID

	// Create Item
	item, err = s.CreateItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to register item: %w", err)
	}

	// Create Accounts
	for _, account := range accounts {
		account.ItemID = item.ItemID
		_, err = s.account.CreateAccount(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("failed to insert account: %w", err)
		}
	}

	return item, nil
}
