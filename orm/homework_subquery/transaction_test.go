// Copyright 2021 gotomicro
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTx_Commit(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		mock.ExpectClose()
		_ = db.Close()
	}()

	// 事务正常提交
	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)


}

func TestTx_Rollback(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mockDB.Close() }()

	db, err := OpenDB(mockDB)
	if err != nil {
		t.Fatal(err)
	}

	// 事务回滚
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	assert.Nil(t, err)
	err = tx.Rollback()
	assert.Nil(t, err)
}

func TestDBWithMiddleware(t *testing.T) {
	var res []byte
	var mdl1 Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, qc *QueryContext) *QueryResult {
			res = append(res, '1')
			return next(ctx, qc)
		}
	}
	var mdl2 Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, qc *QueryContext) *QueryResult {
			res = append(res, '2')
			return next(ctx, qc)
		}
	}

	var mdl3 Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, qc *QueryContext) *QueryResult {
			res = append(res, '3')
			return next(ctx, qc)
		}
	}
	var last Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, qc *QueryContext) *QueryResult {
			return &QueryResult{
				Err: errors.New("mock error"),
			}
		}
	}

	db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory",
		DBWithMiddleware(mdl1, mdl2, mdl3, last))
	require.NoError(t, err)

	_, err = NewSelector[TestModel](db).Get(context.Background())
	assert.Equal(t, errors.New("mock error"), err)
	assert.Equal(t, "123" ,string(res))
}