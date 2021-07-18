package ledger

import (
	"fmt"
	"time"

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

func (f TimeFilter) ToSql(column string) sq.Sqlizer {
	return f.operation.ToSqlizer(column, f.time)
}

type StringFilter struct {
	operation Operation
	str       string
}

func NewStringFilter(op Operation, str string) (StringFilter, error) {
	if !op.IsValid() {
		return StringFilter{}, fmt.Errorf("operation %s is not valid", string(op))
	}

	return StringFilter{
		op, str,
	}, nil

}

func (f StringFilter) ToSql(column string) sq.Sqlizer {
	return f.operation.ToSqlizer(column, f.str)
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

func (f NumberFilter) ToSql(column string) sq.Sqlizer {
	return f.operation.ToSqlizer(column, f.number)
}

type OrderByFilter struct {
	column    string
	direction Direction
}

func NewOrderByFilter(d Direction, c string) (OrderByFilter, error) {
	if !d.IsValid() {
		return OrderByFilter{}, fmt.Errorf("%s is not a valid order by direction", string(d))
	}

	return OrderByFilter{
		c, d,
	}, nil
}

func (o OrderByFilter) ToSql() sq.Sqlizer {

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

var AllDirections = []Direction{AscDirection, DescDirection}

func (d Direction) IsValid() bool {
	for _, dir := range AllDirections {
		if d == dir {
			return true
		}
	}

	return false
}
