package micro

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/micro/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(server *Server)

type Server struct {
	name string
	registry registry.Registry
	registerTimeout time.Duration
	*grpc.Server
	listener net.Listener
	weight uint32
	group string
}

func NewServer(name string, opts...ServerOption) (*Server, error) {
	res := &Server{
		name: name,
		Server: grpc.NewServer(),
		registerTimeout: time.Second * 10,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func ServerWithWeight(weight uint32) ServerOption {
	return func(server *Server) {
		server.weight = weight
	}
}

func ServerWithGroup(group string) ServerOption {
	return func(server *Server) {
		server.group = group
	}
}

// Start 当用户调用这个方法的时候，就是服务已经准备好
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	// 有注册中心，要注册了
	if s.registry != nil {
		// 在这里注册
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		err = s.registry.Register(ctx, registry.ServiceInstance{
			Name:s.name,
			// 你的定位信息从哪里来？
			Address: listener.Addr().String(),
			Group: s.group,
		})
		if err != nil {
			return err
		}
		// 这里已经注册成功了
		//defer func() {
			// 忽略或者 log 一下错误
			//_ = s.registry.Close()
			//_ = s.registry.UnRegister(registry.ServiceInstance{})
		//}()
	}

	err = s.Serve(listener)
	return err
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}

func ServerWithRegistry(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = r
	}
}