package dataloaders

import (
	"context"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/item"
	"github.com/ddouglas/ledger/internal/server/gql/dataloaders/generated"
	"github.com/ddouglas/ledger/internal/transaction"
)

type Service interface {
	AccountsByItemIDLoader() *generated.AccountsByItemIDLoader
	CategoryLoader() *generated.CategoryLoader
	InstitutionLoader() *generated.InstitutionLoader
	MerchantLoader() *generated.MerchantLoader
	MerchantAliasLoader() *generated.MerchantAliasLoader
}

type service struct {
	wait        time.Duration
	batch       int
	item        item.Service
	transaction transaction.Service
}

func New(item item.Service, transaction transaction.Service) Service {
	return &service{
		wait:        time.Duration(time.Millisecond * 100),
		batch:       100,
		item:        item,
		transaction: transaction,
	}
}

func (s *service) InstitutionLoader() *generated.InstitutionLoader {
	return generated.NewInstitutionLoader(generated.InstitutionLoaderConfig{
		MaxBatch: s.batch,
		Wait:     s.wait,
		Fetch: func(ctx context.Context, keys []string) ([]*ledger.PlaidInstitution, []error) {

			var errors = make([]error, 0)
			var results = make([]*ledger.PlaidInstitution, len(keys))

			for i, k := range keys {
				record, err := s.item.PlaidInstitution(ctx, k)
				if err != nil {
					errors = append(errors, err)
					return nil, errors
				}

				results[i] = record
			}

			return results, nil

		},
	})
}

func (s *service) AccountsByItemIDLoader() *generated.AccountsByItemIDLoader {
	return generated.NewAccountsByItemIDLoader(generated.AccountsByItemIDLoaderConfig{
		MaxBatch: s.batch,
		Wait:     s.wait,
		Fetch: func(ctx context.Context, keys []string) ([][]*ledger.Account, []error) {
			var errors = make([]error, 0)
			var results = make([][]*ledger.Account, len(keys))
			var user = internal.UserFromContext(ctx)

			for i, k := range keys {
				records, err := s.item.ItemAccountsByUserID(ctx, user.ID, k)
				if err != nil {
					errors = append(errors, err)
					return nil, errors
				}

				results[i] = records
			}

			return results, nil
		},
	})
}

func (s *service) CategoryLoader() *generated.CategoryLoader {
	return generated.NewCategoryLoader(generated.CategoryLoaderConfig{
		MaxBatch: s.batch,
		Wait:     s.wait,
		Fetch: func(ctx context.Context, keys []string) ([]*ledger.PlaidCategory, []error) {
			var errors = make([]error, 0)
			var results = make([]*ledger.PlaidCategory, len(keys))

			for i, k := range keys {
				record, err := s.item.PlaidCategory(ctx, k)
				if err != nil {
					errors = append(errors, err)
					return nil, errors
				}

				results[i] = record
			}

			return results, nil
		},
	})
}

func (s *service) MerchantLoader() *generated.MerchantLoader {
	return generated.NewMerchantLoader(generated.MerchantLoaderConfig{
		MaxBatch: s.batch,
		Wait:     s.wait,
		Fetch: func(ctx context.Context, keys []string) ([]*ledger.Merchant, []error) {
			var errors = make([]error, 0)
			var results = make([]*ledger.Merchant, len(keys))

			for i, k := range keys {
				record, err := s.transaction.Merchant(ctx, k)
				if err != nil {
					errors = append(errors, err)
					return nil, errors
				}

				results[i] = record
			}

			return results, nil
		},
	})
}

func (s *service) MerchantAliasLoader() *generated.MerchantAliasLoader {
	return generated.NewMerchantAliasLoader(generated.MerchantAliasLoaderConfig{
		MaxBatch: s.batch,
		Wait:     s.wait,
		Fetch: func(ctx context.Context, keys []string) ([][]*ledger.MerchantAlias, []error) {
			var errors = make([]error, 0)
			var results = make([][]*ledger.MerchantAlias, len(keys))

			for i, k := range keys {
				records, err := s.transaction.MerchantAliasesByMerchantID(ctx, k)
				if err != nil {
					errors = append(errors, err)
					return nil, errors
				}
				results[i] = records
			}

			return results, nil
		},
	})
}
