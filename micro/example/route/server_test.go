package route

import (
	"context"
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/micro"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"gitee.com/geektime-geekbang/geektime-go/micro/registry/etcd"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	var eg errgroup.Group
	for i := 0; i < 3; i++ {
		var group = "A"
		if i % 2 == 0 {
			group = "B"
			// 压力测试
			// group = "stress"
		}
		server, err := micro.NewServer("user-service",
			micro.ServerWithRegistry(r), micro.ServerWithGroup(group))
		require.NoError(t, err)
		us := &UserServiceServer{group: group}
		gen.RegisterUserServiceServer(server, us)
		// 启动 8081,8082, 8083 三个端口
		port := fmt.Sprintf(":808%d", i+1)
		eg.Go(func() error {
			return  server.Start(port)
		})
	}
	err = eg.Wait()
	t.Log(err)
}

type UserServiceServer struct {
	group string
	gen.UnimplementedUserServiceServer
}

func (s UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	//go func() {
	// 转异步
	//	fmt.Println(s.group)
	//	// 做一些事情
	//}()
	// 返回一个 202
	fmt.Println(s.group)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}