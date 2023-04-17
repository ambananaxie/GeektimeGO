//go:build e2e
package ratelimit

import (
	"context"
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"github.com/go-redis/redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestRedisFixWindowLimiter_BuildServerInterceptor(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	interceptor := NewRedisFixWindowLimiter(rdb, "user-service", time.Second * 3, 1).BuildServerInterceptor()
	cnt:=0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt ++
		return &gen.GetByIdResp{}, nil
	}
	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)

	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t,  errors.New("触及了瓶颈"), err)
	assert.Nil(t, resp)

	// 睡一个三秒，确保窗口新建了
	time.Sleep(time.Second * 3)
	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)
}

func TestFixWindowLimiter_LimitUnary(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	testCases := []struct {
		name     string
		key      string
		rate     int
		interval time.Duration

		before func(t *testing.T)
		after  func(t *testing.T)

		wantLimit bool
		wantErr   error
	}{
		{
			// 初始化状态
			name:     "init",
			key:      "my-service",
			rate:     1,
			interval: time.Minute,
			before:   func(t *testing.T) {},
			after: func(t *testing.T) {
				val, err := rdb.Get(context.Background(), "my-service").Result()
				require.NoError(t, err)
				assert.Equal(t, "1", val)
				_, err = rdb.Del(context.Background(), "my-service").Result()
				require.NoError(t, err)
			},
		},
		{
			// 初始化状态，但是失败
			name:      "init but limit",
			key:       "my-service",
			rate:      0,
			wantLimit: true,
			interval:  time.Minute,
			before:    func(t *testing.T) {},
			after: func(t *testing.T) {
				_, err := rdb.Get(context.Background(), "my-service").Result()
				require.Equal(t, redis.Nil, err)
			},
		},
		{
			// 触发限流，但是失败
			name:      "limit",
			key:       "my-service",
			rate:      5,
			wantLimit: true,
			interval:  time.Minute,
			before: func(t *testing.T) {
				val, err := rdb.Set(context.Background(), "my-service", 5, time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", val)
			},
			after: func(t *testing.T) {
				val, err := rdb.Get(context.Background(), "my-service").Result()
				require.NoError(t, err)
				assert.Equal(t, "5", val)
				_, _ = rdb.Del(context.Background(), "my-service").Result()
			},
		},
		{
			// 窗口移动，未触发限流
			name:     "window shift",
			key:      "my-service",
			rate:     5,
			interval: time.Minute,
			before: func(t *testing.T) {
				val, err := rdb.Set(context.Background(), "my-service", 5, time.Second).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", val)
				time.Sleep(time.Second * 2)
			},
			after: func(t *testing.T) {
				val, err := rdb.Get(context.Background(), "my-service").Result()
				require.NoError(t, err)
				assert.Equal(t, "1", val)
				_, _ = rdb.Del(context.Background(), "my-service").Result()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			l := NewRedisFixWindowLimiter(rdb, tc.key, tc.interval, tc.rate)
			limit, err := l.limit(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLimit, limit)
		})
	}
}
