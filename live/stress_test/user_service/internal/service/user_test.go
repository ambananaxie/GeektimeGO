package service

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/cache"
	userapi "gitee.com/geektime-geekbang/geektime-go/live/stress_test/api/user/gen"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository/dao"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository/dao/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
)

// 你可以考虑传入一个 repo 的 mock 对象来执行单元测试
type UserServiceTestSuite struct {
	suite.Suite
	db *gorm.DB
	us userapi.UserServiceServer
}

func (s *UserServiceTestSuite) SetupSuite() {
	c := cache.NewBuildInMapCache(time.Second)
	db, err := gorm.Open(sqlite.Open("file:user_app.db?cache=shared&mode=memory"), &gorm.Config{})
	require.NoError(s.T(), err)
	s.db = db
	err = db.AutoMigrate(&model.User{})
	require.NoError(s.T(), err)
	d := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(d, c)
	s.us = NewUserService(repo)
}

func (s *UserServiceTestSuite) TearDownTest() {
	s.db.Exec("truncate table `users`;")
}

func (s *UserServiceTestSuite) TestCreateUser() {
	t := s.T()
	testCases := []struct{
		name string
		req *userapi.CreateUserReq
		wantResp *userapi.CreateUserResp
		wantErr error
		after func(t *testing.T)
	} {
		{
			name: "created",
			req: &userapi.CreateUserReq{
				User: &userapi.User{
					Name: "DaMing",
					Avatar: "这是我的头像",
					Email: "abc@demo.com",
					Password: "12345678",
				},
			},
			wantResp: &userapi.CreateUserResp{
				User: &userapi.User{
					Id: 1,
					Name: "DaMing",
					Avatar: "这是我的头像",
					Email: "abc@demo.com",
					Password: "12345678",
				},
			},
			after: func(t *testing.T) {

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := s.us.CreateUser(context.Background(), tc.req)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}