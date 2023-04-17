package orm

type Assignment struct {
	col string
	// val any
	// 在 UPDATE 里面我们改成了这个 Expression 的结构
	val Expression
}

func (Assignment) assign() {}

func Assign(col string, val any) Assignment {
	return Assignment{
		col: col,
		val: valueOf(val),
	}
}
