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

// var ssoSessions = map[string]any{}
var ssoSessions = cache.New(time.Minute * 15, time.Second)

func TestSSOServer(t *testing.T) {
	whiteList := map[string]string {
		"server_a": "http://aaa.com:8081/token",
		"server_b": "http://bbb.com:8082/token",
	}
	tpl, err := template.ParseGlob("template/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{
		T: tpl,
	}
	server := web.NewHTTPServer(web.ServerWithTemplateEngine(engine))
	server.Get("/login", func(ctx *web.Context) {
		// 就在这里，你要判断有没有登录


		clientId, _ := ctx.QueryValue("client_id")
		if err != nil {
			_ = ctx.Render("login.gohtml", map[string]string{"ClientId": clientId})
			return
		}

		// 如果 client id 和已有 session 归属不同的主体，那么还是要重新登陆

		// 假如说这里 client id 归属不同的主体，怎么处理了？

		ck, err := ctx.Req.Cookie(fmt.Sprintf("token"))
		if err != nil {
			_ = ctx.Render("login.gohtml", map[string]string{"ClientId": clientId})
			return
		}

		// 就是要建立一个 client_id 到 session 的映射

		_, ok := ssoSessions.Get(ck.Value)
		//_, ok := ssoSessions.Get(clientId)
		if !ok {
			_ = ctx.Render("login.gohtml", map[string]string{"ClientId": clientId})
			return
		}

		// 直接颁发 token
		token := uuid.New().String()
		ssoSessions.Set(clientId, token, time.Minute)
		http.Redirect(ctx.Resp, ctx.Req, whiteList[clientId] + "?token=" + token, 302)
	})

	server.Post("/login", func(ctx *web.Context) {
		// 我在这儿模拟登录
		if err != nil {
			ctx.RespServerError("系统错误")
			return
		}
		// 校验账号和密码
		email, _ := ctx.FormValue("email")
		password, _ := ctx.FormValue("password")
		clientId, _ := ctx.FormValue("client_id")
		if email == "abc@biz.com" && password == "123" {
			// 认为登录成功
			// 要防止 token 被盗走，不能使用 uuid
			id := uuid.New().String()
			http.SetCookie(ctx.Resp, &http.Cookie{
				Name: "token",
				Value: id,
				Expires: time.Now().Add(time.Minute * 15),
			})
			ssoSessions.Set(id, &User{Name: "Tom"}, time.Minute * 15)
			token := uuid.New().String()
			ssoSessions.Set(clientId, token, time.Minute)
			http.Redirect(ctx.Resp, ctx.Req, whiteList[clientId] + "?token=" + token, 302)
			return
		}
		ctx.RespServerError("用户账号名密码不对")
	})

	// 我要提供一个校验 token 的接口，怎么提供？
	// 谁都可以发，怎么保护这里？？？？
	// 1. 频率限制：
	// 2. 来源
	server.Post("/token/validate", func(ctx *web.Context) {
		token, err := ctx.QueryValue("token")
		if err != nil {
			_ = ctx.RespServerError("拿不到 token")
			return
		}
		signature, err := io.ReadAll( ctx.Req.Body)
		if err != nil {
			_ = ctx.RespServerError("拿不到签名")
			return
		}
		clientId, _ := Decrypt(signature)

		// 这里要干嘛？
		val, ok := ssoSessions.Get(clientId)
		// 放这里有隐患
		//ssoSessions.Delete(clientId)
		if !ok {
			// 可能过期了，或者说这个 client id 根本没有过来登录
			_ = ctx.RespServerError("没登录")
			return
		}
		if token != val {
			_ = ctx.RespServerError("token 不对")
			return
		}

		// 只能使用一次
		ssoSessions.Delete(clientId)

		// 我应该给 a server 一些什么数据？
		// access token + refresh token
		_ = ctx.RespJSONOK(Tokens{
			AccessToken: uuid.New().String(),
			RefreshToken: uuid.New().String(),
		})
	})

	go func() {
		testBizAServer(t)
	}()

	go func() {
		testBizBServer(t)
	}()

	// 要在这里提供登录的地方
	server.Start(":8000")
}


type Tokens struct {
	AccessToken string `json:"access_token"`
	AccessTokenExpiration float64 `json:"access_token_expiration"`
	RefreshToken string `json:"refresh_token"`
}