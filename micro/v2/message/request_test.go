package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	testCases := []struct{
		name string
		req *Request
	} {
		{
			name: "normal",
			req: &Request{
				RequestID: 123,
				Version: 12,
				Compresser: 13,
				Serializer: 14,
				ServiceName: "user-service",
				MethodName: "GetById",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b": "a",
				},
				Data: []byte("hello, world"),
			},
		},

		{
			name: "data with \n ",
			req: &Request{
				RequestID: 123,
				Version: 12,
				Compresser: 13,
				Serializer: 14,
				ServiceName: "user-service",
				MethodName: "GetById",
				Data: []byte("hello \n world"),
			},
		},

		// 你可以禁止开发者（框架使用者）在 meta 里面使用 \n 和 \r，所以不会出现这种情况
		//{
		//	name: "data with \n ",
		//	resp: &Request{
		//		RequestID: 123,
		//		Version: 12,
		//		Compresser: 13,
		//		Serializer: 14,
		//		ServiceName: "user-\nservice",
		//		MethodName: "GetById",
		//		Data: []byte("hello \n world"),
		//	},
		//},

		{
			name: "no meta",
			req: &Request{
				RequestID: 123,
				Version: 12,
				Compresser: 13,
				Serializer: 14,
				ServiceName: "user-service",
				MethodName: "GetById",
			},
		},

		{
			name: "no meta with data",
			req: &Request{
				RequestID: 123,
				Version: 12,
				Compresser: 13,
				Serializer: 14,
				ServiceName: "user-service",
				MethodName: "GetById",
				Data: []byte("hello, world"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.CalculateHeaderLength()
			tc.req.CalculateBodyLength()
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})
	}
}

