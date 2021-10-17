package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/server/gql/generated"
	"github.com/ddouglas/ledger/internal/server/gql/model"
)

func (r *queryResolver) Items(ctx context.Context) ([]*ledger.Item, error) {
	user := internal.UserFromContext(ctx)

	return r.item.ItemsByUserID(ctx, user.ID)
}

func (r *queryResolver) Transactions(ctx context.Context, itemID string, accountID string, filters *model.TransactionFilter) (*ledger.PaginatedTransactions, error) {
	user := internal.UserFromContext(ctx)

	_, err := r.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		r.logger.WithError(err).Error("failed to verify ownership")
		return nil, errors.New("failed to verify ownership")
	}

	_, err = r.account.Account(ctx, itemID, accountID)
	if err != nil {
		r.logger.WithError(err).Error("failed to fetch account")
		return nil, errors.New("failed to fetch account")
	}

	transFilters := buildTransactionFilters(filters)

	var results = new(ledger.PaginatedTransactions)
	results.Transactions, err = r.transaction.TransactionsPaginated(ctx, itemID, accountID, transFilters)
	if err != nil {
		r.logger.WithError(err).Error("failed to fetch transactions")
		return nil, errors.New("failed to fetch transactions")
	}

	results.Total, err = r.transaction.TransactionsCount(ctx, itemID, accountID, transFilters)
	if err != nil {
		r.logger.WithError(err).Error("failed to fetch transaction count")
		return nil, errors.New("failed to fetch transaction count")
	}

	return results, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
