package broadcast

import (
	"context"
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/micro"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"gitee.com/geektime-geekbang/geektime-go/micro/registry/etcd"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestUseBroadCast(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	var eg errgroup.Group
	var servers []*UserServiceServer
	for i := 0; i < 3; i++ {
		server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r))
		require.NoError(t, err)
		us := &UserServiceServer{
			idx: i,
		}
		servers = append(servers, us)
		gen.RegisterUserServiceServer(server, us)
		// 启动 8081,8082, 8083 三个端口
		port := fmt.Sprintf(":808%d", i+1)
		eg.Go(func() error {
			return  server.Start(port)
		})
		defer func() {
			_ = server.Close()
		}()
	}
	time.Sleep(time.Second * 3)

	client, err := micro.NewClient(micro.ClientInsecure(),
		micro.ClientWithRegistry(r, time.Second * 3))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()
	require.NoError(t, err)
	ctx, respChan := UseBroadCast(ctx)
	go func() {
		for res := range respChan {
			fmt.Println(res.Err, res.Reply)
		}
	}()
	bd := NewClusterBuilder(r, "user-service", grpc.WithInsecure())
	cc, err := client.Dial(ctx, "user-service", grpc.WithUnaryInterceptor(bd.BuildUnaryInterceptor()))
	require.NoError(t, err)
	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
	require.NoError(t, err)
	t.Log(resp)
	for _, s := range servers {
		require.Equal(t, 1, s.cnt)
	}
}

type UserServiceServer struct {
	idx int
	cnt int
	gen.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	s.cnt ++
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: fmt.Sprintf("hello, world %d", s.idx),
		},
	}, nil
}