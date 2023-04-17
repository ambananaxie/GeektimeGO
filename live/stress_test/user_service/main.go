package main

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository/dao"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/service"
	"github.com/Shopify/sarama"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	"net"
	// rstore "gitee.com/geektime-geekbang/geektime-go/web/session/redis"
	"github.com/go-redis/redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"log"
	_ "net/http/pprof"
	"os"
)

// 这里各种地址都是直接写死的，在真实的环境替换为从配置文件里面读取就可以
// 随便用一个配置框架，大体上都差不多的
func main() {
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

	liveDB, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/userapp"))
	if err != nil {
		panic(err)
	}

	shadowDB, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3307)/userapp_shadow"))
	if err != nil {
		panic(err)
	}

	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "abc",
	})
	c := &ShadowCache{
		c: cache.NewRedisCache(rc),
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &ShadowPool{
			shadow: shadowDB.ConnPool,
			live: liveDB.ConnPool,
		},
	}))
	if err != nil {
		panic(err)
	}
	repo := repository.NewUserRepository(dao.NewUserDAO(db), c)
	us := service.NewUserService(repo, producer)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		vals := metadata.ValueFromIncomingContext(ctx, "stress_test")
		if len(vals) != 0 {
			ctx = context.WithValue(ctx, "stress_test", vals[0])
		}
		return handler(ctx, req)
	}))
	userapi.RegisterUserServiceServer(server, us)

	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	if err = server.Serve(l); err != nil {
		panic(err)
	}

	// 路由注册和 middleware 注册，可以抽取出来作为一个单独的方法，也可以将路由注册部分下沉到 handler 包
	// 例如为 Handler 定义一个新的方法，该方法会注册所有的路由
	// 我一般喜欢在一个集中的地方注册
	// handler => controller MVC 模式
	//userHdl := handler.NewUserHandler(userSvr, sessMgr)
	//server := initSever()
	//server.Post("/signup", userHdl.SignUp)
	//server.Post("/login", userHdl.Login)

	// handler.SetService(userSvr)
	// server.Post("/login", handler.Login)

	//server.Post("/login", userHdl.LoginV1(userSvr, sessMgr))
	// server.Post("/login", userHdl.LoginV2(userSvr, sessMgr))
	//server.Get("/profile", userHdl.Profile)
	//server.Post("/update", userHdl.Update)

	//if err = server.Start(":8081"); err != nil {
	//	panic(err)
	//}
}

func initZipkin() {
	exporter, err := zipkin.New(
		"http://localhost:19411/api/v2/spans",
		zipkin.WithLogger(log.New(os.Stderr, "userapp", log.Ldate|log.Ltime|log.Llongfile)),
	)
	if err != nil {
		panic(err)
	}
	batcher := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("userapp"),
		)),
	)
	otel.SetTracerProvider(tp)
}
