package ledger

import (
	"fmt"
	"time"
)

var format = "2006-01-02"

type Date time.Time

func (d Date) ToTime() time.Time {
	return time.Time(d)
}

func (d *Date) UnmarshalJSON(b []byte) error {

	p, err := time.Parse(format, string(b))
	if err != nil {
		return fmt.Errorf("failed parse date, must be in format %s", format)
	}

	*d = Date(p)
	return nil

}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(t.Format(format)), nil
}
