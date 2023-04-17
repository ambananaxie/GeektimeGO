package errhdl

import "gitee.com/geektime-geekbang/geektime-go/web"

type MiddlewareBuilder struct {
	// 这种设计只能返回固定的值
	// 不能做到动态渲染
	resp map[int][]byte
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		resp: map[int][]byte{},
	}
}

func (m *MiddlewareBuilder) AddCode(status int, data []byte) *MiddlewareBuilder {
	m.resp[status]=data
	return m
}

func (m MiddlewareBuilder) Build() web.Middleware{
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			next(ctx)
			resp, ok := m.resp[ctx.RespStatusCode]
			if ok {
				// 篡改结果
				ctx.RespData = resp
			}
		}
	}
}