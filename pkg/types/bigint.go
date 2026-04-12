package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

// BigInt 防止前端 JavaScript 精度丢失
type BigInt int64

// MarshalJSON 实现 json.Marshaler 接口
// 将 int64 转为字符串输出，避免 JS 数值精度丢失
func (b BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(b), 10))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
// 兼容字符串和数字两种输入格式
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

// Value 实现 driver.Valuer 接口
func (b BigInt) Value() (driver.Value, error) {
	return int64(b), nil
}

// Scan 实现 sql.Scanner 接口
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

// ToBigIntSlice []int64 转 []BigInt
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

// ToInt64Slice []BigInt 转 []int64
func ToInt64Slice(ids []BigInt) []int64 {
	if ids == nil {
		return nil
	}
	res := make([]int64, len(ids))
	for i, id := range ids {
		res[i] = int64(id)
	}
	return res
}
