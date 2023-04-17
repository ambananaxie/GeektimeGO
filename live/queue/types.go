package queue

import (
	"context"
)

type Queue[T any] interface {
	// 定义方法
	// 入队和出队两个方法
	// Enqueue()
	// Dequeue
	// EnQueue(time.Second, &User{})
	// EnQueue(timeout time.Duration, data any)error
	// EnQueueV2(ms int, time.Unit, data any)error

	// Go 倾向这种设计

	EnQueue(ctx context.Context, data T)error
	DeQueue(ctx context.Context) (T, error)

	IsFull()bool
	IsEmpty()bool
	Len()uint64
}
