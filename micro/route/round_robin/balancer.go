package round_robin

import (
	"gitee.com/geektime-geekbang/geektime-go/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

type Balancer struct {
	index int32
	connections []subConn
	length int32
	filter route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]subConn, 0, len(b.connections))
	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
	}
	if len(candidates) ==0 {
		// 你也可以考虑筛选完之后，没有任何符合条件的节点，就用默认节点
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := atomic.AddInt32(&b.index, 1)
	c := candidates[int(idx) % len(candidates)]
	return balancer.PickResult{
		SubConn: c.c,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
	Filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]subConn, 0, len(info.ReadySCs))
	for c, ci := range info.ReadySCs {
		connections = append(connections, subConn{
			c:c,
			addr: ci.Address,
		})
	}
	return &Balancer{
		connections: connections,
		index: -1,
		length: int32(len(info.ReadySCs)),
		filter: b.Filter,
	}
}

type subConn struct {
	c balancer.SubConn
	addr resolver.Address
}
