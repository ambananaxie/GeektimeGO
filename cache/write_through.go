package cache

import (
	"context"
	"log"
	"time"
)

type WriteThroughCache struct {
	Cache
	StoreFunc func(ctx context.Context, key string, val any) error
}

func (w *WriteThroughCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val)
	if err != nil {
		return err
	}
	return w.Cache.Set(ctx, key, val, expiration)
}

func (w *WriteThroughCache) SetV1(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.Cache.Set(ctx, key, val, expiration)
	if err != nil {
		return err
	}
	return w.StoreFunc(ctx, key, val)
}

func (w *WriteThroughCache) SetV2(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val)
	go func() {
		er := w.Cache.Set(ctx, key, val, expiration)
		if er != nil {
			log.Fatalln(er)
		}
	}()
	return err
}

func (w *WriteThroughCache) SetV3(ctx context.Context, key string, val any, expiration time.Duration) error {
	go func() {
		err := w.StoreFunc(ctx, key, val)
		if err != nil {
			log.Fatalln(err)
		}
		if err = w.Cache.Set(ctx, key, val, expiration); err != nil {
			log.Fatalln(err)
		}
	}()
	return nil
}

type WriteThroughCacheV1[T any] struct {
	Cache
	StoreFunc func(ctx context.Context, key string, val T) error
}

func (w *WriteThroughCacheV1[T]) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	err := w.StoreFunc(ctx, key, val.(T))
	if err != nil {
		return err
	}
	return w.Cache.Set(ctx, key, val, expiration)
}