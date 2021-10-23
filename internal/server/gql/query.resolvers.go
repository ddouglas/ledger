package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/server/gql/generated"
	"github.com/ddouglas/ledger/internal/server/gql/model"
)

func (r *queryResolver) Categories(ctx context.Context) ([]*ledger.PlaidCategory, error) {
	return r.item.PlaidCategories(ctx)
}

func (r *queryResolver) Items(ctx context.Context) ([]*ledger.Item, error) {
	user := internal.UserFromContext(ctx)

	return r.item.ItemsByUserID(ctx, user.ID)
}

func (r *queryResolver) LinkToken(ctx context.Context) (string, error) {
	user := internal.UserFromContext(ctx)

	token, err := r.gateway.LinkToken(ctx, user)
	if err != nil {
		r.logger.WithError(err).Error("failed to fetch link token from plaid")
		return "", errors.New("failed to fetch link token from plaid")
	}

	return token, nil
}

func (r *queryResolver) TransactionsPaginated(ctx context.Context, itemID string, accountID string, filters *model.TransactionFilter) (*ledger.PaginatedTransactions, error) {
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

func (r *queryResolver) Transactions(ctx context.Context, itemID string, accountID string, filters *model.TransactionFilter) ([]*ledger.Transaction, error) {
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

	transactions, err := r.transaction.TransactionsPaginated(ctx, itemID, accountID, transFilters)
	if err != nil {
		r.logger.WithError(err).Error("failed to fetch transactions")
		return nil, errors.New("failed to fetch transactions")
	}

	return transactions, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *queryResolver) Accounts(ctx context.Context, itemID string) ([]*ledger.Account, error) {
	panic(fmt.Errorf("not implemented"))
}
