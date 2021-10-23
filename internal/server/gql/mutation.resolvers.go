package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal"
	"github.com/ddouglas/ledger/internal/server/gql/generated"
)

func (r *mutationResolver) DeleteReceipt(ctx context.Context, itemID string, transactionID string) (bool, error) {
	err := r.transaction.RemoveReceiptFromTransaction(ctx, itemID, transactionID)

	return err == nil, err
}

func (r *mutationResolver) UpdateTransaction(ctx context.Context, itemID string, transactionID string, input *ledger.UpdateTransactionInput) (*ledger.Transaction, error) {
	user := internal.UserFromContext(ctx)

	_, err := r.item.ItemByUserID(ctx, user.ID, itemID)
	if err != nil {
		r.logger.WithError(err).Error("failed to verify ownership")
		return nil, errors.New("failed to verify ownership")
	}

	transaction, err := r.transaction.Transaction(ctx, itemID, transactionID)
	if err != nil {
		r.logger.WithError(err).Error("failed to fetch transaction")
		return nil, errors.New("failed to fetch transaction")
	}

	transaction.FromUpdateTransactionInput(input)

	transaction, err = r.transaction.UpdateTransaction(ctx, transactionID, transaction)
	if err != nil {
		r.logger.WithError(err).Error("failed to update transaction")
		return nil, errors.New("failed to update transaction")
	}

	return transaction, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
