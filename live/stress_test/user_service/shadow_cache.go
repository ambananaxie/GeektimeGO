package main

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	"time"
)

type ShadowCache struct {
	c cache.Cache
}

func (s *ShadowCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	// 改 key 的方案
	if ctx.Value("stress_test") == "true" {
		key= "shadow_" + key
	}
	return s.c.Set(ctx, key, val, expiration)
}

func (s *ShadowCache) Get(ctx context.Context, key string) (any, error) {
	// 改 key 的方案
	if ctx.Value("stress_test") == "true" {
		key= "shadow_" + key
	}
	return s.c.Get(ctx, key)
}

func (s *ShadowCache) Delete(ctx context.Context, key string) error {
	if ctx.Value("stress_test") == "true" {
		key= "shadow_" + key
	}
	return s.c.Delete(ctx, key)
}

func (s *ShadowCache) LoadAndDelete(ctx context.Context, key string) (any, error) {
	return s.c.LoadAndDelete(ctx, key)
}

type ShadowCacheV2 struct {
	live cache.Cache
	test cache.Cache
}

func (s *ShadowCacheV2) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	if ctx.Value("stress_test") == "true" {
		return s.test.Set(ctx, key, val, expiration)
	}
	return s.live.Set(ctx, key, val, expiration)
}

func (s *ShadowCacheV2) Get(ctx context.Context, key string) (any, error) {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowCacheV2) Delete(ctx context.Context, key string) error {
	// TODO implement me
	panic("implement me")
}

func (s *ShadowCacheV2) LoadAndDelete(ctx context.Context, key string) (any, error) {
	// TODO implement me
	panic("implement me")
}

