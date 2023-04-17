package ratelimit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue *list.List
	interval int64
	rate int
	mutex sync.Mutex
}

func NewSlideWindowLimiter(interval time.Duration, rate int) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		queue: list.New(),
		interval: interval.Nanoseconds(),
		rate: rate,
	}
}

func (t *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now().UnixNano()
		boundary := now - t.interval

		// 快路径
		t.mutex.Lock()
		length := t.queue.Len()
		if length < t.rate {
			resp, err = handler(ctx, req)
			// 记住了请求的时间戳
			t.queue.PushBack(now)
			t.mutex.Unlock()
			return
		}

		// 慢路径
		timestamp := t.queue.Front()
		// 这个循环把所有不在窗口内的数据都删掉了
		for timestamp != nil && timestamp.Value.(int64) < boundary {
			t.queue.Remove(timestamp)
			timestamp = t.queue.Front()
		}
		length = t.queue.Len()
		t.mutex.Unlock()
		if length >= t.rate {
			err = errors.New("到达瓶颈")
			return
		}
		resp, err = handler(ctx, req)
		// 记住了请求的时间戳
		t.queue.PushBack(now)
		return
	}
}