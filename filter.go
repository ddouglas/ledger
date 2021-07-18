package ledger

import (
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
)

type TimeFilter struct {
	operation Operation
	time      time.Time
}

func NewTimeFilter(op Operation, t time.Time) (TimeFilter, error) {
	if !op.IsValid() {
		return TimeFilter{}, fmt.Errorf("operation %s is not valid", string(op))
	}

	return TimeFilter{
		op, t,
	}, nil
}

func (f TimeFilter) ToSqlizer() squirrel.Sqlizer {

}

type StringFilter struct {
	operation Operation
	string    string
}

func NewStringFilter(op Operation, str string) (StringFilter, error) {
	if !op.IsValid() {
		return StringFilter{}, fmt.Errorf("operation %s is not valid", string(op))
	}

	return StringFilter{
		op, str,
	}, nil

}

type NumberFilter struct {
	operation Operation
	number    int64
}

func NewNumberFilter(op Operation, number int64) (NumberFilter, error) {
	if !op.IsValid() {
		return NumberFilter{}, fmt.Errorf("operation %s is not valid", string(op))
	}

	return NumberFilter{
		op, number,
	}, nil
}

type OrderByFilter struct {
	Direction Direction
}

type Operation string

const (
	EqOperation    Operation = "="
	NotEqOperation Operation = "!="
	GtOperation    Operation = ">"
	GtEqOperation  Operation = ">="
	LtOperation    Operation = "<"
	LtEqOperation  Operation = "<="
	// EqOperation Operation = "="
	// EqOperation Operation = "="
)

var AllOps = []Operation{
	EqOperation, NotEqOperation, GtOperation,
	GtEqOperation, LtOperation, LtEqOperation,
}

func (o Operation) IsValid() bool {
	for _, op := range AllOps {
		if op == o {
			return true
		}
	}

	return false
}

func (o Operation) ToSqlizer(column string, value interface{}) sq.Sqlizer {
	switch o {
	case EqOperation:
		return sq.Eq{column: value}
	case NotEqOperation:
		return sq.NotEq{column: value}
	case GtEqOperation:
		return sq.GtOrEq{column: value}
	case GtOperation:
		return sq.Gt{column: value}
	case LtEqOperation:
		return sq.LtOrEq{column: value}
	case LtOperation:
		return sq.Lt{column: value}
	}

	return nil
}

type Direction string

const (
	AscDirection  Direction = "ASC"
	DescDirection Direction = "DESC"
)
