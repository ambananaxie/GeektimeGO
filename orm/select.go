package orm

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
)

// Selectable 是一个标记接口
// 它代表的是查找的列，或者聚合函数等
// SELECT XXX 部分
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	table TableReference
	where []Predicate
	columns []Selectable

	sess Session
}

// func (db *DB) NewSelector[T any]()*Selector[T] {
// 	return &Selector[T]{
// 		sb: &strings.Builder{},
// 		db: db,
// 	}
// }

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core: c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

// func (r *Selector[T]) Demo[S any]() (*Query, error) {
// 	panic("implement me")
// }

func (s *Selector[T]) Build() (*Query, error) {
	defer func() {
		s.reset()
	}()
	if s.model == nil {
		var err error
		s.model, err = s.r.Get(new(T))
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteString("SELECT ")

	if err := s.buildColumns(); err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")

	if err := s.buildTable(s.table); err != nil {
		return nil, err
	}

	// 我怎么把表名拿到
	//if s.table == "" {
	//	s.sb.WriteByte('`')
	//	s.sb.WriteString(s.model.TableName)
	//	s.sb.WriteByte('`')
	//} else {
	//	// segs := strings.Split(r.table, ".")
	//	// sb.WriteByte('`')
	//	// sb.WriteString(segs[0])
	//	// sb.WriteByte('`')
	//	// sb.WriteByte('.')
	//	// sb.WriteByte('`')
	//	// sb.WriteString(segs[1])
	//	// sb.WriteByte('`')
	//	s.sb.WriteString(s.table)
	//}
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL: s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildTable(table TableReference) error {
	switch t := table.(type) {
	case nil:
		// 这是代表完全没有调用 FROM，也就是最普通的形态
		s.quote(s.model.TableName)
	case Table:
		// 这个地方是拿到指定的表的元数据
		m, err := s.r.Get(t.entity)
		if err != nil {
			return err
		}
		s.quote(m.TableName)
		if t.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(t.alias)
		}
	case Join:
		s.sb.WriteByte('(')
		// 构造左边
		err := s.buildTable(t.left)
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(t.typ)
		s.sb.WriteByte(' ')
		// 构造右边
		err = s.buildTable(t.right)
		if err != nil {
			return err
		}

		if len(t.using) >0 {
			s.sb.WriteString(" USING (")
			// 拼接 USING (xx, xx)
			for i, col := range t.using {
				if i > 0 {
					s.sb.WriteByte(',')
				}
				err = s.buildColumn(Column{name: col})
				if err != nil {
					return err
				}
			}
			s.sb.WriteByte(')')
		}

		if len(t.on) > 0 {
			s.sb.WriteString(" ON ")
			p := t.on[0]
			for i := 1; i < len(t.on); i++ {
				p = p.And(t.on[i])
			}
			if err = s.buildExpression(p); err != nil {
				return err
			}
		}

		s.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

// 我们挪上去 builder 那里
// func (s *Selector[T]) buildExpression(expr Expression) error {
// 	switch exp := expr.(type){
// 	case nil:
// 	case Predicate:
// 		// 在这里处理 p
// 		// p.left 构建好
// 		// p.op 构建好
// 		// p.right 构建好
// 		_, ok := exp.left.(Predicate)
// 		if ok {
// 			s.sb.WriteByte('(')
// 		}
// 		if err := s.buildExpression(exp.left); err != nil {
// 			return err
// 		}
// 		if ok {
// 			s.sb.WriteByte(')')
// 		}
//
// 		if exp.op != "" {
// 			s.sb.WriteByte(' ')
// 			s.sb.WriteString(exp.op.String())
// 			s.sb.WriteByte(' ')
// 		}
// 		_, ok = exp.right.(Predicate)
// 		if ok {
// 			s.sb.WriteByte('(')
// 		}
// 		if err := s.buildExpression(exp.right); err != nil {
// 			return err
// 		}
// 		if ok {
// 			s.sb.WriteByte(')')
// 		}
// 	case Column:
// 		// 这种写法很隐晦
// 		exp.alias = ""
// 		return s.buildColumn(exp)
// 	case value:
// 		s.sb.WriteByte('?')
// 		s.addArg(exp.val)
// 	case RawExpr:
// 		s.sb.WriteByte('(')
// 		s.sb.WriteString(exp.raw)
// 		s.addArg(exp.args...)
// 		s.sb.WriteByte(')')
// 	default:
// 		return errs.NewErrUnsupportedExpression(expr)
// 	}
// 	return nil
// }

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		// 没有指定列
		s.sb.WriteByte('*')
		return nil
	}

	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c)
			if err != nil {
				return err
			}
		case Aggregate:
			// 聚合函数名
			s.sb.WriteString(c.fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(Column{name: c.arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
			// 聚合函数本身的别名
			if c.alias != "" {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(c.alias)
				s.sb.WriteByte('`')
			}
		case RawExpr:
			s.sb.WriteString(c.raw)
			s.addArg(c.args...)
		}
	}

	return nil
}

func (s *Selector[T]) addArg(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, vals...)
}

