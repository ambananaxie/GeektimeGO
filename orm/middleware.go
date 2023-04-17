package orm

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
)

type QueryContext struct {
	// 查询类型，标记增删改查
	Type string

	// 代表的是查询本身
	Builder QueryBuilder

	Model *model.Model
}

type QueryResult struct {
	// Result 在不同查询下类型是不同的
	// SELECT 可以是 *T, 也可以是 []*T
	// 其它就是类型 Result
	Result any
	// 查询本身出的问题
	Err error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult
// type Handler func(ctx context.Context, qc *QueryContext) (*QueryResult, error)

type Middleware func(next Handler)Handler

