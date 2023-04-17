package ratelimit

import (
	"context"
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestFixWindowLimiter_BuildServerInterceptor(t *testing.T) {
	// 三秒钟只能有一个请求
	interceptor := NewFixWindowLimiter(time.Second * 3, 1).BuildServerInterceptor()
	cnt:=0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt ++
		return &gen.GetByIdResp{}, nil
	}
	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)

	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t,  errors.New("触发瓶颈了"), err)
	assert.Nil(t, resp)

	// 睡一个三秒，确保窗口新建了
	time.Sleep(time.Second * 3)
	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)
}