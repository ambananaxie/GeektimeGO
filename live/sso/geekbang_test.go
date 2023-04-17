package sso

import (
	"bytes"
	"encoding/json"
	"fmt"
	web "gitee.com/geektime-geekbang/geektime-go/web"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	"html/template"
	"io"
	"net/http"
	"testing"
	"time"
)

// var ssoSessions = map[string]any{}
var geekSessions = cache.New(time.Minute * 15, time.Second)
//var geekSessions = ssoSessions

// 使用 Redis

// 我要先启动一个业务服务器
// 我们在业务服务器上，模拟一个单机登录的过程
func testGeekBangServer(t *testing.T)  {
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{
		T: tpl,
	}
	server := web.NewHTTPServer(
		web.ServerWithTemplateEngine(engine),
		web.ServerWithMiddleware(LoginMiddlewareServerGeekbang))

	// 我要求我这里，必须登录了才能看到，那该怎么办

	// 如果收到一个 HTTP 请求，
	// 方法是 GET
	// 请求是路径是/profile
	// 那么就执行方法里面的逻辑
	server.Get("/home", func(ctx *web.Context) {
		cookie, err := ctx.Req.Cookie("geek_sessid")
		if err != nil {
			_ = ctx.RespServerError("服务器错误")
			return
		}

		//var storageDriver ***
		u, _ := geekSessions.Get(cookie.Value)

		ctx.RespString(200, "hello, " + u.(User).Name)
	})

	server.Get("/callback", func(ctx *web.Context) {
		code, err := ctx.QueryValue("code")
		if err != nil {
			_ = ctx.RespServerError("code 不对")
			return
		}
		signature := Encrypt("server_geek")
		// 我拿到了这个 code
		req, err := http.NewRequest(http.MethodPost,
			"http://auth.com:8000/token?code=" +code, bytes.NewBuffer([]byte(signature)))
		if err != nil {
			_ = ctx.RespServerError("解析 code 失败")
			return
		}
		t.Log(req)
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			_ = ctx.RespServerError("解析 code 失败")
			return
		}
		bode, _ := io.ReadAll(resp.Body)
		var tokens Tokens
		_ = json.Unmarshal(bode, &tokens)
		// 于是你就拿到了两个 token

		// 我拿到了这个 code
		// 正常这个 token 要小心被窃取，所以请求还会带上client id（对应于 appid）
		req, err = http.NewRequest(http.MethodGet,
			"http://resource.com:8082/profile?token=" +tokens.AccessToken, nil)
		if err != nil {
			_ = ctx.RespServerError("解析 code 失败")
			return
		}

		resp, err = (&http.Client{}).Do(req)
		if err != nil {
			_ = ctx.RespServerError("获取用户信息失败")
			return
		}
		bode, err = io.ReadAll(resp.Body)
		if err != nil {
			ctx.RespServerError("资源服务器异常")
			return
		}
		var u User
		_ = json.Unmarshal(bode, &u)
		// 往下要干嘛？
		// 这里就是彻底的登录成功了
		ssid := uuid.New().String()
		geekSessions.Set(ssid, u, time.Minute * 15)
		ctx.SetCookie(&http.Cookie{
			Name: "geek_sessid",
			Value: ssid,
		})

		// 你是要跳过去你最开始的请求的 home 那里
		http.Redirect(ctx.Resp, ctx.Req,"http://geek.com:8081/home", 302)
	})

	err = server.Start(":8081")
	t.Log(err)
}



// 登录校验的 middleware
func LoginMiddlewareServerGeekbang(next web.HandleFunc) web.HandleFunc {
	return func(ctx *web.Context) {
		if ctx.Req.URL.Path == "/callback" {
			next(ctx)
			return
		}
		redirect := fmt.Sprintf("http://auth.com:8000/auth?type=code&client_id=server_geek&scope=basic")
		cookie, err := ctx.Req.Cookie("geek_sessid")
		if err != nil {
			_ = ctx.Render("login_geek.gohtml", map[string]string{"RedirectURL": redirect})
			return
		}

		//var storageDriver ***
		ssid := cookie.Value
		_, ok := geekSessions.Get(ssid)
		if !ok {
			_ = ctx.Render("login_geek.gohtml", map[string]string{"RedirectURL": redirect})
			return
		}
		// 这边就是登录了
		next(ctx)
	}
}