package main

import (
	"gitee.com/geektime-geekbang/geektime-go/cache"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository/dao"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/service"
	"github.com/Shopify/sarama"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net"
	"testing"

	// rstore "gitee.com/geektime-geekbang/geektime-go/web/session/redis"
	"github.com/go-redis/redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	_ "net/http/pprof"
)

func TestShadowServer(t *testing.T) {
	initZipkin()
	// 在 main 函数的入口里面完成所有的依赖组装。
	// 这个部分你可以考虑替换为 google 的 wire 框架，达成依赖注入的效果
	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(lg)
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	// cfg.Producer.
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, cfg)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/userapp"))
	if err != nil {
		panic(err)
	}

	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "abc",
	})
	c := cache.NewRedisCache(rc)

	repo := repository.NewUserRepository(dao.NewUserDAO(db), c)
	us := service.NewUserService(repo, producer)
	server := grpc.NewServer()
	userapi.RegisterUserServiceServer(server, us)

	l, err := net.Listen("tcp", ":8091")
	if err != nil {
		panic(err)
	}
	if err = server.Serve(l); err != nil {
		panic(err)
	}
}
