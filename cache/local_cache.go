package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache：键不存在")
	//errKeyExpired = errors.New("cache：键过期")
)

type BuildInMapCacheOption func(cache *BuildInMapCache)

type BuildInMapCache struct {
	data map[string]*item
	//data sync.Map
	mutex sync.RWMutex
	close chan struct{}
	onEvicted func(key string, val any)
	//onEvicted func(ctx context.Context, key string, val any)
	//onEvicteds []func(key string, val any)
}

func NewBuildInMapCache(interval time.Duration, opts...BuildInMapCacheOption) *BuildInMapCache {
	res := &BuildInMapCache{
		data: make(map[string]*item, 100),
		close: make(chan struct{}),
		onEvicted: func(key string, val any) {

		},
		// maxCnt: 1000
	}

	for _, opt := range opts {
		opt(res)
	}

	go func() {
		ticker := time.NewTicker(interval)
		for  {
			select {
			case t := <- ticker.C:
				res.mutex.Lock()
				i := 0
				for key, val := range res.data {
					if i > 10000 {
						break
					}
					if val.deadlineBefore(t) {
						res.delete(key)
					}
					i ++
				}
				res.mutex.Unlock()
			case <- res.close:
				return
			}
		}
	}()

	return res
}

func BuildInMapCacheWithEvictedCallback(fn func(key string, val any)) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.onEvicted = fn
	}
}

func (b *BuildInMapCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.set(key, val, expiration)
}

func (b *BuildInMapCache) set(key string, val any, expiration time.Duration) error {
	var dl time.Time
	if expiration >0 {
		dl = time.Now().Add(expiration)
	}
	b.data[key]=&item{
		val: val,
		deadline: dl,
	}
	return nil
}

func (b *BuildInMapCache) Get(ctx context.Context, key string) (any, error) {
	b.mutex.RLock()
	res, ok := b.data[key]
	b.mutex.RUnlock()
	if !ok {
		//return nil, errKeyNotFound
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
	}
	now := time.Now()
	if res.deadlineBefore(now) {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		res, ok = b.data[key]
		if !ok {
			//return nil, errKeyNotFound
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
		}
		if res.deadlineBefore(now) {
			b.delete(key)
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, key)
			//return nil, fmt.Errorf("%w, key: %s", errKeyExpired, key)
		}
	}
	return res.val, nil
}

func (b *BuildInMapCache) Delete(ctx context.Context, key string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.delete(key)
	return nil
}

func (b *BuildInMapCache) LoadAndDelete(ctx context.Context, key string) (any, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	val, ok := b.data[key]
	if !ok {
		return nil, errKeyNotFound
	}
	b.delete(key)
	return val.val, nil
}

func (b *BuildInMapCache) delete(key string) {
	itm, ok := b.data[key]
	if !ok {
		return
	}
	delete(b.data, key)
	b.onEvicted(key, itm.val)
}

// 我要是调用两次 close?
func (b *BuildInMapCache) Close() error {
	select {
	case b.close <- struct{}{}:
	default:
		return errors.New("重复关闭")
	}
	return nil
}

type item struct {
	val any
	deadline time.Time
}

func (i *item) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}

