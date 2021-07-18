package ledger

import "time"

type TimeFilter struct {
	Operation Operation
	Time      time.Time
}

type StringFilter struct {
	Operation Operation
	String    string
}

type NumberFilter struct {
	Operation Operation
	Number    int
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

type Direction string

const (
	AscDirection  Direction = "ASC"
	DescDirection Direction = "DESC"
)
