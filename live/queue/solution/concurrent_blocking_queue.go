package solution

import (
	"context"
	"errors"
	"sync"
	"time"
)


type FakeCondV1[T any] struct {
	sync.Cond
	queue *ConcurrentBlockingQueueV1[T]
}

func (c *FakeCondV1[T]) EmptyWaitTimeout(timeout time.Duration) error {
	go func(timout time.Duration) {
		time.Sleep(timeout)
		c.Signal()
	}(timeout)

	if c.queue.IsEmpty() {
		c.Wait()
	}

	if c.queue.IsEmpty() {
		return errors.New("timeout")
	} else {
		return nil
	}
}

func (c *FakeCondV1[T]) FullWaitTimeout(timeout time.Duration) error {
	go func(timout time.Duration) {
		time.Sleep(timeout)
		c.Signal()
	}(timeout)

	if c.queue.IsFull() {
		c.Wait()
	}

	if c.queue.IsFull() {
		return errors.New("timeout")
	} else {
		return nil
	}
}

type ConcurrentBlockingQueueV1[T any] struct {
	data         []T
	mutex        *sync.Mutex
	notfullCond  *FakeCondV1[T]
	notemptyCond *FakeCondV1[T]
	maxSize      int
}

func NewConcurrentBlockingQueueV1[T any](maxsize int) *ConcurrentBlockingQueueV1[T] {
	c := &sync.Mutex{}
	queue := &ConcurrentBlockingQueueV1[T]{
		data:    make([]T, 0, maxsize),
		mutex:   c,
		maxSize: maxsize,
	}
	queue.notfullCond = &FakeCondV1[T]{
		Cond:  *sync.NewCond(c),
		queue: queue,
	}
	queue.notemptyCond = &FakeCondV1[T]{
		Cond:  *sync.NewCond(c),
		queue: queue,
	}
	return queue
}

func (c *ConcurrentBlockingQueueV1[T]) EnQueue(ctx context.Context, data T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()

	//控制住了超时
	for c.IsFull() { //不能使用if

		c.mutex.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := c.notfullCond.FullWaitTimeout(time.Millisecond)
			if err != nil {
				c.data = append(c.data, data)
			}
		}
	}
	c.notemptyCond.Signal()
	c.mutex.Unlock()
	return nil
}
func (c *ConcurrentBlockingQueueV1[T]) DeQueue(ctx context.Context) (T, error) {
	var t T
	if ctx.Err() != nil {
		return t, ctx.Err()
	}
	c.mutex.Lock()
	// defer c.mutex.Unlock()
	for c.IsEmpty() {
		c.mutex.Unlock()
		select {
		case <-ctx.Done():
			return t, ctx.Err()
		default:
			err := c.notemptyCond.EmptyWaitTimeout(time.Millisecond)
			if err != nil {
				t = c.data[0]
				c.data = c.data[1:]
			}
		}
	}
	c.notfullCond.Signal()
	c.mutex.Unlock()
	return t, nil
}
func (c *ConcurrentBlockingQueueV1[T]) Len() uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return uint64(len(c.data))
}
func (c *ConcurrentBlockingQueueV1[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == 0
}
func (c *ConcurrentBlockingQueueV1[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.data) == c.maxSize
}
