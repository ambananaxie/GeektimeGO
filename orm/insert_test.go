package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInserter_SQLite_upsert(t *testing.T) {
	db := memoryDB(t, DBWithDialect(DialectSQLite))
	testCases := []struct{
		name string
		i QueryBuilder

		wantErr error
		wantQuery *Query
	}{
		{
			name: "upsert-update value",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id: 12,
				FirstName: "Tom",
				Age: 18,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().ConflictColumns("Id").Update(Assign("FirstName", "Deng"),
				Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name`=?,`age`=?;",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}, "Deng", 19},
			},
		},
		{
			name: "upsert-update column",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id: 12,
				FirstName: "Tom",
				Age: 18,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id: 13,
				FirstName: "DaMing",
				Age: 19,
				LastName: &sql.NullString{String: "Deng", Valid: true},
			}).OnDuplicateKey().ConflictColumns("FirstName", "LastName").Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?) " +
					"ON CONFLICT(`first_name`,`last_name`) DO UPDATE SET `first_name`=excluded.`first_name`,`age`=excluded.`age`;",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(13), "DaMing", int8(19), &sql.NullString{String: "Deng", Valid: true}},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct{
		name string
		i QueryBuilder

		wantErr error
		wantQuery *Query
	}{
		{
			// 只插入一行
			name: "no row",
			i: NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			// 只插入一行
			name: "single row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id: 12,
				FirstName: "Tom",
				Age: 18,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}},
			},
		},
		{
			// 插入多行
			name: "multiple rows",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id: 12,
				FirstName: "Tom",
				Age: 18,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id: 13,
				FirstName: "DaMing",
				Age: 19,
				LastName: &sql.NullString{String: "Deng", Valid: true},
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(13), "DaMing", int8(19), &sql.NullString{String: "Deng", Valid: true}},
			},
		},

		{
			// 插入多行、部分列
			name: "partial columns",
			i: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id: 12,
				FirstName: "Tom",
				Age: 18,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id: 13,
				FirstName: "DaMing",
				Age: 19,
				LastName: &sql.NullString{String: "Deng", Valid: true},
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`) VALUES (?,?),(?,?);",
				Args: []any{int64(12), "Tom", int64(13), "DaMing"},
			},
		},
		{
			name: "upsert-update value",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id: 12,
				FirstName: "Tom",
				Age: 18,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().Update(Assign("FirstName", "Deng"),
				Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=?,`age`=?;",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}, "Deng", 19},
			},
		},
		{
			name: "upsert-update column",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id: 12,
				FirstName: "Tom",
				Age: 18,
				LastName: &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id: 13,
				FirstName: "DaMing",
				Age: 19,
				LastName: &sql.NullString{String: "Deng", Valid: true},
			}).OnDuplicateKey().Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`age`=VALUES(`age`);",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(13), "DaMing", int8(19), &sql.NullString{String: "Deng", Valid: true}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestInserter_Exec(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	testCases := []struct{
		name string
		i *Inserter[TestModel]
		wantErr error
		affected int64
	}{
		{
			name: "query error",
			i: func() *Inserter[TestModel]{
				return NewInserter[TestModel](db).Values(&TestModel{}).
					Columns("Invalid")
			}(),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "db error",
			i: func() *Inserter[TestModel]{
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(errors.New("db error"))
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			wantErr: errors.New("db error"),
		},
		{
			name: "exec",
			i: func() *Inserter[TestModel]{
				res := driver.RowsAffected(1)
				mock.ExpectExec("INSERT INTO .*").
					WillReturnResult(res)
				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			affected: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.affected, affected)
		})
	}
}
