
package orm

import "gitee.com/geektime-geekbang/geektime-go/orm/homework_subquery_answer/internal/errs"

// 将内部的 sentinel error 暴露出去
var (
	// ErrNoRows 代表没有找到数据
	ErrNoRows = errs.ErrNoRows
)
