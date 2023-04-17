package sso

import (
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	"html/template"
	"io"
	"net/http"
	"testing"
	"time"
)

// var authSessions = map[string]any{}
var authSessions = cache.New(time.Minute * 15, time.Second)

func TestAuthServer(t *testing.T) {
	whiteList := map[string]string {
		"server_geek": "http://geek.com:8081/callback?code=",
		"server_b": "http://bbb.com:8082/token",
	}
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{
		T: tpl,
	}
	server := web.NewHTTPServer(web.ServerWithTemplateEngine(engine))
	server.Get("/auth", func(ctx *web.Context) {
		clientId, _ := ctx.QueryValue("client_id")
		scope, _ := ctx.QueryValue("scope")
		_ = ctx.Render("confirm.gohtml", map[string]string{"ClientId": clientId, "Scope": scope})
		return
	})

	server.Post("/auth", func(ctx *web.Context) {
		// 我在这儿模拟登录
		if err != nil {
			ctx.RespServerError("系统错误")
			return
		}
		// 校验账号和密码
		scope, _ := ctx.FormValue("scope")
		clientId, _ := ctx.FormValue("client_id")
		fmt.Println(scope, clientId)
		code := uuid.New().String()
		authSessions.Set(code, map[string]string{
			"client_id": clientId,
			"scope": scope,
		}, time.Minute * 16)
		http.Redirect(ctx.Resp, ctx.Req, whiteList[clientId] + code, 302)
	})


	server.Post("/token", func(ctx *web.Context) {
		code, err := ctx.QueryValue("code")
		if err != nil {
			_ = ctx.RespServerError("拿不到 code")
			return
		}
		signature, err := io.ReadAll( ctx.Req.Body)
		if err != nil {
			_ = ctx.RespServerError("拿不到签名")
			return
		}
		clientId, _ := Decrypt(signature)

		// 这里要干嘛？
		val, ok := authSessions.Get(code)

		if !ok {
			_ = ctx.RespServerError("非法 code")
			return
		}
		data := val.(map[string]string)
		codeClientID  := data["client_id"]
		if clientId != codeClientID {
			_ = ctx.RespServerError("有人劫持了 code")
			return
		}

		// 只能使用一次
		authSessions.Delete(code)

		accessToken := uuid.New().String()

		// 要建立一个 accessToke 到权限（scope）的映射
		authSessions.Set(accessToken, data["scope"], time.Minute * 15)
		// 我应该给 a server 一些什么数据？
		// access token + refresh token
		_ = ctx.RespJSONOK(Tokens{
			AccessToken: accessToken,
			AccessTokenExpiration: (time.Minute * 15).Seconds(),
			RefreshToken: uuid.New().String(),
		})
	})

	server.Post("/token/validate", func(ctx *web.Context) {
		token, _ := ctx.QueryValue("token")
		scope, ok := authSessions.Get(token)
		if !ok {
			ctx.RespServerError("非法 token")
			return
		}
		_ = ctx.RespString(200, scope.(string))
	})

	go func() {
		testGeekBangServer(t)
	}()

	go func() {
		testResourceServer(t)
	}()

	// 要在这里提供登录的地方
	server.Start(":8000")
}
