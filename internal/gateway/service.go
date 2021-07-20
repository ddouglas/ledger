package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/plaid/plaid-go/plaid"
	"github.com/volatiletech/null"
)

type Service interface {
	Accounts(ctx context.Context, accessToken string) ([]*ledger.Account, error)
	ExchangePublicToken(ctx context.Context, publicToken string) (itemID, accessToken string, err error)
	Item(ctx context.Context, accessToken string) (*ledger.Item, error)
	LinkToken(ctx context.Context, user *ledger.User) (string, error)
	Transactions(ctx context.Context, accessToken string, startDate, endDate time.Time, accountIDs []string) ([]*ledger.Transaction, error)
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

func (s *service) Item(ctx context.Context, accessToken string) (*ledger.Item, error) {

	response, err := s.client.GetItem(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item for provided access token: %w", err)
	}

	plaidItem := response.Item
	item := &ledger.Item{
		ItemID:                plaidItem.ItemID,
		AccessToken:           accessToken,
		InstitutionID:         null.StringFromPtr(&plaidItem.InstitutionID),
		Webhook:               null.StringFromPtr(&plaidItem.Webhook),
		Error:                 null.NewString(plaidItem.Error.Error(), plaidItem.Error.ErrorCode != ""),
		AvailableProducts:     ledger.SliceString(plaidItem.AvailableProducts),
		BilledProducts:        ledger.SliceString(plaidItem.BilledProducts),
		ConsentExpirationTime: null.NewTime(plaidItem.ConsentExpirationTime, !plaidItem.ConsentExpirationTime.IsZero()),
		ItemStatus:            ledger.ItemStatus(response.Status),
	}

	return item, nil

}

func (s *service) ExchangePublicToken(ctx context.Context, publicToken string) (itemID, accessToken string, err error) {

	response, err := s.client.ExchangePublicToken(publicToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to exchange public token: %w", err)
	}

	return response.ItemID, response.AccessToken, nil

}

func (s *service) Transactions(ctx context.Context, accessToken string, startDate, endDate time.Time, accountIDs []string) ([]*ledger.Transaction, error) {

	opts := plaid.GetTransactionsOptions{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Count:     100,
	}

	if len(accountIDs) > 0 {
		opts.AccountIDs = accountIDs
	}

	response, err := s.client.GetTransactionsWithOptions(accessToken, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	plaidTransactions := make([]plaid.Transaction, 0, response.TotalTransactions)
	plaidTransactions = append(plaidTransactions, response.Transactions...)

	for len(plaidTransactions) < response.TotalTransactions {
		opts.Offset = len(plaidTransactions)
		optsResponse, err := s.client.GetTransactionsWithOptions(accessToken, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch transactions with options: %w", err)
		}

		plaidTransactions = append(plaidTransactions, optsResponse.Transactions...)
		time.Sleep(time.Second)

	}

	var transactions = make([]*ledger.Transaction, 0, len(plaidTransactions))

	for _, plaidTransaction := range plaidTransactions {
		transaction := new(ledger.Transaction)
		transaction.FromPlaidTransaction(plaidTransaction)
		transaction.ItemID = response.Item.ItemID

		// Plaid returns positives as negatives and vise versa. Here we invest it
		// Withdraws from an account are now negative and Deposits are positive.
		transaction.Amount = transaction.Amount * -1

		transactions = append(transactions, transaction)
	}

	return transactions, nil

}

func (s *service) Accounts(ctx context.Context, accessToken string) ([]*ledger.Account, error) {

	response, err := s.client.GetAccounts(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts from plaid: %w", err)
	}

	var accounts = make([]*ledger.Account, 0, len(response.Accounts))
	for _, plaidAccount := range response.Accounts {
		account := new(ledger.Account)
		account.FromPlaidAccount(response.Item.ItemID, plaidAccount)
		accounts = append(accounts, account)

	}

	return accounts, nil

}
