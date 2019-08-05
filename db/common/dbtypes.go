package common

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

// NullString represents a string that may be null, but more easy to use than sql.NullString.
// just set a value by pointer.
// NullString implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullString struct {
	V *string
}

// Scan implements the Scanner interface.
func (n *NullString) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = ""

		return
	}

	v := sql.NullString{}
	if err := v.Scan(value); err != nil {
		return err
	}

	*n.V = v.String
	return
}

// Value implements the driver Valuer interface.
func (n NullString) Value() (driver.Value, error) {
	if n.V == nil {
		return nil, nil
	}

	return *n.V, nil
}

// NullInt represents a string that may be null, but more easy to use than sql.NullInt64.
// just set a value by pointer.
// NullInt implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullInt struct {
	V *int
}

// Scan implements the Scanner interface.
func (n *NullInt) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = 0

		return
	}

	v := sql.NullInt64{}
	if err := v.Scan(value); err != nil {
		return err
	}

	*n.V = int(v.Int64)
	return
}

// Value implements the driver Valuer interface.
func (n NullInt) Value() (driver.Value, error) {
	if n.V == nil {
		return nil, nil
	}

	return int64(*n.V), nil
}

// NullInt64 represents a string that may be null, but more easy to use than sql.NullInt64.
// just set a value by pointer.
// NullInt64 implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullInt64 struct {
	V *int64
}

// Scan implements the Scanner interface.
func (n *NullInt64) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = 0

		return
	}

	v := sql.NullInt64{}
	if err := v.Scan(value); err != nil {
		return err
	}

	*n.V = v.Int64
	return
}

// Value implements the driver Valuer interface.
func (n NullInt64) Value() (driver.Value, error) {
	if n.V == nil {
		return nil, nil
	}

	return *n.V, nil
}

// NullTime represents a time.Time that may be null.
// NullTime implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullTime struct {
	V *time.Time
}

// Scan implements the Scanner interface.
func (n *NullTime) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = time.Time{}

		return
	}

	var ok bool
	*n.V, ok = value.(time.Time)
	if !ok {
		return errors.New("value isn't a time.Time")
	}
	return
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if n.V == nil {
		return nil, nil
	}

	return *n.V, nil
}

// NullFloat64 represents a float64 that may be null.
// NullFloat64 implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullFloat64 struct {
	V *float64
}

// Scan implements the Scanner interface.
func (n *NullFloat64) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = 0.

		return
	}

	v := sql.NullFloat64{}
	if err := v.Scan(value); err != nil {
		return err
	}

	*n.V = v.Float64
	return
}

// Value implements the driver Valuer interface.
func (n NullFloat64) Value() (driver.Value, error) {
	if n.V == nil {
		return nil, nil
	}

	return *n.V, nil
}

// NullBool represents a bool that may be null.
// NullBool implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullBool struct {
	V *bool
}

// Scan implements the Scanner interface.
func (n *NullBool) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = false

		return
	}

	v := sql.NullBool{}
	if err := v.Scan(value); err != nil {
		return err
	}

	*n.V = v.Bool
	return
}

// Value implements the driver Valuer interface.
func (n NullBool) Value() (driver.Value, error) {
	if n.V == nil {
		return nil, nil
	}

	return *n.V, nil
}

// NullDecimal represents a bool that may be null.
// NullDecimal implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullDecimal struct {
	V *decimal.Decimal
}

// Scan implements the Scanner interface.
func (n *NullDecimal) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = decimal.Zero

		return
	}

	v := decimal.NullDecimal{}
	if err := v.Scan(value); err != nil {
		return err
	}

	*n.V = v.Decimal
	return
}

// Value implements the driver Valuer interface.
func (n NullDecimal) Value() (driver.Value, error) {
	if n.V == nil {
		return nil, nil
	}

	return n.V.String(), nil
}

// NullUUID represents a uuid.UUID that may be null.
// NullUUID implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullUUID struct {
	V *uuid.UUID
}

func (n *NullUUID) Reset() {
	n.V = nil
}

// Scan implements the Scanner interface.
func (n *NullUUID) Scan(value interface{}) (err error) {
	if n.V == nil {
		return
	}

	if value == nil {
		*n.V = uuid.Nil

		return
	}

	*n.V = uuid.FromBytesOrNil(value.([]byte))
	return
}

// Value implements the driver Valuer interface.
func (n NullUUID) Value() (driver.Value, error) {
	if n.V == nil || uuid.Equal(*n.V, uuid.Nil) {
		return nil, nil
	}

	return (*n.V).Bytes(), nil
}
