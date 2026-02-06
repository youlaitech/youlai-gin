package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime 自定义时间类型
type LocalTime time.Time

const (
	// TimeFormat 标准时间格式：YYYY-MM-DD HH:mm:ss
	TimeFormat = "2006-01-02 15:04:05"
)

// MarshalJSON 输出指定时间格式
func (t LocalTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf(`"%s"`, time.Time(t).Format(TimeFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON 解析 JSON
func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// 去掉引号
	str := string(data)
	if len(str) > 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	parsed, err := time.ParseInLocation(TimeFormat, str, time.Local)
	if err != nil {
		return err
	}

	*t = LocalTime(parsed)
	return nil
}

// Value 实现 driver.Valuer
func (t LocalTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

// Scan 实现 sql.Scanner
func (t *LocalTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	if v, ok := value.(time.Time); ok {
		*t = LocalTime(v)
		return nil
	}

	return fmt.Errorf("cannot scan type %T into LocalTime", value)
}

// String 实现 Stringer
func (t LocalTime) String() string {
	return time.Time(t).Format(TimeFormat)
}

// Time 转换为 time.Time
func (t LocalTime) Time() time.Time {
	return time.Time(t)
}

// Now 返回当前时间
func Now() LocalTime {
	return LocalTime(time.Now())
}
