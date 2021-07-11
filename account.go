package ledger

import (
	"time"

	"github.com/volatiletech/null"
)

type AccountRepository interface {
	// Account(ctx context.Context, itemID string, accountID string) (*Account, error)
	// Accounts(ctx context.Context, itemID string) ([]*Account, error)
	// CreateAccount(ctx context.Context, account *Account) (*Account, error)
	// UpdateAccount(ctx context.Context, itemID, accountID string, account *Account) (*Account, error)
	// DeleteAccount(ctx context.Context, itemID, accountID string) error
}

type Account struct {
	AccountID    string          `db:"account_id" json:"accountID"`
	Balance      *AccountBalance `json:"balance"`
	Mask         string          `db:"mask" json:"mask"`
	Name         string          `db:"name" json:"name"`
	OfficialName string          `db:"official_name" json:"officialName"`
	Subtype      string          `db:"subtype" json:"subtype"`
	Type         string          `db:"type" json:"type"`
}

type AccountBalance struct {
	Available              int         `db:"available" json:"available"`
	Current                int         `db:"current" json:"current"`
	IsoCurrencyCode        string      `db:"iso_currency_code" json:"isoCurrencyCode"`
	Limit                  null.String `db:"limit" json:"limit"`
	UnofficialCurrencyCode null.String `db:"unofficial_currency_code" json:"unofficialCurrencyCode"`
	LastUpdated            null.Time   `db:"last_updated_datetime" json:"lastUpdatedDatetime"`
	CreatedAt              time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt              time.Time   `db:"updated_at" json:"updated_at"`
}
