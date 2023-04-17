package queue

import (
	"context"
	"sync"
	"time"
)

type DelayQueue[T Delayable] struct {
	pq *PriorityQueue[T]
	mu sync.RWMutex
	dequeueSignal *cond
	enqueueSignal *cond
}

func NewDelayQueue[T Delayable](capacity int) *DelayQueue[T] {
	return &DelayQueue[T]{
		pq: NewPriorityQueue[T](capacity, func(src, dst T) int {
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			if srcDelay > dstDelay {
				return 1
			}
			if srcDelay == dstDelay {
				return 0
			}
			return -1
		}),
	}
}

// 入队和并发阻塞队列没太大区别
func (d *DelayQueue[T]) EnQueue(ctx context.Context, data T) error {
	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		default:
		}

		// 跑过来这边，逻辑就是
		// 如果入队后的元素，过期时间更短，那么就要唤醒出队的
		// 或者，一点都不管，就直接唤醒出队的
		d.mu.Lock()

		// 获得堆顶元素
		// top, err := d.pq.Peek()
		err := d.pq.Enqueue(data)
		switch err {
		case nil:

			// 入队成功
			// 发送入队信号，唤醒出队阻塞的


			// 优化
			// 如果新添加进来的元素，比原来堆顶元素deadline还早，说明是新的堆顶，则通知消费者堆顶变更了
			// if data.Deadline().Before(top.Deadline()) {
			//
			// }
			d.enqueueSignal.broadcast()
			return nil
		case ErrOutOfCapacity:
			// 阻塞，开始睡觉了
			ch := d.dequeueSignal.signalCh()
			select {
			case <- ch:
			case <- ctx.Done():
				return ctx.Err()
			}
		default:
			d.mu.Unlock()
			return err
		}
	}


}

// func (d *DelayQueue[T]) EnQueue(ctx context.Context, data T) error {
//
//
//
// }


// 出队就有讲究了：
// 1. Delay() 返回 <= 0 的时候才能出队
// 2. 如果队首的 Delay()=300ms >0，要是 sleep，等待 Delay() 降下去
// 3. 如果正在 sleep 的过程，有新元素来了，
//    并且 Dealay() = 200 比你正在sleep 的时间还要短，你要调整你的 sleep 时间
// 4. 如果 sleep 的时间还没到，就超时了，那么就返回
// sleep 本质上是阻塞（你可以用 time.Sleep，你也可以用 channel）
func (c *DelayQueue[T]) DeQueue(ctx context.Context) (T, error) {
	var timer *time.Timer
	for {
		select {
		case <- ctx.Done():
			var t T
			return t, ctx.Err()
		default:
		}

		// 我该干啥？
		c.mu.Lock()

		// 主要是顾虑锁被人持有很久，以至于早就超时了
		select {
		case <- ctx.Done():
			var t T
			c.mu.Unlock()
			return t, ctx.Err()
		default:
		}

		// 我拿到堆顶
		val, err := c.pq.Peek()
		switch err {
		case nil:
			// 拿到堆顶元素了
			delay := val.Delay()
			if delay <= 0 {
				val, _ = c.pq.Dequeue()
				c.dequeueSignal.broadcast()
				return val, nil
			}
			// 要在这里解锁
			signal := c.enqueueSignal.signalCh()
			if timer == nil {
				timer = time.NewTimer(delay)
			} else {
				timer.Reset(delay)
			}

			// 你一定要在进去 select 之前解锁
			select {
			case <- timer.C:
				// 在这里，不能这么写
				// c.mu.Lock()
				// val, err = c.pq.Dequeue()
				// c.mu.Unlock()
				// return val, err
			case <- ctx.Done():
				var t T
				return t, ctx.Err()
			case <- signal:
			}

		case ErrEmptyQueue:
			// 这个分支代表，队列为空
			signal := c.enqueueSignal.signalCh()
			// 你一定要在进去 select 之前解锁
			select {
			case <- ctx.Done():
				var t T
				return t, ctx.Err()
			case <- signal:
			}
		default:
			c.mu.Unlock()
			var t T
			return t, err
		}
	}
}

type Delayable interface {
	Delay() time.Duration
	// Deadline() time.Time
}

type cond struct {
	signal chan struct{}
	l      sync.Locker
}

func newCond(l sync.Locker) *cond {
	return &cond{
		signal: make(chan struct{}),
		l:      l,
	}
}

// broadcast 唤醒等待者
// 如果没有人等待，那么什么也不会发生
// 必须加锁之后才能调用这个方法
// 广播之后锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *cond) broadcast() {
	signal := make(chan struct{})
	old := c.signal
	c.signal = signal
	c.l.Unlock()
	close(old)
}

// signalCh 返回一个 channel，用于监听广播信号
// 必须在锁范围内使用
// 调用后，锁会被释放，这也是为了确保用户必然是在锁范围内调用的
func (c *cond) signalCh() <-chan struct{} {
	res := c.signal
	c.l.Unlock()
	return res
}