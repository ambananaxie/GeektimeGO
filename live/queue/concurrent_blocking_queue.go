package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex *sync.Mutex
	data []T
	// notFull chan struct{}
	// notEmpty chan struct{}
	maxSize int

	notEmptyCond *Cond
	notFullCond *Cond

	count int
	head int
	tail int

	zero T
}

func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		// 即便是 ring buffer，一次性分配完内存，也是有缺陷的
		// 如果不想一开始就把所有的内存都分配好，可以用链表
		data: make([]T, maxSize),
		mutex: m,
		// notFull: make(chan struct{}, 1),
		// notEmpty: make(chan struct{}, 1),
		maxSize: maxSize,
		notFullCond: NewCond(m),
		notEmptyCond: NewCond(m),
	}
}

// func (c *ConcurrentBlockingQueue[T]) Get(index int) (T, error) {
// 	index = (c.head + index) % c.maxSize
// }

func (c *ConcurrentBlockingQueue[T]) EnQueue(ctx context.Context, data T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	for c.isFull() {
		err := c.notFullCond.WaitWithTimeout(ctx)
		if err != nil {
			return err
		}
	}

	// 按需扩容
	// if len(c.data) < c.maxSize {
	// 	c.data = append(c.data, data)
	// 	c.tail ++
	// 	c.count ++
	// } else {
	// 	c.data[c.tail] = data
	// 	c.tail ++
	// 	c.count ++
	// 	if c.tail == c.maxSize {
	// 		c.tail = 0
	// 	}
	// }

	c.data[c.tail] = data
	c.tail ++
	c.count ++
	if c.tail == c.maxSize {
		c.tail = 0
	}

	c.notEmptyCond.Broadcast()
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
		err := c.notEmptyCond.WaitWithTimeout(ctx)
		if err != nil {
			var t T
			return t, err
		}
	}

	// 这里要不要考虑缩容？

	t := c.data[c.head]
	c.data[c.head] = c.zero
	c.head ++
	c.count --
	if c.head == c.maxSize {
		c.head = 0
	}
	c.notFullCond.Broadcast()
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
	return c.count == c.maxSize
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}


func (c *ConcurrentBlockingQueue[T]) isEmpty() bool {
	return c.count == 0
}

func (c *ConcurrentBlockingQueue[T]) Len() uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return uint64(c.count)
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

// Conditional variable implementation that uses channels for notifications.
// Only supports .Broadcast() method, however supports timeout based Wait() calls
// unlike regular sync.Cond.
type Cond struct {
	L sync.Locker
	n unsafe.Pointer
}

func NewCond(l sync.Locker) *Cond {
	c := &Cond{L: l}
	n := make(chan struct{})
	c.n = unsafe.Pointer(&n)
	return c
}

// Waits for Broadcast calls. Similar to regular sync.Cond, this unlocks the underlying
// locker first, waits on changes and re-locks it before returning.
func (c *Cond) Wait() {
	n := c.NotifyChan()
	c.L.Unlock()
	<-n
	c.L.Lock()
}

// Same as Wait() call, but will only wait up to a given timeout.
func (c *Cond) WaitWithTimeout(ctx context.Context) error {
	n := c.NotifyChan()
	c.L.Unlock()
	select {
	case <-n:
		c.L.Lock()
		return nil
	case <- ctx.Done():
		c.L.Lock()
		return ctx.Err()
	}
}

// Returns a channel that can be used to wait for next Broadcast() call.
func (c *Cond) NotifyChan() <-chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

// Broadcast call notifies everyone that something has changed.
func (c *Cond) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	close(*(*chan struct{})(ptrOld))
}