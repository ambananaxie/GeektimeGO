//go:build e2e
package web

import "testing"

func TestNamespace_Get(t *testing.T) {
	s := NewHTTPServer()
	api := s.Namespace("/api")
	api.Get("/test", func(ctx *Context) {
		_ = ctx.RespString(200, "hello, world, 这是我的")
	})
	v1API := api.Namespace("/v1")
	v1API.Get("/test", func(ctx *Context) {
		_ = ctx.RespString(200, "hello, world, 这是我的 v1")
	})
	_ = s.Start(":8082")
}
