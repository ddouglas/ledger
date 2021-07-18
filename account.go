package ledger

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/plaid/plaid-go/plaid"
	"github.com/volatiletech/null"
)

type AccountRepository interface {
	Account(ctx context.Context, itemID string, accountID string) (*Account, error)
	Accounts(ctx context.Context, itemID string) ([]*Account, error)
	AccountsByItemID(ctx context.Context, itemID string) ([]*Account, error)
	AccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error)
	CreateAccount(ctx context.Context, account *Account) (*Account, error)
	UpdateAccount(ctx context.Context, itemID, accountID string, account *Account) (*Account, error)
	DeleteAccount(ctx context.Context, itemID, accountID string) error
}

type Account struct {
	ItemID       string          `db:"item_id" json:"itemID"`
	AccountID    string          `db:"account_id" json:"accountID"`
	Balance      *AccountBalance `json:"balance"`
	Mask         null.String     `db:"mask" json:"mask"`
	Name         null.String     `db:"name" json:"name"`
	OfficialName null.String     `db:"official_name" json:"officialName"`
	Subtype      null.String     `db:"subtype" json:"subtype"`
	Type         null.String     `db:"type" json:"type"`
	CreatedAt    time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time       `db:"updated_at" json:"updated_at"`
}

func (a *Account) FromPlaidAccount(itemID string, account plaid.Account) {

	*a = Account{
		ItemID:       itemID,
		AccountID:    account.AccountID,
		Mask:         null.NewString(account.Mask, account.Mask != ""),
		Name:         null.NewString(account.Name, account.Name != ""),
		OfficialName: null.NewString(account.OfficialName, account.OfficialName != ""),
		Subtype:      null.NewString(account.Subtype, account.Subtype != ""),
		Type:         null.NewString(account.Type, account.Type != ""),
		Balance: &AccountBalance{
			Available:              account.Balances.Available,
			Current:                account.Balances.Current,
			Limit:                  account.Balances.Limit,
			ISOCurrencyCode:        account.Balances.ISOCurrencyCode,
			UnofficialCurrencyCode: null.NewString(account.Balances.UnofficialCurrencyCode, account.Balances.UnofficialCurrencyCode != ""),
		},
	}

}

type AccountBalance struct {
	Available              float64     `db:"available" json:"available"`
	Current                float64     `db:"current" json:"current"`
	Limit                  float64     `db:"limit" json:"limit"`
	ISOCurrencyCode        string      `db:"iso_currency_code" json:"isoCurrencyCode"`
	UnofficialCurrencyCode null.String `db:"unofficial_currency_code" json:"unofficialCurrencyCode"`
	LastUpdated            null.Time   `db:"last_updated_datetime" json:"lastUpdatedDatetime"`
}
