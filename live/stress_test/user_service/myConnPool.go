package main

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

type ConnPoolWrapper struct {
	prod gorm.ConnPool
	test gorm.ConnPool
}

func (c ConnPoolWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.GetConnPool(ctx).PrepareContext(ctx, query)
}

func (c ConnPoolWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.GetConnPool(ctx).ExecContext(ctx, query, args...)
}

func (c ConnPoolWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.GetConnPool(ctx).QueryContext(ctx, query, args...)
}

func (c ConnPoolWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.GetConnPool(ctx).QueryRowContext(ctx, query, args...)
}

func (c *ConnPoolWrapper) GetConnPool(ctx context.Context) gorm.ConnPool {
	if ctx.Value("stress_test") == "true" {
		return c.test
	}
	return c.prod
}
