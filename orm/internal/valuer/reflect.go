package valuer

import (
	"database/sql"
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"gitee.com/geektime-geekbang/geektime-go/orm/model"
	"reflect"
)

type reflectValue struct {
	model *model.Model
	// 对应于 T 的指针
	// val any
	val reflect.Value
}

var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return reflectValue{
		model: model,
		val: reflect.ValueOf(val).Elem(),
	}
}
func (r reflectValue) Field(name string) (any, error) {
	// 检测 name 是否合法
	// _, ok := r.val.Type().FieldByName(name)
	// if !ok {
	// 	// 报错
	// }
	val := r.val.FieldByName(name)
	// if val == (reflect.Value{}) {
	// 	// 报错
	// }

	return val.Interface(), nil
}

func (r reflectValue) SetColumns(rows *sql.Rows) error {
	// 在这里，继续处理结果集

	// 我怎么知道你 SELECT 出来了哪些列？
	// 拿到了 SELECT 出来的列
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	// 怎么利用 cs 来解决顺序问题和类型问题


	// 通过 cs 来构造 vals
	// 怎么构造呢？
	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		// c 是列名
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 反射创建一个实例
		// 这里创建的实例是原本类型的指针类型
		// 例如 fd.Type = int，那么val 是 *int
		val := reflect.New(fd.Type)
		vals = append(vals, val.Interface())
		// 记得要调用 Elem，因为 fd.Type = int，那么val 是 *int
		valElems = append(valElems, val.Elem())
	}

	// 第一个问题：类型要匹配
	// 第二个问题：顺序要匹配

	// SELECT id, first_name, age, last_name
	// SELECT first_name, age, last_name, id
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	// 想办法把 vals 塞进去 结果 tp 里面
	tpValueElem := r.val
	for i, c := range cs {
		// c 是列名
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		tpValueElem.FieldByName(fd.GoName).
			Set(valElems[i])
	}

	return nil
}

