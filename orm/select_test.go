package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/valuer"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelector_Join(t *testing.T) {
	db := memoryDB(t)
	type Order struct {
		Id int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId int

		UsingCol1 string
		UsingCol2 string
	}

	type Item struct {
		Id int
	}

	testCases := []struct{
		name string
		s QueryBuilder
		wantQuery *Query
		wantErr error
	} {
		{
			name: "specify table",
			s: NewSelector[Order](db).From(TableOf(&OrderDetail{})),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order_detail`;",
			},
		},
		{
			name: "join-using",
			s: func() QueryBuilder{
				t1 := TableOf(&Order{})
				t2 := TableOf(&OrderDetail{})
				t3 := t1.Join(t2).Using("UsingCol1", "UsingCol2")
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` JOIN `order_detail` USING (`using_col1`,`using_col2`));",
			},
		},
		{
			name: "left join",
			s: func() QueryBuilder{
				t1 := TableOf(&Order{})
				t2 := TableOf(&OrderDetail{})
				t3 := t1.LeftJoin(t2).Using("UsingCol1", "UsingCol2")
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` LEFT JOIN `order_detail` USING (`using_col1`,`using_col2`));",
			},
		},
		{
			name: "right join",
			s: func() QueryBuilder{
				t1 := TableOf(&Order{})
				t2 := TableOf(&OrderDetail{})
				t3 := t1.RightJoin(t2).Using("UsingCol1", "UsingCol2")
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` RIGHT JOIN `order_detail` USING (`using_col1`,`using_col2`));",
			},
		},
		{
			name: "join-on",
			s: func() QueryBuilder{
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").EQ(t2.C("OrderId")))
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`);",
			},
		},
		{
			name: "join table",
			s: func() QueryBuilder{
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").EQ(t2.C("OrderId")))
				t4 := TableOf(&Item{}).As("t4")
				t5 := t3.Join(t4).On(t2.C("ItemId").EQ(t4.C("Id")))
				return NewSelector[Order](db).From(t5)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM " +
					"((`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) " +
					"JOIN `item` AS `t4` ON `t2`.`item_id` = `t4`.`id`);",
			},
		},
		{
			name: "table join ",
			s: func() QueryBuilder{
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").EQ(t2.C("OrderId")))
				t4 := TableOf(&Item{}).As("t4")
				t5 := t4.Join(t3).On(t2.C("ItemId").EQ(t4.C("Id")))
				return NewSelector[Order](db).From(t5)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`item` AS `t4` JOIN (`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) ON `t2`.`item_id` = `t4`.`id`);",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestSelector_Select(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct{
		name string
		s QueryBuilder
		wantQuery *Query
		wantErr error
	} {
		{
			name: "invalid column",
			s: NewSelector[TestModel](db).Select(C("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "multiple columns",
			s: NewSelector[TestModel](db).Select(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "SELECT `first_name`,`last_name` FROM `test_model`;",
			},
		},
		{
			name: "columns alias",
			s: NewSelector[TestModel](db).Select(C("FirstName").As("my_name"), C("LastName")),
			wantQuery: &Query{
				SQL: "SELECT `first_name` AS `my_name`,`last_name` FROM `test_model`;",
			},
		},
		{
			name: "avg",
			s: NewSelector[TestModel](db).Select(Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name: "avg alias",
			s: NewSelector[TestModel](db).Select(Avg("Age").As("avg_age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`) AS `avg_age` FROM `test_model`;",
			},
		},
		{
			name: "sum",
			s: NewSelector[TestModel](db).Select(Sum("Age")),
			wantQuery: &Query{
				SQL: "SELECT SUM(`age`) FROM `test_model`;",
			},
		},
		{
			name: "count",
			s: NewSelector[TestModel](db).Select(Count("Age")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(`age`) FROM `test_model`;",
			},
		},
		{
			name: "max",
			s: NewSelector[TestModel](db).Select(Max("Age")),
			wantQuery: &Query{
				SQL: "SELECT MAX(`age`) FROM `test_model`;",
			},
		},
		{
			name: "min",
			s: NewSelector[TestModel](db).Select(Min("Age")),
			wantQuery: &Query{
				SQL: "SELECT MIN(`age`) FROM `test_model`;",
			},
		},
		{
			name: "aggregate invalid columns",
			s: NewSelector[TestModel](db).Select(Min("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "multiple aggregate",
			s: NewSelector[TestModel](db).Select(Min("Age"), Max("Age")),
			wantQuery: &Query{
				SQL: "SELECT MIN(`age`),MAX(`age`) FROM `test_model`;",
			},
		},
		{
			name: "raw expression",
			s: NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT `first_name`)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT `first_name`) FROM `test_model`;",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct{
		name string

		builder QueryBuilder

		wantQuery *Query
		wantErr error
	}{
		{
			name: "no from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name: "empty where",
			builder:  NewSelector[TestModel](db).Where(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name: "where",
			builder:  NewSelector[TestModel](db).Where(C("Age").EQ(18)),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name: "not",
			builder:  NewSelector[TestModel](db).Where(Not(C("Age").EQ(18))),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name: "and",
			builder:  NewSelector[TestModel](db).Where(C("Age").EQ(18).And(C("FirstName").EQ("Tom"))),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name: "or",
			builder:  NewSelector[TestModel](db).Where(C("Age").EQ(18).Or(C("FirstName").EQ("Tom"))),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name: "invalid column",
			builder:  NewSelector[TestModel](db).Where(C("Age").EQ(18).Or(C("XXXX").EQ("Tom"))),
			wantErr: errs.NewErrUnknownField("XXXX"),
		},

		{
			name: "raw expression as predicate",
			builder:  NewSelector[TestModel](db).Where(Raw("`id`<?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE (`id`<?);",
				Args: []any{18},
			},
		},
		{
			name: "raw expression used in predicate",
			builder:  NewSelector[TestModel](db).Where(C("Id").EQ(Raw("`age`+?", 1))),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE `id` = (`age`+?);",
				Args: []any{1},
			},
		},

		{
			name: "columns alias in where",
			builder:  NewSelector[TestModel](db).Where(C("Id").As("my_id").EQ(18)),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{18},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

type TestModel struct {
	Id        int64
	// ""
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func (TestModel) CreateSQL() string {
	return `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	// 对应于 query error
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))

	// 对应于 no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// data
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)


	// scan error
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("abc", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	testCases := []struct{
		name string
		s *Selector[TestModel]

		wantErr error
		wantRes *TestModel
	} {
		{
			name: "invalid query",
			s: NewSelector[TestModel](db).Where(C("XXX").EQ(1)),
			wantErr: errs.NewErrUnknownField("XXX"),
		},
		{
			name: "query error",
			s: NewSelector[TestModel](db).Where(C("Id").EQ(1)),
			wantErr: errors.New("query error"),
		},
		{
			name: "no rows",
			s: NewSelector[TestModel](db).Where(C("Id").EQ(1)),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			s: NewSelector[TestModel](db).Where(C("Id").EQ(1)),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		// {
		// 	name: "scan error",
		// 	r: NewSelector[TestModel](db).Where(C("Id").EQ(1)),
		// 	wantErr: ErrNoRows,
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func memoryDB(t *testing.T, opts...DBOption) *DB {
	db, err := Open("sqlite3",
		"file:test.db?cache=shared&mode=memory",
		// 仅仅用于单元测试，不会发起真的查询
		opts...)
	require.NoError(t, err)
	return db
}

// 在 orm 目录下执行
// go test -bench=BenchmarkQuerier_Get -benchmem -benchtime=10000x
func BenchmarkQuerier_Get(b *testing.B) {
	db, err := Open("sqlite3", fmt.Sprintf("file:benchmark_get.db?cache=shared&mode=memory"))
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.db.Exec(TestModel{}.CreateSQL())
	if err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`)" +
		"VALUES (?,?,?,?)", 12, "Deng", 18, "Ming")

	if err != nil {
		b.Fatal(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.Run("unsafe", func(b *testing.B) {
		db.creator = valuer.NewUnsafeValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.creator = valuer.NewReflectValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}