package cache

import (
	"context"
	"fmt"
)

type BloomFilterCache struct {
	ReadThroughCache
}

func NewBloomFilterCache(cache Cache, bf BloomFilter,
	loadFunc func(ctx context.Context, key string) (any, error)) *BloomFilterCache {
	return &BloomFilterCache{
		ReadThroughCache: ReadThroughCache{
			Cache: cache,
			LoadFunc: func(ctx context.Context, key string) (any, error) {
				if !bf.HasKey(ctx, key) {
					return nil, errKeyNotFound
				}
				return loadFunc(ctx, key)
			},
		},
	}
}


type BloomFilterCacheV1 struct {
	ReadThroughCache
	Bf BloomFilter
}

func (r *BloomFilterCacheV1) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound && r.Bf.HasKey(ctx, key){
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

type BloomFilter interface {
	HasKey(ctx context.Context, key string) bool
}
