package gin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func TestUserController_GetUser(t *testing.T) {
	g := gin.Default()
	ctrl := &UserController{}
	g.GET("/user", ctrl.GetUser)
	apiG := g.Group("/api")
	apiG.GET("/test", func(context *gin.Context) {
		context.String(200, "hello, world")
	})
	v1Api := apiG.Group("/v1")
	v1Api.GET("/test", func(context *gin.Context) {
		context.String(200, "hello, world, v1")
	})
	g.POST("/user", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello %s", "world")
	})

	g.GET("/static", func(context *gin.Context) {
		// 读文件
		// 写响应
	})
	_ = g.Run(":8082")
}
