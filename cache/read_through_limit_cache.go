package cache

import (
	"context"
	"fmt"
	"time"
)


// RateLimitReadThroughCache 你一定要赋值 LoadFunc 和 Expiration
// Expiration 是你的过期时间
type RateLimitReadThroughCache struct {
	Cache
	LoadFunc func(ctx context.Context, key string) (any, error)
	Expiration time.Duration
	//loadFunc func(ctx context.Context, key string) (any, error)
	//LoadFunc func(key string) (any, error)
	//logFunc func()
	//g singleflight.Group
}

func (r *RateLimitReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound && ctx.Value("limited") == nil {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			//_ = r.Cache.Set(ctx, key, val, r.Expiration)
			er := r.Cache.Set(ctx, key, val, r.Expiration)
			if er != nil {
				return val, fmt.Errorf("%w, 原因：%s", ErrFailedToRefreshCache, er.Error())
			}
		}
	}
	return val, err
}