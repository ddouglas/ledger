package gateway

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/pkg/mem"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Accounts(ctx context.Context, accessToken string) ([]*ledger.Account, error)
	ExchangePublicToken(ctx context.Context, publicToken string) (itemID, accessToken string, err error)
	Item(ctx context.Context, accessToken string) (*ledger.Item, error)
	LinkToken(ctx context.Context, user *ledger.User) (string, error)
	Transactions(ctx context.Context, accessToken string, startDate, endDate time.Time, accountIDs []string) ([]*ledger.Transaction, error)
	WebhookVerificationKey(ctx context.Context, keyID string) (*plaid.WebhookVerificationKey, error)

	ImportCategories(ctx context.Context)
	ImportInstitutions(ctx context.Context)
	PlaidCategory(ctx context.Context, id string) (*ledger.PlaidCategory, error)
	PlaidInstitution(ctx context.Context, id string) (*ledger.PlaidInstitution, error)

	ledger.PlaidRepository
}

func New(optFuncs ...configOption) Service {

	s := &service{}
	for _, optFunc := range optFuncs {
		optFunc(s)
	}
	return s
}

func (s *service) PlaidCategory(ctx context.Context, id string) (*ledger.PlaidCategory, error) {

	category, err := s.cache.FetchPlaidCategory(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "[gateway.PlaidCategory]")
	}

	if category != nil {
		return category, nil
	}

	category, err = s.PlaidRepository.PlaidCategory(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "[gateway.PlaidCategory]")
	}

	err = s.cache.SavePlaidCategory(ctx, category)
	if err != nil {
		return nil, errors.Wrap(err, "[gateway.PlaidCategory]")
	}

	return category, nil

}

func (s *service) PlaidInstitution(ctx context.Context, id string) (*ledger.PlaidInstitution, error) {

	category, err := s.cache.FetchPlaidInstitution(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "[gateway.PlaidInstitution]")
	}

	if category != nil {
		return category, nil
	}

	category, err = s.PlaidRepository.PlaidInstitution(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "[gateway.PlaidInstitution]")
	}

	err = s.cache.SavePlaidInstitution(ctx, category)
	if err != nil {
		return nil, errors.Wrap(err, "[gateway.PlaidInstitution]")
	}

	return category, nil

}

func (s *service) ImportCategories(ctx context.Context) {

	txn := s.newrelic.StartTransaction("import-plaid-categories")
	ctx = newrelic.NewContext(ctx, txn)
	defer txn.End()

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service": "gateway",
		"method":  "ImportCategories",
	})
	entry.Info("fetching categories")

	response, err := s.client.GetCategories()
	if err != nil {
		entry.WithError(err).Errorln("failed to fetch categories from plaid")
		return
	}

	entry.WithField("count_categories", len(response.Categories))

	for _, plaidCategory := range response.Categories {

		_, err := s.PlaidRepository.PlaidCategory(ctx, plaidCategory.CategoryID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			entry.WithField("category_id", plaidCategory.CategoryID).WithError(err).Errorln()
			return
		}

		// Only want to create missing categories
		if err != nil && errors.Is(err, sql.ErrNoRows) {

			entry.WithField("category_id", plaidCategory.CategoryID).Info("new category detected, creating record in db")

			category := &ledger.PlaidCategory{
				ID:        plaidCategory.CategoryID,
				Name:      strings.Join(plaidCategory.Hierarchy, " - "),
				Group:     plaidCategory.Group,
				Hierarchy: plaidCategory.Hierarchy,
			}

			_, err = s.PlaidRepository.CreatePlaidCategory(ctx, category)
			if err != nil {
				entry.WithField("category_id", plaidCategory.CategoryID).WithError(err).Errorln()
				return
			}

		}

	}

}

func (s *service) ImportInstitutions(ctx context.Context) {

	var txn = s.newrelic.StartTransaction("import-plaid-institutions")
	ctx = newrelic.NewContext(ctx, txn)
	defer txn.End()

	var count, offset int = 500, 0
	var countryCodes = []string{"US"}

	entry := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"service": "gateway",
		"method":  "ImportInstitutions",
		"offset":  offset, "count": count,
	})
	entry.Info("fetching institutions")

	response, err := s.client.GetInstitutions(count, offset, countryCodes)
	if err != nil {
		entry.WithError(err).Errorln("failed to fetch institutions from plaid")
		return
	}

	plaidInstitutions := append(make([]plaid.Institution, 0, response.Total), response.Institutions...)

	for len(plaidInstitutions) < response.Total {
		offset = len(plaidInstitutions)
		entry = entry.WithFields(logrus.Fields{
			"offset": offset,
		})
		entry.Info("fetching institutions")
		innerResponse, err := s.client.GetInstitutions(count, offset, countryCodes)
		if err != nil {
			entry.WithError(err).Errorln("failed to fetch paginated institutions from plaid")
			return
		}

		plaidInstitutions = append(plaidInstitutions, innerResponse.Institutions...)
		entry.WithField("plaidInstitutionsLength", len(plaidInstitutions))
		mem.PrintMemUsage()
	}

	for _, plaidInstitution := range plaidInstitutions {

		_, err := s.PlaidRepository.PlaidInstitution(ctx, plaidInstitution.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			entry.WithError(err).Errorln()
			return
		}

		if err != nil && errors.Is(err, sql.ErrNoRows) {

			entry.WithField("institution_id", plaidInstitution.ID).Info("new institution detected, creating record in db")

			institution := &ledger.PlaidInstitution{
				ID:   plaidInstitution.ID,
				Name: plaidInstitution.Name,
			}

			_, err = s.PlaidRepository.CreatePlaidInstitution(ctx, institution)
			if err != nil {
				entry.WithError(err).Errorln()
				return
			}

		}
		mem.PrintMemUsage()

	}

}
