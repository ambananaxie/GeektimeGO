package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func Connect(network, addr string) error {
	conn, err := net.DialTimeout(network, addr, time.Second * 3)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()
	for i := 0; i < 10; i++ {
		_, err = conn.Write([]byte("hello"))
		if err != nil {
			return err
		}
		res := make([]byte, 128)
		_, err = conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println(string(res))
	}
	return nil
}

type Client struct {
	network string
	addr string
}

func (c *Client) Send(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second * 3)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	reqLen := len(data)

	// 我要在这，构建请求数据
	// data = reqLen 的 64 位表示 + respData
	req := make([]byte, reqLen+ numOfLengthBytes)
	// 第一步：
	// 把长度写进去前八个字节
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	// 第二步：
	// 写入数据
	copy(req[numOfLengthBytes:], data)

	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}

	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return "", err
	}

	// 我响应有多长？
	length := binary.BigEndian.Uint64(lenBs)

	respBs := make([]byte, length)
	_, err = conn.Read(respBs)
	if err != nil {
		return "", err
	}

	return string(respBs), nil
}