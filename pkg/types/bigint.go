package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

// BigInt prevents precision loss in frontend JavaScript
type BigInt int64

// MarshalJSON implements json.Marshaler.
// It converts int64 to string for JSON output.
func (b BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(b), 10))
}

// UnmarshalJSON implements json.Unmarshaler.
// It handles both string and number inputs for flexibility.
func (b *BigInt) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*b = BigInt(int64(val))
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		*b = BigInt(i)
	}
	return nil
}

// Value implements driver.Valuer
func (b BigInt) Value() (driver.Value, error) {
	return int64(b), nil
}

// Scan implements sql.Scanner
func (b *BigInt) Scan(value interface{}) error {
	if value == nil {
		*b = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*b = BigInt(v)
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*b = BigInt(i)
	default:
		return fmt.Errorf("cannot scan type %T into BigInt", value)
	}
	return nil
}

// ToBigIntSlice converts []int64 to []BigInt
func ToBigIntSlice(ids []int64) []BigInt {
	if ids == nil {
		return nil
	}
	res := make([]BigInt, len(ids))
	for i, id := range ids {
		res[i] = BigInt(id)
	}
	return res
}
