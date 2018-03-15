package types

import (
	"database/sql/driver"
	"errors"
)

// NullInt is an implementation of a int for the MySQL type int/tinyint ....
type NullInt int

// Value implements the driver.Valuer interface,
// and turns the bytes into a integer for MySQL storage.
func (n NullInt) Value() (driver.Value, error) {
	return NullInt(n), nil
}

// Scan implements the sql.Scanner interface,
// and turns the bytes incoming from MySQL into a integer
func (n *NullInt) Scan(src interface{}) error {
	if src != nil {
		v, ok := src.(int64)
		if !ok {
			return errors.New("bad []byte type assertion")
		}

		*n = NullInt(v)
		return nil
	}
	*n = NullInt(0)
	return nil
}

// NullFloat is an implementation of a int for the MySQL type numeric ....
type NullFloat float64

// Scan implements the sql.Scanner interface,
// and turns the bytes incoming from MySQL into a numeric
func (n *NullFloat) Scan(src interface{}) error {
	if src != nil {
		v, ok := src.(int64)
		if !ok {
			return errors.New("bad []byte type assertion")
		}

		*n = NullFloat(v)
		return nil
	}
	*n = NullFloat(0)
	return nil
}
