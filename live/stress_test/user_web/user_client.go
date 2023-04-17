package main

import (
	"context"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"google.golang.org/grpc"
)

// 装饰器模式/dipatcher
type UserServiceClient struct {
	client userapi.UserServiceClient
	shadowClient userapi.UserServiceClient
}

func (u *UserServiceClient) CreateUser(ctx context.Context, in *userapi.CreateUserReq, opts ...grpc.CallOption) (*userapi.CreateUserResp, error) {
	return u.GetUserServiceClient(ctx).CreateUser(ctx, in, opts...)
}

func (u *UserServiceClient) FindById(ctx context.Context, in *userapi.FindByIdReq, opts ...grpc.CallOption) (*userapi.FindByIdResp, error) {
	return u.GetUserServiceClient(ctx).FindById(ctx, in, opts...)
}

func (u *UserServiceClient) Login(ctx context.Context, in *userapi.LoginReq, opts ...grpc.CallOption) (*userapi.LoginResp, error) {
	return u.GetUserServiceClient(ctx).Login(ctx, in, opts...)
}

func (u *UserServiceClient) GetUserServiceClient(ctx context.Context) userapi.UserServiceClient {
	if ctx.Value("stress_test") == "true" {
		return u.shadowClient
	}
	return u.client
}

