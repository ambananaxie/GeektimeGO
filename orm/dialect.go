package orm

import (
	"gitee.com/geektime-geekbang/geektime-go/orm/internal/errs"
)

var (
	DialectMySQL Dialect = mysqlDialect{}
	DialectSQLite Dialect = sqliteDialect{}
	DialectPostgreSQL Dialect = postgreDialect{}
)

type Dialect interface {
	// quoter 就是为了解决引号问题
	// MySQL `
	quoter() byte

	buildUpsert(b *builder, upsert *Upsert) error
}

type standardSQL struct {

}

func (s standardSQL) quoter() byte {
	// TODO implement me
	panic("implement me")
}

func (s standardSQL) buildUpsert(b *builder, upsert *Upsert) error {
	// TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) quoter() byte {
	return '`'
}

func (s mysqlDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range upsert.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			// 字段不对，或者说列不对
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=")
			if err := b.buildExpression(a.val); err != nil {
				return err
			}
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			// 字段不对，或者说列不对
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}

type sqliteDialect struct {
	standardSQL
}

func (s sqliteDialect) quoter() byte {
	return '`'
}

func (s sqliteDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON CONFLICT(")
	for i, col := range upsert.conflictColumns {
		if i > 0 {
			b.sb.WriteByte(',')
		}
		err := b.buildColumn(Column{name: col})
		if err != nil {
			return err
		}
	}
	b.sb.WriteString(") DO UPDATE SET ")
	for idx, assign := range upsert.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			// 字段不对，或者说列不对
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=")
			if err := b.buildExpression(a.val); err != nil {
				return err
			}
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			// 字段不对，或者说列不对
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		default:
			return errs.NewErrUnsupportedAssignable(assign)
		}
	}
	return nil
}


type postgreDialect struct {
	standardSQL
}