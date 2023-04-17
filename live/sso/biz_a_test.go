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
var aSessions = cache.New(time.Minute * 15, time.Second)
//var aSessions = ssoSessions

// 使用 Redis

// 我要先启动一个业务服务器
// 我们在业务服务器上，模拟一个单机登录的过程
func testBizAServer(t *testing.T)  {
	server := web.NewHTTPServer(web.ServerWithMiddleware(LoginMiddlewareServerA))

	// 我要求我这里，必须登录了才能看到，那该怎么办

	// 如果收到一个 HTTP 请求，
	// 方法是 GET
	// 请求是路径是/profile
	// 那么就执行方法里面的逻辑
	server.Get("/profile", func(ctx *web.Context) {
		ctx.RespJSONOK(&User{
			Name: "Tom",
			Age: 18,
		})
	})

	server.Get("/token", func(ctx *web.Context) {
		token, err := ctx.QueryValue("token")
		if err != nil {
			_ = ctx.RespServerError("token 不对")
			return
		}
		signature := Encrypt("server_a")
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
		aSessions.Set(ssid, tokens, time.Minute * 15)
		ctx.SetCookie(&http.Cookie{
			Name: "a_sessid",
			Value: ssid,
		})

		// 你是要跳过去你最开始的 profile 那里
		http.Redirect(ctx.Resp, ctx.Req,"http://aaa.com:8081/profile", 302)
	})

	err := server.Start(":8081")
	t.Log(err)
}



// 登录校验的 middleware
func LoginMiddlewareServerA(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/token" {
			next(ctx)
			return
		}
		redirect := fmt.Sprintf("http://sso.com:8000/login?client_id=server_a")
		cookie, err := ctx.Req.Cookie("a_sessid")
		if err != nil {
			http.Redirect(ctx.Resp, ctx.Req, redirect, 302)
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		tokens, ok := aSessions.Get(ssid)
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

// https://account.geekbang.org/account/oauth/callback?type=wechat&ident=d0435d&login=0&cip=0&redirect=https%3A%2F%2Faccount.geekbang.org%2Fthirdlogin%3Fremember%3D1%26type%3Dwechat%26is_bind%3D0%26gk_cus_user_wechat%3Duniversity%26platform%3Dtime%26wechat%3Dwechatuniversity%26redirect%3Dhttps%253A%252F%252Fu.geekbang.org%252F%26failedurl%3Dhttps%3A%2F%2Faccount.geekbang.org%2Fsignin%3Fgk_cus_user_wechat%3Duniversity%26redirect%3Dhttps%253A%252F%252Fu.geekbang.org%252F&code=091Fsi100L89dP1NOX30023fhp3Fsi13&state=31c542e1a6778b0f9c4fef86b96ba4a8