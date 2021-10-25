package transaction

import (
	"context"

	"github.com/ddouglas/ledger"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

func (s *service) CreateMerchant(ctx context.Context, merchant *ledger.Merchant) (*ledger.Merchant, error) {

	txn, err := s.starter.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction")
	}

	merchant, err = s.MerchantRepository.CreateMerchantTx(ctx, txn, merchant)
	if err != nil {
		_ = txn.Rollback()
		return merchant, errors.Wrap(err, "failed to create merchant")
	}

	_, err = s.MerchantRepository.CreateMerchantAliasTx(ctx, txn, &ledger.MerchantAlias{
		AliasID:    uuid.Must(uuid.NewV4()).String(),
		MerchantID: merchant.ID,
		Alias:      merchant.Name,
	})
	if err != nil {
		_ = txn.Rollback()
		return nil, errors.Wrap(err, "failed to create merchant alias")
	}

	return merchant, txn.Commit()

}

func (s *service) ConvertMerchantToAlias(ctx context.Context, parentMerchantID, childMerchantID string) (*ledger.Merchant, error) {

	childMerchant, err := s.Merchant(ctx, childMerchantID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to locate merchant with id %s", childMerchantID)
	}

	parentMerchant, err := s.Merchant(ctx, parentMerchantID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to locate merchant with id %s", parentMerchantID)
	}

	txn, err := s.starter.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction")
	}

	_, err = s.CreateMerchantAliasTx(ctx, txn, &ledger.MerchantAlias{
		AliasID:    uuid.Must(uuid.NewV4()).String(),
		MerchantID: parentMerchantID,
		Alias:      childMerchant.Name,
	})
	if err != nil {
		_ = txn.Rollback()
		return nil, errors.Wrapf(err, "failed to create an alias of %s", childMerchantID)
	}

	err = s.UpdateTransactionMerchantTx(ctx, txn, childMerchantID, parentMerchantID)
	if err != nil {
		_ = txn.Rollback()
		return nil, errors.Wrapf(err, "failed to update transactions of child merchant to parent merchant")
	}

	err = s.DeleteMerchantTx(ctx, txn, childMerchantID)
	if err != nil {
		_ = txn.Rollback()
		return nil, errors.Wrapf(err, "failed to delete child merchant")
	}

	return parentMerchant, txn.Commit()

}
