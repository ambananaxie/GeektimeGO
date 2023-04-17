package main

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

type ShadowPool struct {
	live gorm.ConnPool
	shadow gorm.ConnPool

	// live *gorm.DB
	// shadow *gorm.DB
}

func (s *ShadowPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.PrepareContext(ctx, query)
	}
	return s.live.PrepareContext(ctx, query)
}

func (s *ShadowPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.ExecContext(ctx, query, args...)
	}
	return s.live.ExecContext(ctx, query, args...)
}

func (s *ShadowPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.QueryContext(ctx, query, args...)
	}
	return s.live.QueryContext(ctx, query, args...)
}

func (s *ShadowPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.QueryRowContext(ctx, query, args...)
	}
	return s.live.QueryRowContext(ctx, query, args...)
}

