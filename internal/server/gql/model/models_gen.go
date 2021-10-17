// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type TransactionFilter struct {
	CategoryID        *string          `json:"categoryID"`
	FromTransactionID *string          `json:"fromTransactionID"`
	Limit             *uint64          `json:"limit"`
	StartDate         *string          `json:"startDate"`
	EndDate           *string          `json:"endDate"`
	DateInclusive     *bool            `json:"dateInclusive"`
	OnDate            *string          `json:"onDate"`
	TransactionType   *TransactionType `json:"transactionType"`
}

type TransactionType string

const (
	TransactionTypeExpenses TransactionType = "EXPENSES"
	TransactionTypeIncome   TransactionType = "INCOME"
)

var AllTransactionType = []TransactionType{
	TransactionTypeExpenses,
	TransactionTypeIncome,
}

func (e TransactionType) IsValid() bool {
	switch e {
	case TransactionTypeExpenses, TransactionTypeIncome:
		return true
	}
	return false
}

func (e TransactionType) String() string {
	return string(e)
}

func (e *TransactionType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TransactionType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TransactionType", str)
	}
	return nil
}

func (e TransactionType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
