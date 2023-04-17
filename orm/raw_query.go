package orm

import (
	"context"
	"database/sql"
)

type RawQuerier[T any] struct {
	core
	sess Session
	sql string
	args []any
}

func (r RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL: r.sql,
		Args: r.args,
	}, nil
}

func RawQuery[T any](sess Session, query string, args...any) *RawQuerier[T] {
	 c := sess.getCore()
	return &RawQuerier[T]{
		sql: query,
		args: args,
		sess: sess,
		core:c,
	}
}

func (i RawQuerier[T]) Exec(ctx context.Context) Result {
	var err error
	i.model, err = i.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, i.sess, i.core, &QueryContext{
		Type: "RAW",
		Builder: i,
		Model: i.model,
	} )
	// var t *T
	// if val, ok := res.Result.(*T); ok {
	// 	t = val
	// }
	// return t, res.Err
	var sqlRes sql.Result
	if res.Result != nil {
		sqlRes =  res.Result.(sql.Result)
	}
	return Result{
		err: res.Err,
		res: sqlRes,
	}
}

func (s *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Type: "RAW",
		Builder: s,
		Model: s.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}



func (r RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// TODO implement me
	panic("implement me")
}

