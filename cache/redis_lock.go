package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"time"
)

var (
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold         = errors.New("redis-lock: 你没有持有锁")

	//go:embed lua/unlock.lua
	luaUnlock string

	//go:embed lua/refresh.lua
	luaRefresh string

	//go:embed lua/lock.lua
	luaLock string
)

// Client 就是对 redis.Cmdable 的二次封装
type Client struct {
	client redis.Cmdable
	//valGenerator func() string
	g singleflight.Group
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) SingleflightLock(ctx context.Context,
	key string,
	expiration time.Duration,
	timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	for {
		flag := false
		resCh := c.g.DoChan(key, func() (interface{}, error) {
			flag = true
			return c.Lock(ctx, key, expiration, timeout, retry)
		})
		select {
		case res := <- resCh:
			if flag {
				c.g.Forget(key)
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*Lock), nil
			}
		case <- ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) Lock(ctx context.Context,
	key string,
	expiration time.Duration,
	timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	//val := c.valGenerator()
	for {
		// 在这里重试
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		if res == "OK" {
			return &Lock{
				client: c.client,
				key: key,
				value: val,
				expiration: expiration,
				unlockChan: make(chan struct{}, 1),
			}, nil
		}

		interval, ok := retry.Next()
		if !ok {
			return nil, fmt.Errorf("redis-lock: 超出重试限制, %w", ErrFailedToPreemptLock)
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}
		select {
		case <- timer.C:
		case <- ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) TryLock(ctx context.Context,
	key string,
	expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		// 代表的是别人抢到了锁
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		client: c.client,
		key: key,
		value: val,
		expiration: expiration,
		unlockChan: make(chan struct{}, 1),
	}, nil
}

//func (c *Client) Unlock(ctx context.Context, key string) error {
//
//}

/*func (c *Client) Unlock(ctx context.Context, lock *Lock) error {

}*/



type Lock struct {
	client redis.Cmdable
	key string
	value string
	expiration time.Duration
	unlockChan chan struct{}
}

func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	timeoutChan := make(chan struct{}, 1)
	// 间隔多久续约一次
	ticker := time.NewTicker(interval)
	for{
		select {
		case <- ticker.C:
			// 刷新的超时时间怎么设置
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// 出现了 error 了怎么办？
			err := l.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <- timeoutChan:
			// 刷新的超时时间怎么设置
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// 出现了 error 了怎么办？
			err := l.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}

		case <-l.unlockChan:
			return nil
		}
	}
}

func (l *Lock) Refresh(ctx context.Context, ) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.value, l.expiration.Seconds()).Int64()
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}

func (l *Lock) Unlock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Int64()
	defer func() {
		//close(l.unlockChan)
		select {
		case l.unlockChan <- struct{}{}:
		default:
			// 说明没有人调用 AutoRefresh
		}
	}()
	//if err == redis.Nil {
	//	return ErrLockNotHold
	//}
	if err != nil {
		return err
	}
	if res != 1 {
		return ErrLockNotHold
	}
	return nil
}

//func (l *Lock) Unlock(ctx context.Context) error {
//	// 我现在要先判断一下，这把锁是不是我的锁
//
//	val, err := l.client.Get(ctx, l.key).Result()
//	if err != nil {
//		return err
//	}
//	if val != l.value {
//		return errors.New("锁不是自己的锁")
//	}
//
//	// 在这个地方，键值对被人删了，紧接着另外一个实例加锁
//
//	// 把键值对删掉
//	cnt, err := l.client.Del(ctx, l.key).Result()
//	if err != nil {
//		return err
//	}
//	if cnt != 1 {
//		// 这个地方代表什么？
//		// 代表你加的锁，过期了
//		//log.Info("redis-lock: 解锁失败，锁不存在")
//		//return nil
//		return ErrLockNotHold
//	}
//	return nil
//}
