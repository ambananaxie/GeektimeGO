package message

import (
	"encoding/binary"
)

type Response struct {
	HeadLength uint32
	BodyLength uint32
	RequestID uint32
	Version uint8
	Compresser uint8
	Serializer uint8
	Error []byte

	Data []byte
}

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.HeadLength + resp.BodyLength)

	// 1. 写入头部长度
	binary.BigEndian.PutUint32(bs[:4], resp.HeadLength)
	// 2. 写入 body 长度
	binary.BigEndian.PutUint32(bs[4:8], resp.BodyLength)
	// 3. 写入 Request ID
	binary.BigEndian.PutUint32(bs[8:12], resp.RequestID)
	// 4. 写入 Version
	bs[12] = resp.Version
	bs[13] = resp.Compresser
	bs[14] = resp.Serializer
	cur := bs[15:]

	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]
	copy(cur, resp.Data)
	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}
	// 1. 头四个字节是头部长度
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2. 紧接着，又是四个字节，对应于 body 长度
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3. 又是四个字节，对应于 Request ID
	resp.RequestID = binary.BigEndian.Uint32(data[8:12])
	resp.Version = data[12]
	resp.Compresser = data[13]
	resp.Serializer = data[14]
	if resp.HeadLength > 15 {
		resp.Error = data[15:resp.HeadLength]
	}

	if resp.BodyLength != 0 {
		resp.Data = data[resp.HeadLength:]
	}
	return resp
}

func (resp *Response) CalculateHeaderLength() {
	resp.HeadLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}