package homework_delete

type Deleter[T any] struct {
}

func (d *Deleter[T]) Build() (*Query, error) {
	panic("implement me")
}

// From accepts model definition
func (d *Deleter[T]) From(table string) *Deleter[T] {
	panic("implement me ")
}

// Where accepts predicates
func (d *Deleter[T]) Where(predicates ...Predicate) *Deleter[T] {
	panic("implement me")
}
