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
