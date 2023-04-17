package orm

import (
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
	"strings"
)

type builder struct {
	core
	sb strings.Builder
	args []any
	quoter byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}


func (b *builder) buildColumn(c Column) error {
	switch table := c.table.(type) {
	case nil:
		fd, ok := b.model.FieldMap[c.name]
		// 字段不对，或者说列不对
		if !ok {
			return errs.NewErrUnknownField(c.name)
		}
		b.quote(fd.ColName)
		if c.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(c.alias)
		}
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[c.name]
		if !ok {
			return errs.NewErrUnknownField(c.name)
		}
		if table.alias != "" {
			b.quote(table.alias)
			b.sb.WriteByte('.')
		}
		b.quote(fd.ColName)
		if c.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(c.alias)
		}
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

func (b *builder) addArg(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, vals...)
}

func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

func (b *builder) buildExpression(expr Expression) error {
	switch exp := expr.(type){
	case nil:
		// 这是重构前处理 Predicate 的代码
	// case Predicate:
	// 	// 在这里处理 p
	// 	// p.left 构建好
	// 	// p.op 构建好
	// 	// p.right 构建好
	// 	_, ok := exp.left.(Predicate)
	// 	if ok {
	// 		b.sb.WriteByte('(')
	// 	}
	// 	if err := b.buildExpression(exp.left); err != nil {
	// 		return err
	// 	}
	// 	if ok {
	// 		b.sb.WriteByte(')')
	// 	}
	//
	// 	if exp.op != "" {
	// 		b.sb.WriteByte(' ')
	// 		b.sb.WriteString(exp.op.String())
	// 		b.sb.WriteByte(' ')
	// 	}
	// 	_, ok = exp.right.(Predicate)
	// 	if ok {
	// 		b.sb.WriteByte('(')
	// 	}
	// 	if err := b.buildExpression(exp.right); err != nil {
	// 		return err
	// 	}
	// 	if ok {
	// 		b.sb.WriteByte(')')
	// 	}
	case Column:
		// 这种写法很隐晦
		exp.alias = ""
		return b.buildColumn(exp)
	case value:
		b.sb.WriteByte('?')
		b.addArg(exp.val)
	case RawExpr:
		b.sb.WriteByte('(')
		b.sb.WriteString(exp.raw)
		b.addArg(exp.args...)
		b.sb.WriteByte(')')
	case MathExpr:
		return b.buildBinaryExpr(binaryExpr(exp))
	case Predicate:
		return b.buildBinaryExpr(binaryExpr(exp))
	case binaryExpr:
		return b.buildBinaryExpr(exp)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (b *builder) buildBinaryExpr(e binaryExpr) error {
	err := b.buildSubExpr(e.left)
	if err != nil {
		return err
	}
	if e.op != "" {
		b.sb.WriteByte(' ')
		b.sb.WriteString(e.op.String())
	}
	if e.right != nil {
		b.sb.WriteByte(' ')
		return b.buildSubExpr(e.right)
	}
	return nil
}

func (b *builder) buildSubExpr(subExpr Expression) error {
	switch sub := subExpr.(type) {
	case MathExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case binaryExpr:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(sub); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	case Predicate:
		_ = b.sb.WriteByte('(')
		if err := b.buildBinaryExpr(binaryExpr(sub)); err != nil {
			return err
		}
		_ = b.sb.WriteByte(')')
	default:
		if err := b.buildExpression(sub); err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) reset() {
	b.sb.Reset()
	b.args = nil
}