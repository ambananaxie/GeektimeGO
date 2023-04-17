package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
)

type WeightBalancer struct {
	connections []*weightConn
	//mutex sync.Mutex
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight uint32
	var res *weightConn
	//w.mutex.Lock()
	//defer w.mutex.Unlock()
	for _, c := range w.connections {
		c.mutex.Lock()
		totalWeight = totalWeight + c.efficientWeight
		c.currentWeight = c.currentWeight + c.efficientWeight
		if res == nil || res.currentWeight < c.currentWeight{
			res = c
		}
		c.mutex.Unlock()
	}
	res.mutex.Lock()
	res.currentWeight = res.currentWeight - totalWeight
	res.mutex.Unlock()
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			res.mutex.Lock()
			if info.Err != nil && res.efficientWeight == 0 {
				return
			}
			if info.Err == nil &&  res.efficientWeight == math.MaxUint32 {
				return
			}
			if info.Err != nil {
				res.efficientWeight --
			} else {
				res.efficientWeight ++
			}
			res.mutex.Unlock()

			//for {
			//	weight := atomic.LoadUint32(&res.efficientWeight)
			//	if info.Err != nil && weight == 0 {
			//		return
			//	}
			//	if info.Err == nil && weight == math.MaxUint32 {
			//		return
			//	}
			//	newWeight := weight
			//	if info.Err != nil {
			//		newWeight --
			//	} else {
			//		newWeight ++
			//	}
			//	if atomic.CompareAndSwapUint32(&(res.efficientWeight), weight, newWeight) {
			//		return
			//	}
			//}
		},
	}, nil
}

//func (b *Balancer) done(res *weightConn) func(info balancer.DoneInfo) {
//
//}

type WeightBalancerBuilder struct {

}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*weightConn, 0, len(info.ReadySCs))
	for sub, subInfo := range info.ReadySCs {
		//weightStr := subInfo.Address.Attributes.Value("weight").(string)
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		//if !ok || weightStr == "" {
		//	panic()
		//}
		//weight, err := strconv.ParseUint(weightStr, 10, 64)
		//if err != nil {
		//	panic(err)
		//}

		cs = append(cs, &weightConn{
			c: sub,
			weight: uint32(weight),
			currentWeight: uint32(weight),
			efficientWeight: uint32(weight),
		})
	}
	return &WeightBalancer{
		connections: cs,
	}
}

type weightConn struct {
	mutex sync.Mutex
	c balancer.SubConn
	weight uint32
	currentWeight uint32
	efficientWeight uint32
}




