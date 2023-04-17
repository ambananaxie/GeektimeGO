package solution

import (
"context"
"errors"
"sync"
"time"
)

type MyCond struct {
	isSetTime bool
	isOver    int8
	deadline  time.Duration
	cond      *sync.Cond
	m         *sync.Mutex
}

func NewMyCond(m *sync.Mutex) *MyCond {

	return &MyCond{
		m:      m,
		cond:   sync.NewCond(m),
		isOver: 0,
	}
}

func (c *MyCond) WaitWithTimeOut(ctx context.Context) error {

	c.cond.L.Lock()
	//
	// if c.isSetTime {
	// 	return errors.New("已经设置时间")
	// }

	// c.isSetTime = true
	//go func(dl time.Duration) {
	//	time.Sleep(dl)
	//	if c.isOver == 0 {
	//		c.ch <- struct{}{}
	//		c.isOver = 2
	//	}
	//}(dl)

	// 基本思路：
	// 1. 开一个 goroutine 等待唤醒的信号
	// 2. select + case 监听超时和唤醒信号
	ch := make(chan struct{})
	go func() {
		// goroutine 1-1
		c.cond.Wait()
		select {
		case ch <- struct{}{}:
		default:
			c.cond.Signal()
			c.cond.L.Unlock()
		}
	}()
	c.cond.L.Unlock()

	select {
	case <- ch:
		// 这个分支我被人唤醒了
		if c.isOver == 2 {
			errors.New("已经超时")
		}
		break
	case <-ctx.Done():
		// 这个分支我超时了
		return ctx.Err()
	}

	return nil
}

func (c *MyCond) Signal() error {
	c.cond.Signal()
	//TODO 需要修改通知对应的线程cond的ch，这里应该是需要有一个队列机制
	//c.ch <- struct{}{}
	return nil
}

func (c *MyCond) Broadcast() error {
	c.cond.Broadcast()
	//TODO 需要修改通知对应的线程cond的ch，这里应该是需要有一个队列机制
	return nil
}
