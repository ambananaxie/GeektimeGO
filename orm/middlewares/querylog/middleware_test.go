package querylog

import (
	"context"
	"database/sql"
	"gitee.com/geektime-geekbang/geektime-go/orm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	var query string
	var args []any
	m := (&MiddlewareBuilder{}).LogFunc(func(q string, as []any) {
		query=q
		args =as
	})

	db, err := orm.Open("sqlite3",
		"file:test.db?cache=shared&mode=memory",
		orm.DBWithMiddlewares(m.Build()))
	require.NoError(t, err)
	_, _ = orm.NewSelector[TestModel](db).Where(orm.C("Id").EQ(10)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{10}, args)

	orm.NewInserter[TestModel](db).Values(&TestModel{Id: 18}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);", query)
	assert.Equal(t, []any{int64(18), "", int8(0), (*sql.NullString)(nil)}, args)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
