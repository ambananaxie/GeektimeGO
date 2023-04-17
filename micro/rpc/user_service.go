package rpc

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"log"
	"testing"
	"time"
)

type UserService struct {
	// 用反射来赋值
	// 类型是函数的字段，它不是方法（它不是定义在 UserService 上的方法）
	// 本质上是一个字段
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)

	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
	Msg string
}

type UserServiceServer struct {
	Err error
	Msg string
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println(req)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u *UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	log.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: u.Msg,
		},
	}, u.Err
}

func (u *UserServiceServer) Name() string {
	return "user-service"
}

type UserServiceServerTimeout struct {
	t *testing.T
	sleep time.Duration
	Err error
	Msg string
}

func (u *UserServiceServerTimeout) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	if _, ok := ctx.Deadline(); !ok {
		u.t.Fatal("没有设置超时")
	}
	time.Sleep(u.sleep)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u *UserServiceServerTimeout) Name() string {
	return "user-service"
}
