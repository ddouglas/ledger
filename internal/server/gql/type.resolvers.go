package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/server/gql/generated"
)

func (r *itemResolver) AvailbleProducts(ctx context.Context, obj *ledger.Item) ([]string, error) {
	return []string(obj.AvailableProducts), nil
}

func (r *itemResolver) BilledProducts(ctx context.Context, obj *ledger.Item) ([]string, error) {
	return []string(obj.BilledProducts), nil
}

func (r *itemResolver) UserID(ctx context.Context, obj *ledger.Item) (string, error) {
	return obj.UserID.String(), nil
}

func (r *itemResolver) Institution(ctx context.Context, obj *ledger.Item) (*ledger.PlaidInstitution, error) {
	if !obj.InstitutionID.Valid {
		return nil, nil
	}

	return r.loaders.InstitutionLoader().Load(ctx, obj.InstitutionID.String)
}

func (r *itemResolver) Accounts(ctx context.Context, obj *ledger.Item) ([]*ledger.Account, error) {
	return r.loaders.AccountsByItemIDLoader().Load(ctx, obj.ItemID)
}

func (r *plaidCategoryResolver) Hierarchy(ctx context.Context, obj *ledger.PlaidCategory) ([]string, error) {
	return []string(obj.Hierarchy), nil
}

func (r *transactionResolver) Category(ctx context.Context, obj *ledger.Transaction) (*ledger.PlaidCategory, error) {
	if !obj.CategoryID.Valid {
		return nil, nil
	}

	return r.loaders.CategoryLoader().Load(ctx, obj.CategoryID.String)
}

// Item returns generated.ItemResolver implementation.
func (r *Resolver) Item() generated.ItemResolver { return &itemResolver{r} }

// PlaidCategory returns generated.PlaidCategoryResolver implementation.
func (r *Resolver) PlaidCategory() generated.PlaidCategoryResolver { return &plaidCategoryResolver{r} }

// Transaction returns generated.TransactionResolver implementation.
func (r *Resolver) Transaction() generated.TransactionResolver { return &transactionResolver{r} }

type itemResolver struct{ *Resolver }
type plaidCategoryResolver struct{ *Resolver }
type transactionResolver struct{ *Resolver }
