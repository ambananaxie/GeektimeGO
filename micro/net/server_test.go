package net

import (
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/micro/net/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestHandleConn(t *testing.T) {
	testCases := []struct{
		name string

		mock func(ctrl *gomock.Controller) net.Conn
		wantErr error
	} {
		{
			name: "read error",
			mock: func(ctrl *gomock.Controller) net.Conn {
				conn := mocks.NewMockConn(ctrl)
				conn.EXPECT().Read(gomock.Any()).Return(0, errors.New("read error"))
				return conn
			},
			wantErr: errors.New("read error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			err := handleConn(tc.mock(ctrl))
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
