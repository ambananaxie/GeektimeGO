//go:build e2e

package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	h := &HTTPServer{}

	h.AddRoute(http.MethodGet, "/user", func(ctx Context) {
		fmt.Println("处理第一件事")
		fmt.Println("处理第二件事")
	})

	handler1 := func(ctx Context) {
		fmt.Println("处理第一件事")
	}

	handler2 := func(ctx Context) {
		fmt.Println("处理第二件事")
	}

	// 用户自己去管这种
	h.AddRoute(http.MethodGet, "/user", func(ctx Context) {
		handler1(ctx)
		handler2(ctx)
	})

	h.Get("/user", func(ctx Context) {

	})
	// h.AddRoute1(http.MethodGet, "/user", handler1, handler2)
	// h.AddRoute1(http.MethodGet, "/user")

	// 用法一 完全委托给 http 包
	// http.ListenAndServe(":8081", h)
	// http.ListenAndServeTLS(":443", "", "", h)

	// 用法二 自己手动管
	h.Start(":8081")
}
