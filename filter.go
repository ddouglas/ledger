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
