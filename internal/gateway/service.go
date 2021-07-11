package gateway

import (
	"context"

	"github.com/ddouglas/ledger"
	"github.com/plaid/plaid-go/plaid"
)

type Service interface {
	LinkToken(ctx context.Context, user *ledger.User) (string, error)
}

type service struct {
	client *plaid.Client
	config *config
}

func New(client *plaid.Client, optFuncs ...configOption) Service {

	config := &config{}
	for _, optFunc := range optFuncs {
		optFunc(config)
	}

	return &service{
		client: client,
		config: config,
	}
}

func (s *service) LinkToken(ctx context.Context, user *ledger.User) (string, error) {

	linkConfig := plaid.LinkTokenConfigs{}
	if len(s.config.products) > 0 {
		linkConfig.Products = s.config.products
	}
	if len(s.config.countryCodes) > 0 {
		linkConfig.CountryCodes = s.config.countryCodes
	}
	if s.config.language != nil {
		linkConfig.Language = *s.config.language
	}
	if s.config.webhook != nil {
		linkConfig.Webhook = *s.config.webhook
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
