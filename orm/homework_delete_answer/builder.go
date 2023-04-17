package homework_delete

import (
	"fmt"
	"strings"
)

type builder struct {
	sb    strings.Builder
	args  []any
}

// type Predicates []Predicate
//
// func (ps Predicates) build(s *strings.Builder) error {
// 	// 写在这里
// }

// type predicates struct {
// 	// WHERE 或者 HAVING
// 	prefix string
// 	ps []Predicate
// }

// func (ps predicates) build(s *strings.Builder) error {
//  包含拼接 WHERE 或者 HAVING 的部分
// 	// 写在这里
// }

func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	if err := b.buildExpression(p); err != nil {
		return err
	}
	return nil
}

func (b *builder) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		b.sb.WriteByte('`')
		b.sb.WriteString(exp.name)
		b.sb.WriteByte('`')
	case value:
		b.sb.WriteByte('?')
		b.args = append(b.args, exp.val)
	case Predicate:
		_, lp := exp.left.(Predicate)
		if lp {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.left); err != nil {
			return err
		}
		if lp {
			b.sb.WriteByte(')')
		}

		b.sb.WriteByte(' ')
		b.sb.WriteString(exp.op.String())
		b.sb.WriteByte(' ')

		_, rp := exp.right.(Predicate)
		if rp {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.right); err != nil {
			return err
		}
		if rp {
			b.sb.WriteByte(')')
		}
	default:
		return fmt.Errorf("orm: 不支持的表达式 %v", exp)
	}
	return nil
}
