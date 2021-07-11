package ledger

import "github.com/plaid/plaid-go/plaid"

type ItemRepository interface{}

type Item struct {
	ItemID                string      `json:"item_id"`
	InstitutionID         string      `json:"institution_id"`
	AvailableProducts     []string    `json:"available_products"`
	BilledProducts        []string    `json:"billed_products"`
	ConsentExpirationTime string      `json:"consent_expiration_time"`
	Error                 plaid.Error `json:"error"`
	UpdateType            string      `json:"update_type"`
	Webhook               string      `json:"webhook"`
}
