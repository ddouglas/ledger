package importer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/plaid/plaid-go/plaid"
)

type WebhookMessage struct {
	WebhookType     string       `json:"webhook_type"`
	WebhookCode     string       `json:"webhook_code"`
	ItemID          string       `json:"item_id"`
	Error           *plaid.Error `json:"error,omitempty"`
	NewTransactions int          `json:"new_transactions"`
}

func (s *service) PublishWebhookMessage(ctx context.Context, webhook *WebhookMessage) error {
	data, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = s.redis.Publish(ctx, gateway.PubSubPlaidWebhook, data).Result()
	if err != nil {
		return err
	}

	return nil
}
