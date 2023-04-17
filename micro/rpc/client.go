package rpc

import (
	"context"
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/compresser"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/message"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/serialize"
	"gitee.com/geektime-geekbang/geektime-go/micro/rpc/serialize/json"
	"github.com/silenceper/pool"
	"net"
	"reflect"
	"strconv"
	"time"
)

// InitService 要为 GetById 之类的函数类型的字段赋值
func (c *Client) InitService(service Service) error {
	// 在这里初始化一个 Proxy
	return setFuncField(service, c, c.serializer, c.compresser)
}

func setFuncField(service Service, p Proxy, s serialize.Serializer, c compresser.Compresser) error {
	if service == nil {
		return errors.New("rpc: 不支持 nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	// 只支持指向结构体的一级指针
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}

	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.CanSet() {
			// 这个地方才是真正的将本地调用捕捉到的地方
			fn := func(args []reflect.Value) (results []reflect.Value) {
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())

				// args[0] 是 context
				ctx := args[0].Interface().(context.Context)
				// args[1] 是 req
				reqData, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				reqData, err = c.Compress(reqData)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				meta := make(map[string]string, 2)
				// 我确实设置了超时
				if deadline, ok := ctx.Deadline(); ok {
					meta["deadline"] = strconv.FormatInt(deadline.UnixMilli(), 10)
				}

				if isOneway(ctx) {
					meta["one-way"] = "true"
				}
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Compresser: c.Code(),
					Data:        reqData,
					Serializer: s.Code(),
					Meta: meta,
				}

				req.CalculateHeaderLength()
				req.CalculateBodyLength()

				// 要真的发起调用了
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				var retErr error
				if len(resp.Error) > 0 {
					retErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					err = s.Decode(resp.Data, retVal.Interface())
					if err != nil {
						// 反序列化的 error
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}

				var retErrVal reflect.Value
				if retErr == nil {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					retErrVal = reflect.ValueOf(retErr)
				}

				return []reflect.Value{retVal, retErrVal}
			}
			// 我要设置值给 GetById
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}
	return nil
}

// 长度字段使用的字节数量
const numOfLengthBytes = 8

type Client struct {
	pool pool.Pool
	serializer serialize.Serializer
	compresser compresser.Compresser
}

type ClientOption func(client *Client)

func ClientWithSerializer(sl serialize.Serializer) ClientOption {
	return func(client *Client) {
		client.serializer = sl
	}
}

func ClientWithCompresser(c compresser.Compresser) ClientOption {
	return func(client *Client) {
		client.compresser  = c
	}
}

func NewClient(addr string, opts...ClientOption) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap: 1,
		MaxCap: 30,
		MaxIdle: 10,
		IdleTimeout: time.Minute,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, time.Second * 3)
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	res := &Client{
		pool: p,
		serializer: &json.Serializer{},
		compresser: compresser.DoNothingCompresser{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ch := make(chan struct{})
	defer func() {
		close(ch)
	}()
	var (
		resp *message.Response
		err error
	)
	go func() {
		resp , err = c.doInvoke(ctx, req)
		ch <- struct{}{}
	}()

	select {
	case <- ctx.Done():
		return nil, ctx.Err()
	case <- ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error){
	data := message.EncodeReq(req)
	// 正儿八经地把请求发过去服务端
	resp, err := c.send(ctx, data)
	if err != nil {
		return nil, err
	}
	// 这里才算是中断
	//if ctx.Err() != nil {
	//	return nil, ctx.Err()
	//}
	return message.DecodeResp(resp), nil
}

func (c *Client) send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		c.pool.Put(val)
	}()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	if isOneway(ctx) {
		return nil, errors.New("micro: 这是一个 oneway 调用，你不应该处理任何结果")
	}
	return ReadMsg(conn)
}