// 这种也是可行
// r.Select("first_name,last_name")
// func (r *Selector[T]) SelectV1(cols string) *Selector[T] {
// 	return r
// }
//
// func (r *Selector[T]) Select(cols...string) *Selector[T] {
// 	r.columns = cols
// 	return r
// }

func (s *Selector[T]) Select(cols...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) From(table TableReference) *Selector[T] {
	s.table = table
	return s
}

// ids:=[]int{1, 2, 3,}
// r.Where("id in (?, ?, ?)", ids)
// r.Where("id in (?, ?, ?)", ids...)
// func (r *Selector[T]) Where(query string, args...any) *Selector[T] {
//
// }

func (s *Selector[T]) Where(ps...Predicate) *Selector[T] {
	s.where = ps
	return s
}

// func (r *Selector[T]) GetV1(ctx context.Context) (*T, error) {
// 	q, err := r.Build()
// 	// 这个是构造 SQL 失败
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	db := r.db.db
// 	// 在这里，就是要发起查询，并且处理结果集
// 	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
// 	// 这个是查询错误
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// 你要确认有没有数据
// 	if !rows.Next() {
// 		// 要不要返回 error？
// 		// 返回 error，和 sql 包语义保持一致
// 		return nil, ErrNoRows
// 	}
//
// 	// 在这里，继续处理结果集
//
// 	// 我怎么知道你 SELECT 出来了哪些列？
// 	// 拿到了 SELECT 出来的列
// 	cs, err := rows.Columns()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var vals []any
// 	tp := new(T)
// 	// 起始地址
// 	address := reflect.ValueOf(tp).UnsafePointer()
// 	for _, c := range cs {
// 		// c 是列名
// 		fd, ok := r.model.ColumnMap[c]
// 		if !ok {
// 			return nil, errs.NewErrUnknownColumn(c)
// 		}
// 		// 是不是要计算字段的地址？
// 		// 起始地址 + 偏移量
// 		fdAddress := unsafe.Pointer(uintptr(address) + fd.Offset)
//
// 		// 反射在特定的地址上，创建一个特定类型的实例
// 		// 这里创建的实例是原本类型的指针类型
// 		// 例如 fd.Type = int，那么val 是 *int
// 		val := reflect.NewAt(fd.Type, fdAddress)
// 		vals = append(vals, val.Interface())
// 	}
//
// 	err = rows.Scan(vals...)
// 	return tp, err
// }

//func (s *Selector) GetFloat64(ctx context.Context) (float64, error) {
//
//}
//
//func (s *Selector[T]) Take(ctx context.Context, receiver any) (*T, error) {
//	switch rec := receiver.(type) {
//	case *float64:
//		*rec = 12.3
//	}
//}
//
//func (s *Selector) GetSum(ctx context.Context) (float64, error) {
//
//}
//
//func (s *Selector) GetAsValue(ctx context.Context) (StringValue, error) {
//
//}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Type: "SELECT",
		Builder: s,
		Model: s.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

// func (r *Selector[T]) Get(ctx context.Context) (*T, error) {
// 	var err error
// 	r.model, err = r.r.Get(new(T))
// 	if err != nil {
// 		return nil, err
// 	}
// 	root := r.getHandler
// 	for i := len(r.mdls) - 1; i >= 0; i -- {
// 		root=r.mdls[i](root)
// 	}
// 	res := root(ctx, &QueryContext{
// 		Type: "SELECT",
// 		Builder: r,
// 		Model: r.model,
// 	})
	// var t *T
	// if val, ok := res.Result.(*T); ok {
	// 	t = val
	// }
	// return t, res.Err
// 	if res.Result != nil {
// 		return res.Result.(*T), res.Err
// 	}
// 	return nil, res.Err
// }

// func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
// 	q, err := qc.Builder.Build()
// 	// 这个是构造 SQL 失败
// 	if err != nil {
// 		return &QueryResult{
// 			Err: err,
// 		}
// 	}
//
// 	// 在这里，就是要发起查询，并且处理结果集
// 	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
// 	// 这个是查询错误
// 	if err != nil {
// 		return &QueryResult{
// 			Err: err,
// 		}
// 	}
//
// 	// 你要确认有没有数据
// 	if !rows.Next() {
// 		// 要不要返回 error？
// 		// 返回 error，和 sql 包语义保持一致
// 		return &QueryResult{
// 			Err: ErrNoRows,
// 		}
// 	}
//
// 	// if flag {
// 	// 	val := valuer.NewReflectValue()
// 	// } else {
// 	// 	val := valuer.NewUnsafeValue()
// 	// }
// 	//
// 	tp := new(T)
// 	val := c.creator(c.model, tp)
// 	err = val.SetColumns(rows)
//
// 	// 接口定义好之后，就两件事，一个是用新接口的方法改造上层，
// 	// 一个就是提供不同的实现
// 	return &QueryResult{
// 		Err: err,
// 		Result: tp,
// 	}
// }

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("implement me")
}

// type Orm struct {
// 	orm.Ormer
// }
//
// func (o *Orm) Insert(md interface{}) (int64, error) {
// 	// 在这里做一些事情
// 	res, err := o.Ormer.Insert(md)
// 	// 在这里再做一些事情
// 	return res, err
// }