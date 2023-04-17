package grpc

import (
	"context"
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	us := &Server{}
	server := grpc.NewServer()
	gen.RegisterUserServiceServer(server, us)
	l, err := net.Listen("tcp", ":8081")
	require.NoError(t, err)
	err = server.Serve(l)
	t.Log(err)
}
type Server struct {
	gen.UnimplementedUserServiceServer
}

func (s Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}