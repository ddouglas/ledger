package ledger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null"
)

type ItemRepository interface {
	Item(ctx context.Context, itemID string) (*Item, error)
	ItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*Item, error)
	CreateItem(ctx context.Context, item *Item) (*Item, error)
	UpdateItem(ctx context.Context, itemID string, item *Item) (*Item, error)
	DeleteItem(ctx context.Context, userID uuid.UUID, itemID string) error
}

type Item struct {
	UserID                     uuid.UUID   `db:"user_id" json:"userID"`
	ItemID                     string      `db:"item_id" json:"itemID"`
	AccessToken                string      `db:"access_token" json:"-"`
	InstitutionID              null.String `db:"institution_id" json:"institutionID"`
	Webhook                    null.String `db:"webhook" json:"webhook"`
	Error                      null.String `db:"error" json:"error"`
	AvailableProducts          SliceString `db:"available_products" json:"availableProducts"`
	BilledProducts             SliceString `db:"billed_products" json:"billedProducts"`
	ConsentExpirationTime      null.Time   `db:"consent_expiration_time" json:"consentExpirationTime"`
	UpdateType                 null.String `db:"update_type" json:"updateType"`
	InvestmentsUpdateStatus    null.String `db:"investments_update_status" json:"investmentsUpdateStatus"`
	InvestmentsUpdateDateTime  null.Time   `db:"investments_update_datetime" json:"investmentsUpdateDateTime"`
	TransactionsUpdateStatus   null.String `db:"transactions_update_status" json:"transactionsUpdateStatus"`
	TransactionsUpdateDateTime null.Time   `db:"transactions_update_datetime" json:"transactionsUpdateDateTime"`
	WebhookStatusDateTime      null.Time   `db:"webhook_status_datetime" json:"webhookStatusDatetime"`
	WebhookStatusCodeSent      null.String `db:"webhook_status_code_sent" json:"webhookStatusCodeSent"`
	CreatedAt                  time.Time   `db:"created_at" json:"-"`
	UpdatedAt                  time.Time   `db:"updated_at" json:"-"`

	Institution *Institution `json:"institution,omitempty"`
}

type RegisterItemRequest struct {
	Institution struct {
		InstitutionID string `json:"institution_id"`
		Name          string `json:"name"`
	} `json:"institution"`
	Accounts    []*RegisterItemRequestAccount `json:"accounts"`
	PublicToken string                        `json:"public_token"`
}

type RegisterItemRequestAccount struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Mask    string `json:"mask"`
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
}

type SliceString []string

func (s SliceString) Value() (driver.Value, error) {

	if len(s) == 0 {
		return []byte(`[]`), nil
	}

	return json.Marshal(s)

}

func (s *SliceString) Scan(value interface{}) error {

	switch data := value.(type) {
	case []byte:
		err := json.Unmarshal(data, s)
		if err != nil {
			return fmt.Errorf("failed to scan string into SliceString: %w", err)
		}
	default:
		return fmt.Errorf("failed to scan value into SliceString: unsupported type %T", value)
	}

	return nil
}
