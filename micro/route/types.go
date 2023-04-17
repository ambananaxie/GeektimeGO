package route

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// Filter 返回值 true 就是留下来，false 就是丢掉
type Filter func(info balancer.PickInfo, addr resolver.Address) bool


type GroupFilterBuilder struct {}

func (g GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		target, _ := addr.Attributes.Value("group").(string)
		input, _ := info.Ctx.Value("group").(string)
		return target == input
	}
}
