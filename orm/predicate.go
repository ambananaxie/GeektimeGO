package orm

// 衍生类型
type op string

const (
	opEQ op = "="
	opLT op = "<"
	opGT op = ">"
	opNot op = "NOT"
	opAnd op = "AND"
	opOr op = "OR"
	opAdd = "+"
	opMulti = "*"
)

func (o op) String() string {
	return  string(o)
}

// 这种叫做别名
// type op=string

type Predicate struct {
	left Expression
	op    op
	right Expression
}

// EQ("id", 12)
// EQ(sub, "id", 12)
// EQ(sub.id, 12)
// EQ("sub.id", 12)
// func EQ(column string, right any) Predicate  {
// 	return Predicate{
// 		Column: column,
// 		Op: "=",
// 		Data: right,
// 	}
// }



// Not(C("name").EQ("Tom"))
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// C("id").EQ(12).And(C("name").EQ("Tom"))
func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

// C("id").EQ(12).Or(C("name").EQ("Tom"))
func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

func (Predicate) expr(){}

type value struct {
	val any
}

func (value) expr(){}



