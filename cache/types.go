package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	// Set(ctx context.Context, key string, val []byte, expiration time.Duration) error
	// millis 毫秒数，过期时间
	// Set(key string, val any, mills int64)

	// Get 方法返回值
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error
	// 同时会把被删除的数据返回
	// Delete(key string) (any, error)

	LoadAndDelete(ctx context.Context, key string) (any, error)
}

type CacheV2[T any] interface {
	Set(ctx context.Context, key string, val T, expiration time.Duration) error
	// Set(ctx context.Context, key string, val []byte, expiration time.Duration) error
	// millis 毫秒数，过期时间
	// Set(key string, val any, mills int64)

	// Get 方法返回值
	Get(ctx context.Context, key string) (T, error)
	Delete(ctx context.Context, key string) error
}

//type CacheV3 interface {
//	Set[T any](ctx context.Context, key string, val T, expiration time.Duration) error
	// Set(ctx context.Context, key string, val []byte, expiration time.Duration) error
	// millis 毫秒数，过期时间
	// Set(key string, val any, mills int64)

	// Get 方法返回值
	//Get[T any](ctx context.Context, key string) (T, error)
	//Delete(ctx context.Context, key string) error
	// 同时会把被删除的数据返回
	// Delete(key string) (any, error)
//}