package registry

import (
	"context"
	"io"
)

type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error
	//UnRegister(ctx context.Context, serviceName string) error

	ListServices(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(serviceName string) (<- chan Event, error)
	//Subscribe(ctx context.Context, serviceName string) (<- chan Event, error)
	//Subscribe(serviceName string, callback func(event Event)) error

	io.Closer
}

type ServiceInstance struct {
	Name string
	// Address 就是最关键的，定位信息
	Address string

	// 这边你可以任意加字段，完全取决于你的服务治理需要什么字段
	Weight uint32
	// 可以考虑再加一个分组字段
	Group string

	// 也可以用这个
	//Attributes map[string]string
}

type Event struct {
	// ADD, DELETE, UPDATE...
	Type string
}