package sql_demo

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JsonColumn[T any] struct {
	Val T
	// NULL 的问题
	Valid bool
}

func (j JsonColumn[T]) Value() (driver.Value, error) {
	// NULL
	if !j.Valid {
		return nil, nil
	}
	return json.Marshal(j.Val)
}

func (j *JsonColumn[T]) Scan(src any) error {
	//    int64
	//    float64
	//    bool
	//    []byte
	//    string
	//    time.Time
	//    nil - for NULL values
	var bs []byte
	switch data := src.(type) {
	case string:
		// 你可以考虑额外处理空字符串
		bs = []byte(data)
	case []byte:
		// 你也可以考虑额外处理 []byte{}
		bs = data
	case nil:
		// 说明数据库里面存的就是 NULL
		return nil
	default:
		return errors.New("不支持类型")
	}

	err := json.Unmarshal(bs, &j.Val)
	if err == nil {
		j.Valid = true
	}
	return err
}

