package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly = errors.New("orm: 只支持指向结构体的一级指针")

	// errUnsupportedExpression = errors.New("orm: 不支持的表达式类型")

	ErrNoRows = errors.New("orm: 没有数据")
	// ErrInsertZeroRow 代表插入 0 行
	ErrInsertZeroRow = errors.New("orm: 插入 0 行")
	ErrNoUpdatedColumns = errors.New("orm: 未指定更新的列")
)

// func NewErrUnsupportedExpressionV1(expr any) error {
// 	return fmt.Errorf("%w %v", errUnsupportedExpression, expr)
// }

// @ErrUnsupportedExpression 40001 原因是你输入了乱七八糟的类型
// 解决方案：使用正确的类型
func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("orm: 不支持的表达式类型 %v", expr)
}

func NewErrUnsupportedTable(table any) error {
	return fmt.Errorf("orm: 不支持的TableReference类型 %v", table)
}

func NewErrFailedToRollbackTx(bizErr error, rbErr error, panicked bool) error {
	return fmt.Errorf("orm: 事务闭包回滚失败，业务错误: %w，回滚错误 %s，" +
		"是否 panic: %t", bizErr, rbErr, panicked)
	// return fmt.Errorf("orm: 事务闭包回滚失败，业务错误: %s，回滚错误 %w，" +
	// 	"是否 panic: %t", bizErr, rbErr, panicked)
}

func NewErrUnknownField(name string) error {
	return fmt.Errorf("orm: 未知字段 %s", name)
}

func NewErrUnknownColumn(name string) error {
	return fmt.Errorf("orm: 未知列 %s", name)
}

func NewErrInvalidTagContent(pair string) error {
	return fmt.Errorf("orm: 非法标签值 %s", pair)
}


func NewErrUnsupportedAssignable(expr any) error {
	return fmt.Errorf("orm: 不支持的赋值表达式类型 %v", expr)
}

func NewErrUnsupportedAssignableType(exp any) error {
	return fmt.Errorf("orm: 不支持的 Assignable 表达式 %v", exp)
}