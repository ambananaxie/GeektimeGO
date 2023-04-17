package orm

import "gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"

// 通过这种形式将内部错误，暴露在外面
var ErrNoRows = errs.ErrNoRows
