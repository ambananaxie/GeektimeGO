package cache

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	errOverCapacity = errors.New("cache：超过容量限制")
)

// MaxCntCache 控制住缓存住的键值对数量
type MaxCntCache struct {
	*BuildInMapCache
	cnt int32
	maxCnt int32
}

func NewMaxCntCache(c *BuildInMapCache, maxCnt int32) *MaxCntCache {
	res := &MaxCntCache{
		BuildInMapCache: c,
		maxCnt: maxCnt,
	}
	origin := c.onEvicted

	res.onEvicted = func(key string, val any) {
		atomic.AddInt32(&res.cnt, -1)
		if origin != nil {
			origin(key, val)
		}
	}
	return res
}

func (c *MaxCntCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 这种写法，如果 key 已经存在，你这计数就不准了
	//cnt := atomic.AddInt32(&c.cnt, 1)
	//if cnt > c.maxCnt {
	//	atomic.AddInt32(&c.cnt, -1)
	//	return errOverCapacity
	//}
	//return c.BuildInMapCache.Set(ctx, key, val, expiration)

	//c.mutex.Lock()
	//_, ok := c.data[key]
	//if !ok {
	//	c.cnt ++
	//}
	//if c.cnt > c.maxCnt {
	//	c.mutex.Unlock()
	//	return errOverCapacity
	//}
	//c.mutex.Unlock()
	//return c.BuildInMapCache.Set(ctx, key, val, expiration)

	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.data[key]
	if !ok {
		if c.cnt + 1 > c.maxCnt {
			// 后面，你可以在这里设计复杂的淘汰策略
			return errOverCapacity
		}
		c.cnt ++
	}
	return c.set(key, val, expiration)
}