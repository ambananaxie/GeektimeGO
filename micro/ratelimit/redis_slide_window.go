package ratelimit

import (
	"context"
	_ "embed"
	"errors"
	"github.com/go-redis/redis/v9"
	"google.golang.org/grpc"
	"time"
)


//go:embed lua/fix_window.lua
var luaSlideWindow string

type RedisSlideWindowLimiter struct {
	client redis.Cmdable
	// 例如 user-service
	service string
	interval time.Duration
	// 阈值
	rate int
}

func NewRedisSlideWindowLimiter(client redis.Cmdable, service string,
	interval time.Duration, rate int) *RedisFixWindowLimiter {
	return &RedisFixWindowLimiter{
		client: client,
		service: service,
		interval: interval,
		rate: rate,
	}
}

func (t *RedisSlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 我预期 lua 脚本会返回一个 bool 值，告诉我要不要限流
		// 使用 FullMethod，那就是单一方法上限流，比如说 GetById
		// 使用服务名来限流，那就是在单一服务上 users.UserService
		// 使用应用名，user-service
		limit, err := t.limit(ctx)
		if err != nil {
			return
		}
		if limit {
			err = errors.New("触及了瓶颈")
			return
		}
		resp, err = handler(ctx, req)
		return
	}
}

func (t *RedisSlideWindowLimiter) limit(ctx context.Context) (bool, error){
	return t.client.Eval(ctx, luaSlideWindow, []string{t.service},
		t.interval.Milliseconds(), t.rate, time.Now().UnixMilli()).Bool()
}