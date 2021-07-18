package ledger

import (
	"fmt"
	"time"
)

type TimeFilter struct {
	operation Operation
	time      time.Time
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

type Direction string

const (
	AscDirection  Direction = "ASC"
	DescDirection Direction = "DESC"
)
