package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	index int32
	connections []balancer.SubConn
	length int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) ==0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	c := b.connections[idx % b.length]
	return balancer.PickResult{
		SubConn: c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {

}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, c)
	}
	return &Balancer{
		connections: connections,
		index: -1,
		length: int32(len(info.ReadySCs)),
	}
}


