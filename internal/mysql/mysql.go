package mysql

import (
	"errors"

	"github.com/ddouglas/ledger"
	"github.com/jmoiron/sqlx"
)

type starter struct {
	*sqlx.DB
}

type transaction struct {
	*sqlx.Tx
}

var ErrInvalidTransaction = errors.New("transaction is invalid, unable to use for query")

func NewTransactioner(db *sqlx.DB) ledger.Starter {
	return &starter{db}
}

func (s *starter) Begin() (ledger.Transactioner, error) {
	txn, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}

	return &transaction{txn}, nil
}
