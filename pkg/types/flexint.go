package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexInt 兼容字符串和数字的 int 类型
// 用于前端可能传 "1" (字符串) 或 1 (数字) 的场景
type FlexInt int

// MarshalJSON 实现 json.Marshaler 接口，输出为数字
func (f FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(f))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
// 兼容字符串 "1" 和数字 1 两种输入格式
func (f *FlexInt) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch val := v.(type) {
	case float64:
		*f = FlexInt(int(val))
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		*f = FlexInt(i)
	case nil:
		*f = 0
	default:
		return fmt.Errorf("cannot unmarshal %T into FlexInt", val)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (f FlexInt) Value() (driver.Value, error) {
	return int64(f), nil
}

// Scan 实现 sql.Scanner 接口
func (f *FlexInt) Scan(value interface{}) error {
	if value == nil {
		*f = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*f = FlexInt(v)
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		*f = FlexInt(i)
	default:
		return fmt.Errorf("cannot scan type %T into FlexInt", value)
	}
	return nil
}

// Int 返回 int 值
func (f FlexInt) Int() int {
	return int(f)
}
