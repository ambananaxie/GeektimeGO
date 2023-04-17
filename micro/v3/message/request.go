package message

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	HeadLength uint32
	BodyLength uint32
	RequestID uint32
	Version uint8
	Compresser uint8
	Serializer uint8

	ServiceName string
	MethodName string
	Meta map[string]string

	Data []byte
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.HeadLength + req.BodyLength)

	// 1. 写入头部长度
	binary.BigEndian.PutUint32(bs[:4], req.HeadLength)
	// 2. 写入 body 长度
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	// 3. 写入 Request ID
	binary.BigEndian.PutUint32(bs[8:12], req.RequestID)
	// 4. 写入 Version
	bs[12] = req.Version
	bs[13] = req.Compresser
	bs[14] = req.Serializer
	cur := bs[15:]

	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]
	cur[0] = '\n'
	cur = cur[1:]
	copy(cur, req.MethodName)
	cur = cur[len(req.MethodName):]
	cur[0] = '\n'
	cur = cur[1:]

	for key, value := range req.Meta {
		copy(cur, key)
		cur = cur[len(key):]
		cur[0]='\r'
		cur = cur[1:]
		copy(cur, value)
		cur = cur[len(value):]
		cur[0] = '\n'
		cur = cur[1:]
	}

	copy(cur, req.Data)

	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}
	// 1. 头四个字节是头部长度
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2. 紧接着，又是四个字节，对应于 body 长度
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3. 又是四个字节，对应于 Request ID
	req.RequestID = binary.BigEndian.Uint32(data[8:12])
	req.Version = data[12]
	req.Compresser = data[13]
	req.Serializer = data[14]

	header := data[15:req.HeadLength]
	// 近似于
	// user-service
	// GetById
	index := bytes.IndexByte(header, '\n')
	// 要引入分隔符，切分 service name 和 method name
	req.ServiceName = string(header[:index])
	// index 所在就是分隔符本身，所以你要 + 1
	header = header[index+1:]

	// 切出来 MethodName
	index = bytes.IndexByte(header, '\n')
	req.MethodName = string(header[:index])
	header = header[index + 1:]

	index = bytes.IndexByte(header, '\n')
	if index != -1 {
		meta := make(map[string]string, 4)
		for index != -1 {
			pair := header[:index]
			// \r 的位置
			pairIndex := bytes.IndexByte(pair, '\r')
			key := string(pair[:pairIndex])
			value := string(pair[pairIndex+1:])
			meta[key] =value

			header = header[index+1:]
			index = bytes.IndexByte(header, '\n')
		}
		req.Meta = meta
	}
	if req.BodyLength != 0 {
		req.Data = data[req.HeadLength:]
	}
	return req
}

func (req *Request) CalculateHeaderLength() {
	// 不要忘了分隔符
	headLength := 15 + len(req.ServiceName) + 1 + len(req.MethodName) + 1
	for key, value := range req.Meta {
		headLength += len(key)
		// key 和 value 之间的分隔符
		headLength ++
		headLength += len(value)
		headLength ++
		// 和下一个 key value 的分隔符
	}
	req.HeadLength = uint32(headLength)
}

func (req *Request) CalculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}