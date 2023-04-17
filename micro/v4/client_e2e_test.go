package rpc

import (
	"context"
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/micro/proto/gen"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/serialize/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInitServiceProto(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerializer(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct{
		name string
		mock func()

		wantErr error
		wantResp *GetByIdResp
	} {
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{},
			wantErr: errors.New("mock error"),
		},

		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
	}
}

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	//usClientOneway := &UserService{}
	client, err := NewClient(":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	//err = client.InitService(usClientOneway, true)
	require.NoError(t, err)

	//ctx := context.WithValue(context.Background(), "oneway", true)
	//usClient.GetById(ctx, req)

	testCases := []struct{
		name string
		mock func()

		wantErr error
		wantResp *GetByIdResp
	} {
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{},
			wantErr: errors.New("mock error"),
		},

		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			//resp, er := usClient.GetById(CtxWithOneway(context.Background()), &GetByIdReq{})
			// 我这边就不能用 resp 了，因为啥都没有
			resp, er := usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
			//var respAsync *GetByIdResp
			//var wg sync.WaitGroup
			//wg.Add(1)
			//go func() {
			//	respAsync, err = usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
			//	wg.Done()
			//}()
			//
			//// 干了很多事
			//
			//// 使用 respAsync
			//wg.Wait()
			//respAsync.Msg

			// 回调
			//go func() {
			//	respAsync, err1 := usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
			//	// 随便你怎么处理
			//	respAsync.Msg
			//}()

			// 虚假的单向调用
			//go func() {
			//	_, _ = usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
			//}()

			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestOneway(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct{
		name string
		mock func()

		wantErr error
		wantResp *GetByIdResp
	} {
		{
			name: "oneway",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{},
			wantErr: errors.New("micro: 这是一个 oneway 调用，你不应该处理任何结果"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			ctx := CtxWithOneway(context.Background())
			resp, er := usClient.GetById(ctx, &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}