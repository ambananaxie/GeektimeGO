
package homework_delete

import (
	"reflect"
)

// Deleter builds DELETE query
type Deleter[T any] struct {
	builder
	table string
	where []Predicate
}

// Build returns DELETE query
func (d *Deleter[T]) Build() (*Query, error) {
	_, _ = d.sb.WriteString("DELETE FROM ")

	if d.table == "" {
		var t T
		d.sb.WriteByte('`')
		d.sb.WriteString(reflect.TypeOf(t).Name())
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.table)
	}
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		err := d.buildPredicates(d.where)
		if err != nil {
			return nil, err
		}
	}
	d.sb.WriteByte(';')
	return &Query{SQL: d.sb.String(), Args: d.args}, nil
}

// From accepts model definition
func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

// Where accepts predicates
func (d *Deleter[T]) Where(predicates ...Predicate) *Deleter[T] {
	d.where = predicates
	return d
}
