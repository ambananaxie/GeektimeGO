package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
	close chan struct{}
}

// NewTokenBucketLimiter interval 隔多久产生一个令牌
func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	ch := make(chan struct{}, capacity)
	closeCh := make(chan struct{})
	producer := time.NewTicker(interval)
	go func() {
		defer producer.Stop()
		for {
			select {
			case <- producer.C:
				select {
				case ch <- struct{}{}:
				default:
					// 没人取令牌
				}
			case <- closeCh:
				return
			}
		}
	}()
	return &TokenBucketLimiter{
		tokens: ch,
		close: closeCh,
	}
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 要在这里拿到令牌
		select {
		case <-t.close:
			// 你已经关掉故障检测了
			//resp, err = handler(ctx, req)
			err = errors.New("缺乏保护，拒绝请求")
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <- t.tokens:
			resp, err = handler(ctx, req)
		//default:
		//	err = errors.New("到达瓶颈")
		//	return
		}

		return
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.close)
	return nil
}

//func (t *TokenBucketLimiter) BuildClientInterceptor() grpc.UnaryClientInterceptor {
//
//}

