
package orm

import (
	"context"
)

type Updater[T any] struct {
}

func NewUpdater[T any](db *DB) *Updater[T] {
	panic("implement me")
}

func (u *Updater[T]) Update(t *T) *Updater[T] {
	panic("implement me")
}

func (u *Updater[T]) Set(assigns ...Assignable) *Updater[T] {
	panic("implement me")
}

func (u *Updater[T]) Build() (*Query, error) {
	panic("implement me")
}


func (u *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	panic("implement me")
}

func (u *Updater[T]) Exec(ctx context.Context) Result {
	panic("implement me")
}

func AssignNotNilColumns(entity interface{}) []Assignable {
	panic("implement me")
}

func AssignNotZeroColumns(entity interface{}) []Assignable {
	panic("implement me")
}
