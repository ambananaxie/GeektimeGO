package sso

import (
	"gitee.com/geektime-geekbang/geektime-go/web"
	"io"
	"net/http"
	"testing"
)

func testResourceServer(t *testing.T) {
	server := web.NewHTTPServer()
	server.Get("/profile", func(ctx *web.Context) {
		token, _ := ctx.QueryValue("token")
		// 可能是 RPC 调用，因为授权服务和资源服务，都是同一个公司的
		req, _ := http.NewRequest("POST", "http://auth.com:8000/token/validate?token="+token, nil)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			ctx.RespServerError("token不对")
			return
		}
		data, _ := io.ReadAll(resp.Body)
		// 校验 scope
		if string(data) != "basic" {
			ctx.RespServerError("没有权限")
			return
		}
		_ = ctx.RespJSONOK(User{
			Name: "大明",
		})
	})
	server.Start(":8082")
}
