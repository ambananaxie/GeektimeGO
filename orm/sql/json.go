package sql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// T 是一个可以被 json 处理的类型
type JsonColumn[T any] struct {
	Val T

	// 主要解决 NULL 之类的问题
	Valid bool
}

func (j *JsonColumn[T]) Scan(src any) error {
	var bs []byte
	switch data := src.(type) {
	case []byte:
		bs = data
	case string:
		bs = []byte(data)
	case nil:
		return nil
	default:
		return fmt.Errorf("ekit：JsonColumn.Scan 不支持 src 类型 %v", src)
	}
	err := json.Unmarshal(bs, &j.Val)
	if err == nil {
		j.Valid = true
	}
	return err
}

func (j JsonColumn[T]) Value() (driver.Value, error) {
	// 我没有数据
	if !j.Valid {
		return nil, nil
	}
	return json.Marshal(j.Val)
}

