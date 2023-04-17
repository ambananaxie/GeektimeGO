package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type WeightBalancer struct {
	connections []*weightConn
	totalWeight uint32
}

func (b *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	tgt := rand.Intn(int(b.totalWeight) + 1)
	var idx int
	for i, c := range b.connections {
		tgt = tgt - int(c.weight)
		if tgt <= 0 {
			idx = i
			break
		}
	}
	return balancer.PickResult{
		SubConn: b.connections[idx].c,
		Done: func(info balancer.DoneInfo) {
			// 可以在这里修改权重，但是要考虑并发安全
		},
	}, nil
}

type WeightBalancerBuilder struct {

}

func (b *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*weightConn, 0, len(info.ReadySCs))
	var totalWeight uint32
	for sub, subInfo := range info.ReadySCs {
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		totalWeight += weight
		cs = append(cs, &weightConn{
			c: sub,
			weight: weight,
		})
	}
	return &WeightBalancer{
		connections: cs,
		totalWeight: totalWeight,
	}
}



type weightConn struct {
	c balancer.SubConn
	weight uint32
}
