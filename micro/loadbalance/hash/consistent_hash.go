package hash

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type ConsistentBalancer struct {
	connections []balancer.SubConn
	length int
}

func (b *ConsistentBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 在这个地方你拿不到请求，无法做根据请求特性做负载均衡
	//idx := info.Ctx.Value("user_id")
	//idx := info.Ctx.Value("hash_code")

	return balancer.PickResult{
		SubConn: b.connections[0],
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type ConsistentBalancerBuilder struct {

}

func (b *ConsistentBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, c)
	}
	return &Balancer{
		connections: connections,
		length: len(connections),
	}
}