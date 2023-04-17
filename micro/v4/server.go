package rpc

import (
	"context"
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/message"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/serialize"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/serialize/json"
	"net"
	"reflect"
)

type Server struct {
	services map[string]reflectionStub
	serializers map[uint8]serialize.Serializer
}

func NewServer() *Server {
	res := &Server{
		services: make(map[string]reflectionStub, 16),
		serializers: make(map[uint8]serialize.Serializer, 4),
	}
	res.RegisterSerializer(&json.Serializer{})
	return res
}

func (s *Server) RegisterSerializer(sl serialize.Serializer) {
	s.serializers[sl.Code()] = sl
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s: service,
		value: reflect.ValueOf(service),
		serializers: s.serializers,
	}
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		// 比较常见的就是端口被占用
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

// 我们可以认为，一个请求包含两部分
// 1. 长度字段：用八个字节表示
// 2. 请求数据：
// 响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		req := message.DecodeReq(reqBs)
		if err != nil {
			return err
		}
		ctx := context.Background()
		oneway, ok := req.Meta["one-way"]
		if ok && oneway == "true" {
			ctx = CtxWithOneway(ctx)
		}
		resp, err := s.Invoke(ctx, req)
		if err != nil {
			// 处理业务 error
			resp.Error = []byte(err.Error())
		}

		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	//if isOneway(ctx) {
	//	go func() {
	//		service, ok := s.services[req.ServiceName]
	//		resp := &message.Response{
	//			RequestID: req.RequestID,
	//			Version: req.Version,
	//			Compresser: req.Compresser,
	//			Serializer: req.Serializer,
	//		}
	//		_, _ = service.invoke(ctx, req)
	//	}()
	//	return nil, errors.New("micro: 微服务服务端 oneway 请求")
	//}
	service, ok := s.services[req.ServiceName]
	resp := &message.Response{
		RequestID: req.RequestID,
		Version: req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
	}
	if !ok {
		return resp, errors.New("你要调用的服务不存在")
	}
	if isOneway(ctx) {
		go func() {
			_, _ = service.invoke(ctx, req)
		}()
		return nil, errors.New("micro: 微服务服务端 oneway 请求")
	}
	respData, err := service.invoke(ctx, req)
	//if isOneway(ctx) {
	//	return nil, errors.New("micro: 微服务服务端 oneway 请求")
	//}
	resp.Data = respData
	if err != nil {
		return resp, err
	}
	return resp, nil
}

type reflectionStub struct {
	s     Service
	value reflect.Value
	serializers map[uint8]serialize.Serializer
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	// 反射找到方法，并且执行调用
	method := s.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	// 暂时我们不知道怎么传这个 context，所以我们就直接写死
	in[0]= reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	serializer, ok := s.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("micro: 不支持的序列化协议")
	}
	err := serializer.Decode(req.Data, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq
	results := method.Call(in)

	// results[0] 是返回值
	// results[1] 是error

	if results[1].Interface() != nil {
		err = results[1].Interface().(error)
	}

	var res []byte
	if results[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = serializer.Encode(results[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return  res, err
}