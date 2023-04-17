//go:build v1
package queue

import (
	"context"
	"sync"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex *sync.Mutex
	data []T
	// notFull chan struct{}
	// notEmpty chan struct{}
	maxSize int

	notEmptyCond *cond
	notFullCond *cond
}

func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		data: make([]T, 0, maxSize),
		mutex: m,
		// notFull: make(chan struct{}, 1),
		// notEmpty: make(chan struct{}, 1),
		maxSize: maxSize,
		notFullCond: &cond{
			Cond: sync.NewCond(m),
		},
		notEmptyCond: &cond{
			Cond: sync.NewCond(m),
		},
	}
}


func (c *ConcurrentBlockingQueue[T]) EnQueue(ctx context.Context, data T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	for c.isFull() {
		err := c.notFullCond.WaitTimeout(ctx)
		if err != nil {
			return err
		}
	}
	c.data = append(c.data, data)
	c.notEmptyCond.Signal()
	c.mutex.Unlock()
	// 没有人等 notEmpty 的信号，这一句就会阻塞住
	return nil
}

func (c *ConcurrentBlockingQueue[T]) DeQueue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	c.mutex.Lock()
	for c.isEmpty() {
		err := c.notEmptyCond.WaitTimeout(ctx)
		if err != nil {
			var t T
			return t, err
		}
	}
	t := c.data[0]
	c.data = c.data[1:]
	c.notFullCond.Signal()
	c.mutex.Unlock()
	// 没有人等 notFull 的信号，这一句就会阻塞住
	return t, nil
}

func (c *ConcurrentBlockingQueue[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ConcurrentBlockingQueue[T]) isFull() bool {
	return len(c.data) == c.maxSize
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}


func (c *ConcurrentBlockingQueue[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentBlockingQueue[T]) Len() uint64 {
	return uint64(len(c.data))
}


// func (c *ConcurrentBlockingQueue[T]) EnQueueV1(ctx context.Context, data T) error {
	// select {
	// case <- c.notFullCond.Wait():
	// 	case <- ctx.Done() :
	//
	// }

// 	c.notFullCond.Wait(timeout)
// }

// func (c *ConcurrentBlockingQueue[T]) DeQueueV1(ctx context.Context, data T) error {
// 	c.notFullCond.Signal()
// 	return nil
// }

type cond struct {
	*sync.Cond
}

func (c *cond) WaitTimeout(ctx context.Context) error {
	ch := make(chan struct{})
	go func() {
		c.Cond.Wait()
		select {
		case ch<- struct{}{}:
		default:
			// 这里已经超时返回了
			c.Cond.Signal()
			c.Cond.L.Unlock()
		}
	}()
	select {
	case <- ctx.Done():
		return ctx.Err()
	case <- ch:
		// 你真的被唤醒了
		return nil
	}
}