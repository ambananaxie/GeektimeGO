package valuer

import (
	"database/sql/driver"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"testing"
)

func BenchmarkSetColumns(b *testing.B) {

	fn := func(b *testing.B, creator Creator) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(b, err)
		defer mockDB.Close()

		// 我们需要跑 N 次，也就是要准备 N 行
		mockRows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
		row := []driver.Value{"1", "Tom", "18", "Jerry"}
		for i := 0; i < b.N; i++ {
			mockRows.AddRow(row...)
		}
		mock.ExpectQuery("SELECT XX").WillReturnRows(mockRows)

		rows, err := mockDB.Query("SELECT XX")

		r := model.NewRegistry()
		m, err := r.Get(&TestModel{})
		require.NoError(b, err)

		// 重置计时器
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows.Next()
			val := creator(m, &TestModel{})
			_ = val.SetColumns(rows)
		}
	}

	b.Run("reflect", func(b *testing.B) {
		fn(b, NewReflectValue)
	})

	b.Run("unsafe", func(b *testing.B) {
		fn(b, NewUnsafeValue)
	})
}


