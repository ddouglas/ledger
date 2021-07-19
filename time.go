package ledger

import (
	"fmt"
	"time"
)

var format = "2006-01-02"

type Date time.Time

func (d *Date) UnmarshalJSON(b []byte) error {

	p, err := time.Parse(f, string(b))
	if err != nil {
		return fmt.Errorf("failed parse date, must be in format %s", f)
	}

	*d = Date(p)
	return nil

}

func (d Date) MarshalJSON(v interface{}) ([]byte, error) {}
