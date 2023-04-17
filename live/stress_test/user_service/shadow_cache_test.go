package main

import (
	"context"
	"github.com/go-redis/redis/v9"
	"net"
	"testing"
	"time"
)

func TestRedisClient(t *testing.T) {
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "abc",
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if ctx.Value("stress_test") == "true" {
				return net.Dial("tcp", "localhost:16379")
			}
			return net.Dial("tcp", "localhost:6379")
		},
	})
	rc.Set(context.WithValue(context.Background(),
		"stress_test", "true"), "key1", "value1", time.Minute)

	rc.Set(context.Background(),"key1", "value1", time.Minute)
}
