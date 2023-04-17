package gorm

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

type Product struct {
	gorm.Model
	Code  string `gorm:"column(code)"`
	Price uint
}

func  (p Product) TableName() string {
	return "product_t"
}

func (p *Product) BeforeSave(tx *gorm.DB) (err error) {
	// 影子表
	// if tx.Statement.Context.Value("shadow") == "true" {
	// 	tx.Statement.Table = "shadow_product_t"
	// }

	// 假如要在这里进行影子库的分流，怎么分？能不能分？

	println("before save")
	return
}

func (p *Product) AfterSave(tx *gorm.DB) (err error) {
	println("after save")
	return
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	println("before create")
	return
}

func (p *Product) AfterCreate(tx *gorm.DB) (err error) {
	println("after create")
	// 刷新缓存
	return
}

func (p *Product) BeforeUpdate(tx *gorm.DB) (err error) {
	println("before update")
	return
}

func (p *Product) AfterUpdate(tx *gorm.DB) (err error) {
	println("after update")
	// 刷新缓存
	return
}

func (p *Product) BeforeDelete(tx *gorm.DB) (err error) {
	// tx.Statement.Table="123"
	println("before update")
	return
}

func (p *Product) AfterDelete(tx *gorm.DB) (err error) {
	println("after update")
	return
}

func  (p *Product) AfterFind(tx *gorm.DB) (err error) {
	println("after find")
	return
}

func TestCRUD(t *testing.T) {
	liveDB, err := sql.Open("mysql", "")
	require.NoError(t, err)
	shadowDB, err := sql.Open("mysql", "")
	require.NoError(t, err)
	db, err := gorm.Open(sqlite.Open("file:test.db?cache=shared&mode=memory"),
		&gorm.Config{
		// ConnPool: 传入你的各种读写分离、分库分表、影子库影子表的 ConnPool
		ConnPool: &ShadowPool{
			live: liveDB,
			shadow: shadowDB,
		},
		})
	if err != nil {
		panic("failed to connect database")
	}
	db.Debug()

	// Migrate the schema
	db.AutoMigrate(&Product{})

	// Create
	db.Create(&Product{Code: "D42", Price: 100})


	// Read
	var product Product
	db.WithContext(context.WithValue(context.Background(),
		"use_master", "true")).First(&product, 1) // find product with integer primary key
	db.First(&product, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	db.Model(&product).Update("Price", 200)
	// Update - update multiple fields
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	db.Delete(&product, 1)
}
