package sso

import (
	"github.com/google/uuid"
	"strings"
)

// Encrypt 模拟生成签名
func Encrypt(clintId string) string {
	// 这里你要挑选加密算法
	random := uuid.New().String()
	return clintId + ":" + random
}

// Decrypt 模拟解密
func Decrypt(signature []byte) (clientId string, err error) {
	seg := strings.Split(string(signature), ":")
	return seg[0], nil
}