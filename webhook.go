package ledger

import (
	"context"
	"time"

	"github.com/plaid/plaid-go/plaid"
)

type WebhookRepository interface {
	LogWebhook(ctx context.Context, webhook *WebhookMessage) error
}

type WebhookMessage struct {
	WebhookType         string       `json:"webhook_type"`
	WebhookCode         string       `json:"webhook_code"`
	ItemID              string       `json:"item_id"`
	Error               *plaid.Error `json:"error,omitempty"`
	NewTransactions     int          `json:"new_transactions"`
	RemovedTransactions []string     `json:"removed_transactions,omitempty"`
	// Custom Fields
	StartDate time.Time              `json:"startDate,omitempty"`
	EndDate   time.Time              `json:"endDate,omitempty"`
	Options   *WebhookMessageOptions `json:"options,omitempty"`
}

type WebhookMessageOptions struct {
	AccountIDs []string `json:"accountIDs,omitempty"`
}

type WebhookCode string

const (
	CodeInitialUpdate       WebhookCode = "INITIAL_UPDATE"
	CodeHistoricalUpdate    WebhookCode = "HISTORICAL_UPDATE"
	CodeDefaultUpdate       WebhookCode = "DEFAULT_UPDATE"
	CodeCustomUpdate        WebhookCode = "CUSTOM_UPDATE"
	CodeTransactionsRemoved WebhookCode = "TRANSACTIONS_REMOVED"
)

var AllWebhookCodes = []WebhookCode{
	CodeCustomUpdate, CodeDefaultUpdate, CodeHistoricalUpdate,
	CodeInitialUpdate, CodeTransactionsRemoved,
}

func (c WebhookCode) IsValid() bool {
	for _, code := range AllWebhookCodes {
		if c == code {
			return true
		}
	}

	return false
}

var EmailAllowedCode = []WebhookCode{
	CodeDefaultUpdate,
}

func (c WebhookCode) SendEmail() bool {
	for _, code := range EmailAllowedCode {
		if c == code {
			return true
		}
	}

	return false
}
