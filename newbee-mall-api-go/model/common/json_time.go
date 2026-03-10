package common

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// JSONTime format json time field by myself
type JSONTime struct {
	time.Time
}

// MarshalJSON on JSONTime format Time field with %Y-%m-%d %H:%M:%S
func (t JSONTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

// UnmarshalJSON on JSONTime parse Time field with custom format
func (t *JSONTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	// 处理 null 值
	if str == "null" || str == "\"\"" {
		return nil
	}
	// 去除引号
	str = strings.Trim(str, "\"")

	// 尝试解析自定义格式 "2006-01-02 15:04:05"
	parsedTime, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		// 尝试解析 RFC3339 格式作为备选
		parsedTime, err = time.Parse(time.RFC3339, str)
		if err != nil {
			return fmt.Errorf("cannot parse time: %s", str)
		}
	}
	t.Time = parsedTime
	return nil
}

// Value insert timestamp into mysql need this function.
func (t JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan valueof time.Time
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
