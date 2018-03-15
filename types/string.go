package types

import (
	"database/sql/driver"
	"errors"
)

// NullString is an implementation of a string for the MySQL type char/varchar ....
type NullString string

// Value implements the driver.Valuer interface,
// and turns the string into a bytes for MySQL storage.
func (s NullString) Value() (driver.Value, error) {
	return []byte(s), nil
}

// Scan implements the sql.Scanner interface,
// and turns the bytes incoming from MySQL into a string
func (s *NullString) Scan(src interface{}) error {
	if src != nil {
		v, ok := src.([]byte)
		if !ok {
			return errors.New("bad []byte type assertion")
		}

		*s = NullString(v)
		return nil
	}

	*s = NullString("")
	return nil
}
