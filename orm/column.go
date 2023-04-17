package orm

type Column struct {
	table TableReference
	name string
	alias string
}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) assign() {}

func (c Column) As(alias string) Column {
	return Column{
		name: c.name,
		alias: alias,
		table: c.table,
	}
}

// EQ 代表相等
// C("id").EQ(12)
// sub.C("id").EQ(12)
func (c Column) EQ(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: valueOf(arg),
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: valueOf(arg),
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: valueOf(arg),
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{val: val}
	}
}

func (c Column) expr() {}
func (c Column) selectable() {}

func (c Column) Add(delta int) MathExpr {
	return MathExpr{
		left: c,
		op: opAdd,
		right: value{val: delta},
	}
}

func (c Column) Multi(delta int) MathExpr {
	return MathExpr{
		left: c,
		op: opAdd,
		right: value{val: delta},
	}
}