package importer

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
)

type WebhookMessage struct {
	WebhookType     string       `json:"webhook_type"`
	WebhookCode     string       `json:"webhook_code"`
	ItemID          string       `json:"item_id"`
	Error           *plaid.Error `json:"error,omitempty"`
	NewTransactions int          `json:"new_transactions"`
	// Custom Fields
	StartDate time.Time                     `json:"startDate,omitempty"`
	EndDate   time.Time                     `json:"endDate,omitempty"`
	Options   *ledger.WebhookMessageOptions `json:"options,omitempty"`
}

type WebhookMessageOptions struct {
	AccountIDs []string `json:"accountIDs,omitempty"`
}

func (s *service) PublishWebhookMessage(ctx context.Context, webhook *ledger.WebhookMessage) error {
	// validate that the item this webhook is for exists
	_, err := s.item.Item(ctx, webhook.ItemID)
	if err != nil {
		s.logger.WithField("item_id", webhook.ItemID).WithError(err).Error()
		return errors.Wrapf(err, "[importer.PublishWebhookMessage] unable to locate item with provided item id: %s", webhook.ItemID)
	}

	data, err := json.Marshal(webhook)
	if err != nil {
		return errors.Wrap(err, "[importer.PublishWebhookMessage] failed to marshal message")
	}

	err = s.WebhookRepository.LogWebhook(ctx, webhook)
	if err != nil {
		return errors.Wrap(err, "[importer.PublishWebhookMessage]")
	}

	_, err = s.redis.RPush(ctx, gateway.PubSubPlaidWebhook, data).Result()
	if err != nil {
		return errors.Wrap(err, "[importer.PublishWebhookMessage]")
	}

	return nil
}

// PublishCustomWebhookMessage
func (s *service) PublishCustomWebhookMessage(ctx context.Context, webhook *ledger.WebhookMessage) error {

	if webhook.StartDate.IsZero() || webhook.EndDate.IsZero() {
		return errors.New("startDate and endDate are required")
	}

	if webhook.StartDate.Unix() > webhook.EndDate.Unix() {
		return errors.New("startDate must be earlier than endDate")
	}

	// Verify ItemID this is for exists
	item, err := s.item.Item(ctx, webhook.ItemID)
	if err != nil {
		s.logger.WithField("item_id", webhook.ItemID).WithError(err).Error()
		return fmt.Errorf("failed to verify item %s exists", webhook.ItemID)
	}

	startDateMin := webhook.EndDate.AddDate(0, 0, -1)
	if webhook.StartDate.Unix() > startDateMin.Unix() {
		return errors.New("startDate and endDate must be at least 24 hours apart")
	}

	startDateMax := webhook.EndDate.AddDate(-1, 0, 0)
	if webhook.StartDate.Unix() < startDateMax.Unix() {
		return errors.New("startDate and endDate cannot be more than 12 months apart")
	}

	twoYearsAgo := time.Now().AddDate(-2, 0, 0)
	if webhook.StartDate.Unix() < twoYearsAgo.Unix() {
		return fmt.Errorf("startDate cannot be earlier than %s", twoYearsAgo.Format("2006-01-02 15:04:05"))
	}

	webhook.WebhookType = "TRANSACTIONS"
	webhook.WebhookCode = "CUSTOM_UPDATE"

	err = s.PublishWebhookMessage(ctx, webhook)
	if err != nil {
		return fmt.Errorf("failed to publish webhook message to importer: %w", err)
	}

	item.IsRefreshing = true

	_, err = s.item.UpdateItem(ctx, item.ItemID, item)

	return errors.Wrap(err, "failed to update item")
}

func (s *service) VerifyWebhookMessage(ctx context.Context, header http.Header, message []byte) error {

	verificationJWT := header.Get("Plaid-Verification")
	if verificationJWT == "" {
		return fmt.Errorf("failed to retrieve plaid verification header from request headers")
	}

	parsed, err := jws.Parse(message)
	if err != nil {
		return fmt.Errorf("failed to parse verification header: %w", err)
	}

	messageSignatures := parsed.Signatures()

	// I can't remember what to do after this. I remember checking for length of the signatures
	// Singature returns the JWT Headers, that will allow me to get the KID
	// Once I get I KID, I can reach out ot Plaid to fethc the key,
	// use the key to verify the signature
	// Once signature is verified I know that the hash in the body of the JWT is safe to use
	// SHA256 Hash the message parameter of this function and compare it with the SHA256 in
	// the body of the JWT. If they equal return nil
	if len(messageSignatures) > 1 {
		return fmt.Errorf("expected a single signature, got %d", len(messageSignatures))
	}

	signature := messageSignatures[0]
	protectedHeaders := signature.ProtectedHeaders()
	if protectedHeaders.Algorithm() != "ES256" {
		return fmt.Errorf("expected algo of ES256, got %s", protectedHeaders.Algorithm())
	}

	keyID := protectedHeaders.KeyID()
	if keyID == "" {
		return fmt.Errorf("expected non empty keyID")
	}

	verificationKey, err := s.gateway.WebhookVerificationKey(ctx, keyID)
	if err != nil {
		return err
	}

	verificationKeyBytes, err := json.Marshal(verificationKey)
	if err != nil {
		return fmt.Errorf("failed to marshal key to be parsed by jwk lib: %w", err)
	}

	key, err := jwk.ParseKey(verificationKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse verification key to jwk: %w", err)
	}

	set := jwk.NewSet()
	set.Add(key)

	parsedVerificationHeader, err := jwt.ParseString(verificationJWT, jwt.WithKeySet(set))
	if err != nil {
		return fmt.Errorf("failed to parse verification header: %w", err)
	}

	verificationHeaderClaims := parsedVerificationHeader.PrivateClaims()
	requestBodyHash := verificationHeaderClaims["request_body_sha256"]

	messageHash := sha256.Sum256(message)
	if messageHash != requestBodyHash {
		return fmt.Errorf("webhook cannot be verified. hashes are not equal")
	}

	return nil

}
