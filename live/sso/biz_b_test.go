package sso

import (
	"bytes"
	"encoding/json"
	"fmt"
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

// var ssoSessions = map[string]any{}
var bSessions = cache.New(time.Minute * 15, time.Second)
//var bSessions =
// 使用 Redis

// 我要先启动一个业务服务器
// 我们在业务服务器上，模拟一个单机登录的过程
func testBizBServer(t *testing.T)  {
	server := web.NewHTTPServer(web.ServerWithMiddleware(LoginMiddlewareServerB))

	// 我要求我这里，必须登录了才能看到，那该怎么办

	// 如果收到一个 HTTP 请求，
	// 方法是 GET
	// 请求是路径是/profile
	// 那么就执行方法里面的逻辑
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespJSONOK(&User{
			Name: "Tom B",
			Age: 18,
		})
	})
	server.Get("/token", func(ctx *web.Context) {
		token, err := ctx.QueryValue("token")
		if err != nil {
			_ = ctx.RespServerError("token 不对")
			return
		}
		signature := Encrypt("server_b")
		// 我拿到了这个 token
		req, err := http.NewRequest(http.MethodPost,
			"http://sso.com:8000/token/validate?token=" + token , bytes.NewBuffer([]byte(signature)))
		if err != nil {
			_ = ctx.RespServerError("解析 token 失败")
			return
		}
		t.Log(req)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			_ = ctx.RespServerError("解析 token 失败")
			return
		}
		tokensBs, _ := io.ReadAll(resp.Body)
		var tokens Tokens
		_ = json.Unmarshal(tokensBs, &tokens)
		// 于是你就拿到了两个 token

		// 往下要干嘛？
		// 这里就是彻底的登录成功了
		ssid := uuid.New().String()
		bSessions.Set(ssid, tokens, time.Minute * 15)
		ctx.SetCookie(&http.Cookie{
			Name: "b_sessid",
			Value: ssid,
		})

		// 你是要跳过去你最开始的 profile 那里

		// 你是要跳过去你最开始的 profile 那里
		http.Redirect(ctx.Resp, ctx.Req,"http://bbb.com:8082/profile", 302)
	})
	err := server.Start(":8082")
	t.Log(err)
}



// 登录校验的 middleware
func LoginMiddlewareServerB(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/token" {
			next(ctx)
			return
		}
		redirect := fmt.Sprintf("http://sso.com:8000/login?client_id=server_b")
		cookie, err := ctx.Req.Cookie("b_sessid")
		if err != nil {
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		tokens, ok := bSessions.Get(ssid)
		if !ok {
			// 你没有登录
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}
		log.Println(tokens)
		// 这边就是登录了
		next(ctx)
	}
}
