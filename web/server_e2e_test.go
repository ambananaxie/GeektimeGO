//go:build e2e

package web

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

func TestServer(t *testing.T) {
	h := NewHTTPServer()

	h.addRoute(http.MethodGet, "/user", func(ctx *Context) {
		fmt.Println("处理第一件事")
		fmt.Println("处理第二件事")
	})

	// handler1 := func(ctx *Context) {
	// 	fmt.Println("处理第一件事")
	// }
	//
	// handler2 := func(ctx *Context) {
	// 	fmt.Println("处理第二件事")
	// }

	// 用户自己去管这种
	// h.addRoute(http.MethodGet, "/user", func(ctx *Context) {
	// 	handler1(ctx)
	// 	handler2(ctx)
	// })
	//
	// h.Get("/user", func(ctx *Context) {
	//
	// })

	h.Get("/order/detail", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, order detail"))
	})

	h.Get("/order/abc", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %s", ctx.Req.URL.Path)))
	})

	h.Post("/form", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %s", ctx.Req.URL.Path)))
	})

	h.Get("/valuesv1/:id", func(ctx *Context) {
		id, err := ctx.PathValueV1("id").AsInt64()
		if err != nil {
			ctx.Resp.WriteHeader(400)
			ctx.Resp.Write([]byte("id 输入不对"))
			return
		}

		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %d", id)))
	})

	h.Get("/values/:id", func(ctx *Context) {
		idStr, err := ctx.PathValue("id")
		if err != nil {
			ctx.Resp.WriteHeader(400)
			ctx.Resp.Write([]byte("id 输入不对"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			ctx.Resp.WriteHeader(400)
			ctx.Resp.Write([]byte("id 输入不对"))
			return
		}

		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %d", id)))
	})

	type User struct {
		Name string `json:"name"`
	}

	h.Get("/user/123", func(ctx *Context) {
		ctx.RespJSON(202, User{
			Name: "Tom",
		})
	})

	// h.addRoute1(http.MethodGet, "/user", handler1, handler2)
	// h.addRoute1(http.MethodGet, "/user")

	// 用法一 完全委托给 http 包
	// http.ListenAndServe(":8081", h)
	// http.ListenAndServeTLS(":443", "", "", h)

	// 用法二 自己手动管
	h.Start(":8081")
}

// type SafeContext struct {
// 	Context
// 	mutex sync.RWMutex
// }
//
// func (c *SafeContext) RespJSONOK(val any) error {
// 	c.mutex.Lock()
// 	defer c.mutex.Unlock()
// 	return c.Context.RespJSONOK(val)
// }
