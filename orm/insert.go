package orm

import (
	"context"
	"database/sql"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
)

type UpsertBuilder[T any] struct {
	i *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	assigns []Assignable
	conflictColumns []string
}

// ConflictColumns 这是一个中间方法
func (o *UpsertBuilder[T]) ConflictColumns(cols...string) *UpsertBuilder[T]{
	o.conflictColumns = cols
	return o
}

func (o *UpsertBuilder[T]) Update(assigns...Assignable) *Inserter[T]{
	o.i.onDuplicateKey = &Upsert{
		assigns: assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	builder
	sess Session
	values []*T
	columns []string

	// onDuplicateKey []Assignable
	onDuplicateKey *Upsert
}

func NewInserter[T any](sess Session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		builder: builder{
			core: c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

// func (i *Inserter[T]) Upsert(assigns...Assignable) *Inserter[T] {
// 	i.onDuplicateKey = assigns
// 	return i
// }

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}


// Values 指定插入的数据
func (i *Inserter[T]) Values(vals...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Columns(cols...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	defer func() {
		i.reset()
	}()
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	i.sb.WriteString("INSERT INTO ")
	if i.model == nil {
		m, err := i.r.Get(i.values[0])
		i.model = m
		if err != nil {
			return nil, err
		}
	}

	// 拼接表名
	i.quote(i.model.TableName)
	// 一定要显示指定列的顺序，不然我们不知道数据库中默认的顺序
	// 我们要构造 `test_model`(col1, col2...)
	i.sb.WriteByte('(')

	fields := i.model.Fields
	// 用户指定了
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := i.model.FieldMap[fd]
			// 传入了乱七八糟的列
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}

	// 不能遍历这个 FieldMap，ColMap，因为在 Go 里面 map 的遍历，每一次的顺序都不一样
	// 所以额外引入一个 Fields

	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(field.ColName)
	}
	i.sb.WriteByte(')')

	// 拼接 Values
	i.sb.WriteString(" VALUES ")
	// 预估的参数数量是：我有多少行乘以我有多少个字段
	i.args = make([]any, 0, len(i.values) * len(fields))
	for j, v := range i.values {
		if j >0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		val := i.creator(i.model, v)
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			// 把参数读出来
			arg, err := val.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArg(arg)
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicateKey != nil {
		err := i.dialect.buildUpsert(&i.builder, i.onDuplicateKey)
		if err != nil {
			return nil, err
		}
	}
	i.sb.WriteByte(';')
	return &Query{SQL: i.sb.String(), Args: i.args}, nil
}

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	var err error
	i.model, err = i.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, i.sess, i.core, &QueryContext{
		Type: "INSERT",
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

// func (i *Inserter[T]) execHandler(ctx context.Context, qc *QueryContext) *QueryResult {
// 	q, err := i.Build()
// 	if err != nil {
// 		return &QueryResult{
// 			Err: err,
// 			Result: Result{
// 				err: err,
// 			},
// 		}
// 	}
// 	res, err := i.sess.execContext(ctx, q.SQL, q.Args...)
// 	return &QueryResult{
// 		Err: err,
// 		Result: Result{
// 			err: err,
// 			res: res,
// 		},
// 	}
// }


// type MySQLInserter struct {
//
// }
//
// type PostgreSQLInserter[T any] struct {
// 	Inserter[T]
// }