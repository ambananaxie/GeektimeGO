package grpc_resolver

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestClient(t *testing.T)  {
	cc, err := grpc.Dial("registry:///localhost:8081", grpc.WithInsecure(), grpc.WithResolvers(&Builder{}))
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 13})
	require.NoError(t, err)
	t.Log(resp)
}

