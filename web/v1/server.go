package web

import (
	"net"
	"net/http"
)

type HandleFunc func(ctx Context)

// 确保一定实现了 Server 接口
var _ Server = &HTTPServer{}

type Server interface {
	http.Handler
	Start(add string) error
	// Start1() error

	// AddRoute 路由注册功能
	// method 是 HTTP 方法
	// path 是路由
	// handleFunc 是你的业务逻辑
	AddRoute(method string, path string, handleFunc HandleFunc)
	// 这种允许注册多个，没有必要提供
	// 让用户自己去管
	// AddRoute1(method string, path string, handles ...HandleFunc)
}

//
// type HTTPSServer struct {
// 	HTTPServer
// }

type HTTPServer struct {
	// addr string 创建的时候传递，而不是 Start 接收。这个都是可以的
}

// ServeHTTP 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 你的框架代码就在这里
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	h.serve(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	// 接下来就是查找路由，并且执行命中的业务逻辑
}

func (h *HTTPServer) AddRoute(method string, path string, handleFunc HandleFunc) {
	// 这里注册到路由树里面
	// panic("implement me")
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	h.AddRoute(http.MethodOptions, path, handleFunc)
}

// func (h *HTTPServer) AddRoute1(method string, path string, handleFunc ...HandleFunc) {
// 	panic("implement me")
// }

func (h *HTTPServer) Start(addr string) error {
	// 也可以自己创建 Server
	// http.Server{}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 在这里，可以让用户注册所谓的 after start 回调
	// 比如说往你的 admin 注册一下自己这个实例
	// 在这里执行一些你业务所需的前置条件

	return http.Serve(l, h)
}

func (h *HTTPServer) Start1(addr string) error {
	return http.ListenAndServe(addr, h)
}
