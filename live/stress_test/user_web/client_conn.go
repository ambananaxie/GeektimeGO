package main

import (
	"context"
	"google.golang.org/grpc"
)

type ClientConnWrapper struct {
	// 到生产环境
	cc grpc.ClientConnInterface
	// 到压测环境
	shadowCC grpc.ClientConnInterface
}

func (c *ClientConnWrapper) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	return c.getCC(ctx).Invoke(ctx, method, args, reply, opts...)
}

func (c *ClientConnWrapper) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.getCC(ctx).NewStream(ctx, desc, method, opts...)
}

func (u *ClientConnWrapper) getCC(ctx context.Context) grpc.ClientConnInterface {
	if ctx.Value("stress_test") == "true" {
		return u.shadowCC
	}
	return u.cc
}


func NewClientConnWrapper(ccAddrss string, shadowAddress string) (*ClientConnWrapper, error) {
	cc, err := grpc.Dial(ccAddrss, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	shadowCc, err := grpc.Dial(shadowAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &ClientConnWrapper{
		cc: cc,
		shadowCC: shadowCc,
	}, nil
}