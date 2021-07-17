package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
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
	// Once signature is verified I know that the has in the body of the JWT is safe to use
	// SHA256 Hash the message parameter of this function and compare it with the SHA256 in
	// the body of the JWT. If they equal return nil
	if len(messageSignatures) > 0 {
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

	parsedVerificationHeader := jwt.ParseString(verificationJWT, jwt.WithKeySet(set))

	return nil

}
