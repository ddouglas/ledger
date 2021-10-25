package ledger

import (
	"context"
	"time"
)

type MerchantRepository interface {
	Merchant(ctx context.Context, id string) (*Merchant, error)
	MerchantByAlias(ctx context.Context, alias string) (*Merchant, error)
	Merchants(ctx context.Context) ([]*Merchant, error)
	CreateMerchant(ctx context.Context, merchant *Merchant) (*Merchant, error)
	CreateMerchantTx(ctx context.Context, tx Transactioner, merchant *Merchant) (*Merchant, error)
	UpdateMerchant(ctx context.Context, id string, merchant *Merchant) (*Merchant, error)
	UpdateMerchantTx(ctx context.Context, tx Transactioner, id string, merchant *Merchant) (*Merchant, error)
	DeleteMerchant(ctx context.Context, id string) error
	DeleteMerchantTx(ctx context.Context, tx Transactioner, id string) error

	MerchantAliasesByMerchantID(ctx context.Context, merchantID string) ([]*MerchantAlias, error)
	CreateMerchantAlias(ctx context.Context, alias *MerchantAlias) (*MerchantAlias, error)
	CreateMerchantAliasTx(ctx context.Context, tx Transactioner, alias *MerchantAlias) (*MerchantAlias, error)
	UpdateMerchantAlias(ctx context.Context, aliasID string, alias *MerchantAlias) (*MerchantAlias, error)
	UpdateMerchantAliasTx(ctx context.Context, tx Transactioner, aliasID string, alias *MerchantAlias) (*MerchantAlias, error)
}

type Merchant struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type MerchantAlias struct {
	AliasID    string    `db:"alias_id" json:"aliasID"`
	MerchantID string    `db:"merchant_id" json:"id"`
	Alias      string    `db:"alias" json:"name"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time `db:"updated_at" json:"updatedAt"`
}
