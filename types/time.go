package types

import (
	"database/sql/driver"
	"errors"
	"time"
)

// DateText is an implementation of a string for the MySQL type date.
type DateText string

const (
	formtDate = `2006-01-02`
)

// Value implements the driver.Valuer interface,
// and turns the date into a DateText (date) for MySQL storage.
func (d DateText) Value() (driver.Value, error) {
	t, err := time.Parse(formtDate, string(d))
	if err != nil {
		return nil, err
	}
	return DateText(t.Format(formtDate)), nil
}

// Scan implements the sql.Scanner interface,
// and turns the bitfield incoming from MySQL into a Date
func (d *DateText) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	t, err := time.Parse(formtDate, string(v))
	if err != nil {
		return err
	}
	*d = DateText(t.Format(formtDate))
	return nil
}
