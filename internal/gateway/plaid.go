package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/null"
)

func (s *service) ClearLinkTokenState(ctx context.Context, state uuid.UUID) {

	s.mux.Lock()
	defer s.mux.Unlock()
	_, ok := s.state[state]
	if !ok {
		return
	}

	delete(s.state, state)

}

func (s *service) LinkTokenByState(ctx context.Context, state uuid.UUID) (*ledger.LinkState, error) {

	s.mux.Lock()
	defer s.mux.Unlock()

	linkState, ok := s.state[state]
	if !ok {
		return nil, errors.New("token with provided state does not exist")
	}

	if linkState.Expiration.Unix() < time.Now().Unix() {
		delete(s.state, state)
		return nil, errors.New("token state has expired")
	}

	return linkState, nil

}

func (s *service) LinkToken(ctx context.Context, user *ledger.User) (*ledger.LinkState, error) {

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service": "gateway",
		"method":  "LinkToken",
		"userID":  user.ID,
	})
	entry.Info("fetch link token")

	linkConfig := plaid.LinkTokenConfigs{
		Products:     s.products,
		CountryCodes: s.countryCodes,
		Language:     s.language,
		Webhook:      s.webhook,
		ClientName:   user.Email,
		User: &plaid.LinkTokenUser{
			ClientUserID: user.ID.String(),
		},
	}

	linkResponse, err := s.client.CreateLinkToken(linkConfig)
	if err != nil {
		entry.WithError(err).Error("failed to fetch link token")
		return nil, err
	}

	linkToken := &ledger.LinkState{
		UserID:     user.ID,
		State:      uuid.Must(uuid.NewV4()),
		Token:      linkResponse.LinkToken,
		Expiration: time.Now().Add(time.Minute * 10),
	}

	s.mux.Lock()
	defer s.mux.Unlock()
	s.state[linkToken.State] = linkToken

	entry.Info("token fetched successfully")
	return linkToken, nil

}

func (s *service) WebhookVerificationKey(ctx context.Context, keyID string) (*plaid.WebhookVerificationKey, error) {
	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service": "gateway",
		"method":  "WebhookVerificationKey",
		"keyID":   keyID,
	})
	entry.Info("fetch webhook verification key")

	response, err := s.client.GetWebhookVerificationKey(keyID)
	if err != nil {
		entry.WithError(err).Error("failed to fetch webhook verification key")
		return nil, fmt.Errorf("failed to fetch webhook verification key: %w", err)
	}

	entry.Info("key fetched successfully")
	return &response.Key, nil

}

func (s *service) Item(ctx context.Context, accessToken string) (*ledger.Item, error) {

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service":            "gateway",
		"method":             "Item",
		"accessTokenTrimmed": accessToken[0:8],
	})
	entry.Info("fetching item for accessToken")

	response, err := s.client.GetItem(accessToken)
	if err != nil {
		entry.WithError(err).Error("failed to fetch item")
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

	entry.Info("item fetched successfully")
	return item, nil

}

func (s *service) ExchangePublicToken(ctx context.Context, publicToken string) (itemID, accessToken string, err error) {

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service":     "gateway",
		"method":      "ExchangePublicToken",
		"publicToken": publicToken,
	})
	entry.Info("exchanging public token")

	response, err := s.client.ExchangePublicToken(publicToken)
	if err != nil {
		entry.WithError(err).Info("failed to exchange public token")
		return "", "", fmt.Errorf("failed to exchange public token: %w", err)
	}

	entry.WithField("itemID", response.ItemID).Info("public token exchanged successfully")
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

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service": "gateway",
		"method":  "Transactions",
		"options": opts,
	})
	entry.Info("fetching transactions")

	response, err := s.client.GetTransactionsWithOptions(accessToken, opts)
	if err != nil {
		entry.WithError(err).Error("failed to fetch transactions")
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	plaidTransactions := append(make([]plaid.Transaction, 0, response.TotalTransactions), response.Transactions...)

	for len(plaidTransactions) < response.TotalTransactions {
		opts.Offset = len(plaidTransactions)
		entry := entry.WithField("options", opts)
		optsResponse, err := s.client.GetTransactionsWithOptions(accessToken, opts)
		if err != nil {
			entry.WithError(err).Error("failed to fetch transactions")
			return nil, fmt.Errorf("failed to fetch transactions with options: %w", err)
		}

		plaidTransactions = append(plaidTransactions, optsResponse.Transactions...)
		entry.WithField("plaidTransactionLength", len(plaidTransactions)).Info()

	}

	entry.WithField("plaidTransactionLength", len(plaidTransactions)).Info("transactions fetched successfully")

	var transactions = make([]*ledger.Transaction, 0, len(plaidTransactions))

	for _, plaidTransaction := range plaidTransactions {
		transaction := new(ledger.Transaction)
		transaction.FromPlaidTransaction(plaidTransaction)
		transaction.ItemID = response.Item.ItemID

		// Plaid returns positives as negatives and vise versa. Here we invert the amount
		// Withdraws/Charges from/to an account are now negative and Deposits are positive.
		transaction.Amount = transaction.Amount * -1

		transactions = append(transactions, transaction)
	}

	return transactions, nil

}

func (s *service) Accounts(ctx context.Context, accessToken string) ([]*ledger.Account, error) {

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service":            "gateway",
		"method":             "Accounts",
		"accessTokenTrimmed": accessToken[0:8],
	})
	entry.Info("fetching accounts for accessToken")

	response, err := s.client.GetAccounts(accessToken)
	if err != nil {
		entry.WithError(err).Error("failed to fetch accounts")
		return nil, fmt.Errorf("failed to fetch accounts from plaid: %w", err)
	}

	var accounts = make([]*ledger.Account, 0, len(response.Accounts))
	for _, plaidAccount := range response.Accounts {
		account := new(ledger.Account)
		account.FromPlaidAccount(response.Item.ItemID, plaidAccount)
		accounts = append(accounts, account)

	}

	entry.Info("accounts fetched successfully")
	return accounts, nil

}
