package gateway

import (
	"context"
	"fmt"

	"github.com/ddouglas/ledger"
	"github.com/plaid/plaid-go/plaid"
)

type Service interface {
	LinkToken(ctx context.Context, user *ledger.User) (string, error)
	WebhookVerificationKey(ctx context.Context, keyID string) (*plaid.WebhookVerificationKey, error)
}

type service struct {
	client       *plaid.Client
	products     []string
	language     *string
	webhook      *string
	countryCodes []string
}

func New(optFuncs ...configOption) Service {

	s := &service{}
	for _, optFunc := range optFuncs {
		optFunc(s)
	}
	return s
}

func (s *service) LinkToken(ctx context.Context, user *ledger.User) (string, error) {

	linkConfig := plaid.LinkTokenConfigs{}
	if len(s.products) > 0 {
		linkConfig.Products = s.products
	}
	if len(s.countryCodes) > 0 {
		linkConfig.CountryCodes = s.countryCodes
	}
	if s.language != nil {
		linkConfig.Language = *s.language
	}
	if s.webhook != nil {
		linkConfig.Webhook = *s.webhook
	}

	linkConfig.ClientName = user.Name

	linkConfig.User = &plaid.LinkTokenUser{
		ClientUserID: user.ID.String(),
	}

	linkResponse, err := s.client.CreateLinkToken(linkConfig)
	if err != nil {
		return "", err
	}

	return linkResponse.LinkToken, nil

}

func (s *service) WebhookVerificationKey(ctx context.Context, keyID string) (*plaid.WebhookVerificationKey, error) {

	response, err := s.client.GetWebhookVerificationKey(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch webhook verification key: %w", err)
	}

	return &response.Key, nil

}

func (s *service) Item(ctx context.Context, accessToken string) {

}

func (s *service) ExchangePublicToken(ctx context.Context, publicToken string) (itemID, accessToken string, err error) {

	response, err := s.client.ExchangePublicToken(publicToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to exchange public token: %w", err)
	}

	return response.ItemID, response.AccessToken, nil

}
