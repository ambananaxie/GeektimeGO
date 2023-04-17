package valuer

import (
	"database/sql"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
)

type Value interface {
	Field(name string) (any, error)
	SetColumns(rows *sql.Rows) error
}

type Creator func(model *model.Model, entity any) Value

type Factory interface {
	Create(model *model.Model, entity any) Value
}

type ValuerV1 interface {
	SetColumns(entity any, rows sql.Rows) error
}

// func UnsafeSetColumns(entity any, rows sql.Rows) error {
//
// }
//
// func ReflectSetColumns(entity any, rows sql.Rows) error {
//
// }