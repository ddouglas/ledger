package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type plaidService interface {
	FetchPlaidInstitution(ctx context.Context, id string) (*ledger.PlaidInstitution, error)
	SavePlaidInstitution(ctx context.Context, institution *ledger.PlaidInstitution) error
	FetchPlaidCategory(ctx context.Context, id string) (*ledger.PlaidCategory, error)
	SavePlaidCategory(ctx context.Context, institution *ledger.PlaidCategory) error
}

func plaidInstitutionKey(id string) string {
	return fmt.Sprintf("ledger::plaid::institution::%s", id)
}

func (s *service) FetchPlaidInstitution(ctx context.Context, id string) (*ledger.PlaidInstitution, error) {

	result, err := s.client.Get(ctx, plaidInstitutionKey(id)).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors.Wrapf(err, "[cache.FetchPlaidInstitution] ID: %s", id)
	}

	if err != nil && errors.Is(err, redis.Nil) {
		return nil, nil
	}

	var institution = new(ledger.PlaidInstitution)
	err = json.Unmarshal(result, institution)
	if err != nil {
		return nil, errors.Wrapf(err, "[cache.FetchPlaidInstitution] ID: %s", id)
	}

	return institution, nil

}
func (s *service) SavePlaidInstitution(ctx context.Context, institution *ledger.PlaidInstitution) error {

	data, err := json.Marshal(institution)
	if err != nil {
		return errors.Wrap(err, "[cache.SavePlaidInstitution]")
	}

	_, err = s.client.Set(ctx, plaidInstitutionKey(institution.ID), string(data), time.Hour).Result()

	return errors.Wrap(err, "[cache.SavePlaidInstitution]")

}

func plaidCategoryKey(id string) string {
	return fmt.Sprintf("ledger::plaid::institution::%s", id)
}

func (s *service) FetchPlaidCategory(ctx context.Context, id string) (*ledger.PlaidCategory, error) {

	result, err := s.client.Get(ctx, plaidCategoryKey(id)).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors.Wrapf(err, "[cache.FetchPlaidCategory] ID: %s", id)
	}

	if err != nil && errors.Is(err, redis.Nil) {
		return nil, nil
	}

	var category = new(ledger.PlaidCategory)
	err = json.Unmarshal(result, category)
	if err != nil {
		return nil, errors.Wrapf(err, "[cache.FetchPlaidCategory] ID: %s", id)
	}

	return category, nil

}

func (s *service) SavePlaidCategory(ctx context.Context, cateogry *ledger.PlaidCategory) error {

	data, err := json.Marshal(cateogry)
	if err != nil {
		return errors.Wrap(err, "[cache.SavePlaidCategory]")
	}

	_, err = s.client.Set(ctx, plaidCategoryKey(cateogry.ID), string(data), time.Hour).Result()

	return errors.Wrap(err, "[cache.SavePlaidCategory]")

}
