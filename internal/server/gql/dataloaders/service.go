package dataloaders

import (
	"context"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/item"
	"github.com/ddouglas/ledger/internal/server/gql/dataloaders/generated"
)

type Service interface {
	AccountsByItemIDLoader() *generated.AccountsByItemIDLoader
	CategoryLoader() *generated.CategoryLoader
	InstitutionLoader() *generated.InstitutionLoader
}

type service struct {
	wait  time.Duration
	batch int
	item  item.Service
}

func New(item item.Service) Service {
	return &service{
		wait:  time.Duration(time.Millisecond * 100),
		batch: 100,
		item:  item,
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
