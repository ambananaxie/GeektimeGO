package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/valuer"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
	"log"
)

type DBOption func(db *DB)

// DB 是一个 sql.DB 的装饰器
type DB struct {
	core
	db *sql.DB
}

func Open(driver string, dataSourceName string, opts...DBOption) (*DB, error) {
	db, err:= sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts...DBOption) (*DB, error) {
	res := &DB{
		core: core{
			r:  model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: DialectMySQL,
		},
		db: db,

	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// func (db *DB) Begin() (*Tx, error){
// 	tx, err := db.db.Begin()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &Tx{tx: tx}, nil
// }

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error){
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

type txKey struct {}

// ctx, tx, err := db.BeginTxV2()
// doSomething(ctx, tx)
func (db *DB) BeginTxV2(ctx context.Context,
	opts *sql.TxOptions) (context.Context, *Tx, error){
	val := ctx.Value(txKey{})
	tx, ok := val.(*Tx)
	// 存在一个事务，并且这个事务没有被提交或者回滚
	if ok && !tx.done{
		return ctx, tx, nil
	}
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)
	return ctx, tx, nil
}

// 要求前面的人一定要开好事务
// func (db *DB) BeginTxV3(ctx context.Context,
// 	opts *sql.TxOptions) (*Tx, error){
// 	val := ctx.Value(txKey{})
// 	tx, ok := val.(*Tx)
// 	if ok {
// 		return tx, nil
// 	}
// 	return nil, errors.New("没有开事务")
// }

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) DoTx(ctx context.Context,
	fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollbackTx(err, e, panicked)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return err
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}

func DBWithMiddlewares(mdls...Middleware) DBOption {
	return func(db *DB) {
		db.mdls = mdls
		// db.mdls = append(db.mdls, mdls...)
	}
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

func MustOpen(driver string, dataSourceName string, opts...DBOption) *DB {
	res, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}

func (db *DB) Wait() error {
	err := db.db.Ping()
	for err == driver.ErrBadConn {
		log.Println("数据库启动中")
		err = db.db.Ping()
	}
	return nil
}

