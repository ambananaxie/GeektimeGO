
package orm

import (
	"context"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"reflect"
)

type Updater[T any] struct {
	builder
	sess      Session
	assigns []Assignable
	val     *T
	where   []Predicate
}

func NewUpdater[T any](sess Session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		builder: builder{
			core: c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (u *Updater[T]) Update(t *T) *Updater[T] {
	u.val = t
	return u
}

func (u *Updater[T]) Set(assigns ...Assignable) *Updater[T] {
	u.assigns = assigns
	return u
}

func (u *Updater[T]) Build() (*Query, error) {
	defer func() {
		u.reset()
	}()
	if len(u.assigns) == 0 {
		return nil, errs.ErrNoUpdatedColumns
	}
	var (
		err error
		t   T
	)
	u.model, err = u.r.Get(&t)
	if err != nil {
		return nil, err
	}
	u.sb.WriteString("UPDATE ")
	u.quote(u.model.TableName)
	u.sb.WriteString(" SET ")
	val := u.creator(u.model, u.val)
	for i, a := range u.assigns {
		if i > 0 {
			u.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			arg, err := val.Field(assign.name)
			if err != nil {
				return nil, err
			}
			if err = u.buildColumn(Column{name: assign.name}); err != nil {
				return nil, err
			}
			u.sb.WriteString("=?")

			u.addArg(arg)
		case Assignment:
			if err = u.buildAssignment(assign); err != nil {
				return nil, err
			}
		default:
			return nil, errs.NewErrUnsupportedAssignableType(a)
		}
	}
	if len(u.where) > 0 {
		u.sb.WriteString(" WHERE ")
		if err = u.buildPredicates(u.where); err != nil {
			return nil, err
		}
	}
	u.sb.WriteByte(';')
	return &Query{
		SQL:  u.sb.String(),
		Args: u.args,
	}, nil
}



func (u *Updater[T]) buildAssignment(assign Assignment) error {
	if err := u.buildColumn(Column{name: assign.col}); err != nil {
		return err
	}
	u.sb.WriteByte('=')
	return u.buildExpression(assign.val)
}

func (u *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	u.where = ps
	return u
}

func (u *Updater[T]) Exec(ctx context.Context) Result {
	q, err := u.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := u.sess.execContext(ctx, q.SQL, q.Args...)
	return Result{err: err, res: res}
}

func AssignNotNilColumns(entity interface{}) []Assignable {
	return AssignColumns(entity, func(typ reflect.StructField, val reflect.Value) bool {
		switch val.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
			return !val.IsNil()
		}
		return true
	})
}

func AssignNotZeroColumns(entity interface{}) []Assignable {
	return AssignColumns(entity, func(typ reflect.StructField, val reflect.Value) bool {
		return !val.IsZero()
	})
}

func AssignColumns(entity interface{}, filter func(typ reflect.StructField, val reflect.Value) bool) []Assignable {
	val := reflect.ValueOf(entity).Elem()
	typ := reflect.TypeOf(entity).Elem()
	numField := val.NumField()
	res := make([]Assignable, 0, numField)
	for i := 0; i < numField; i++ {
		fieldVal := val.Field(i)
		fieldTyp := typ.Field(i)
		if filter(fieldTyp, fieldVal) {
			res = append(res, Assign(fieldTyp.Name, fieldVal.Interface()))
		}
	}
	return res
}
