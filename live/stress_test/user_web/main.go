package main

import (
	"context"
	"errors"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_web/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
)

func main() {
	// cc, err := NewClientConnWrapper("localhost:8081", "localhost:9081")
	// if err != nil {
	// 	panic(err)
	// }
	cc, err := grpc.Dial("localhost:8081", 
		grpc.WithInsecure(), 
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			if ctx.Value("stress_test") == "true" {
				ctx = metadata.AppendToOutgoingContext(ctx, "stress_test", "true")
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}))
	us := userapi.NewUserServiceClient(cc)
	userHdl := handler.NewUserHandler(us)
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(func(ctx *gin.Context) {
		// ctx 里面压测标记位
		// ctx.Request.Header => ctx.Request.Context()
		stressTest := ctx.Request.Header.Get("X-Stress-Test")
		if  stressTest == "true" {
			cctx := context.WithValue(ctx.Request.Context(), "stress_test", "true")
			ctx.Request = ctx.Request.WithContext(cctx)
		}
	})
	r.Use(sessions.Sessions("mysession", store))
	r.Use(func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/user/create" || path == "/user/login" {
			ctx.Next()
			return
		}
		sess := sessions.Default(ctx)
		if sess == nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, errors.New("请登录"))
		}
	})
	userGin := r.Group("/users")

	userGin.POST("/create", userHdl.SignUp)
	userGin.POST("/login", userHdl.Login)
	userGin.GET("/profile", userHdl.Profile)
	if err = r.Run(":8082"); err != nil {
		panic(err)
	}
}