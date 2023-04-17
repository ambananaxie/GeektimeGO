package dao

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/live/stress_test/user_service/internal/repository/dao/model"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

//go:generate mockgen -source=user.go -destination=mocks/user_mock.gen.go -package=daomocks UserDAO
type UserDAO interface {
	InsertUser(ctx context.Context, u *model.User) error
	UpdateUser(ctx context.Context, u *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserById(ctx context.Context, id uint64) (*model.User, error)
}

type userDAOWithShadow struct {
	db *gorm.DB // xx/user_db
	shadowDB *gorm.DB
}
type orderDAOWithShadow struct {
	db *gorm.DB // xxx/order_db
	shadowDB *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &userDAO{
		db: db,
	}
}

type userDAO struct {
	db *gorm.DB
}

const operationNamePrefix = "dao."

func (dao *userDAO) UpdateUser(ctx context.Context, u *model.User) error {
	return dao.db.WithContext(ctx).UpdateColumn("name", u.Name).Error
}

func (dao *userDAO) GetUserById(ctx context.Context, id uint64) (*model.User, error) {
	var u model.User
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&u).Error
	return &u, err
}

func (dao *userDAO) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := dao.db.WithContext(ctx).Where("email", email).First(&u).Error
	return &u, err
}

func (dao *userDAO) InsertUser(ctx context.Context, u *model.User) error {
	name := operationNamePrefix + "InsertUser"
	span, _ := opentracing.StartSpanFromContext(ctx, name)
	defer span.Finish()
	return dao.db.WithContext(ctx).Create(u).Error
}