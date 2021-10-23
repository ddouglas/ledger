package resolvers

import (
	"time"

	"github.com/ddouglas/ledger"
	"github.com/ddouglas/ledger/internal/account"
	"github.com/ddouglas/ledger/internal/gateway"
	"github.com/ddouglas/ledger/internal/item"
	"github.com/ddouglas/ledger/internal/server/gql/dataloaders"
	"github.com/ddouglas/ledger/internal/server/gql/model"
	"github.com/ddouglas/ledger/internal/transaction"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/null"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	logger *logrus.Logger

	account     account.Service
	loaders     dataloaders.Service
	gateway     gateway.Service
	item        item.Service
	transaction transaction.Service
}

func New(
	logger *logrus.Logger,

	account account.Service,
	gateway gateway.Service,
	item item.Service,
	loaders dataloaders.Service,
	transaction transaction.Service,
) *Resolver {
	return &Resolver{
		logger: logger,

		account:     account,
		gateway:     gateway,
		item:        item,
		loaders:     loaders,
		transaction: transaction,
	}
}

func buildTransactionFilters(f *model.TransactionFilter) *ledger.TransactionFilter {
	t := new(ledger.TransactionFilter)
	if f == nil {
		return t
	}
	t.CategoryID = null.StringFromPtr(f.CategoryID)
	t.FromTransactionID = null.StringFromPtr(f.FromTransactionID)
	t.Limit = null.Uint64FromPtr(f.Limit)
	if f.StartDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *f.StartDate)
		if err == nil {
			t.StartDate = null.TimeFrom(parsedDate)
		}
	}
	if f.EndDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *f.EndDate)
		if err == nil {
			t.EndDate = null.TimeFrom(parsedDate)
		}
	}
	if f.OnDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *f.OnDate)
		if err == nil {
			t.OnDate = null.TimeFrom(parsedDate)
		}
	}
	t.DateInclusive = null.BoolFromPtr(f.DateInclusive)
	if f.TransactionType != nil {
		if *f.TransactionType == model.TransactionTypeExpenses {
			t.AmountDir = null.Float64From(-1)
		} else if *f.TransactionType == model.TransactionTypeIncome {
			t.AmountDir = null.Float64From(1)
		}
	}

	return t
}
