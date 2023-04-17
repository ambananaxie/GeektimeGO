package gorm

import (
	"context"
	"database/sql"
)

type ShadowPool struct {
	live *sql.DB
	shadow *sql.DB
}

func (s *ShadowPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.BeginTx(ctx, opts)
	}
	return s.live.BeginTx(ctx, opts)
}

func (s *ShadowPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.PrepareContext(ctx, query)
	}
	return s.live.PrepareContext(ctx, query)
}

func (s *ShadowPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.ExecContext(ctx, query)
	}
	return s.live.ExecContext(ctx, query)
}

func (s *ShadowPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.QueryContext(ctx, query)
	}
	return s.live.QueryContext(ctx, query)
}

func (s *ShadowPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if ctx.Value("stress_test") == "true" {
		return s.shadow.QueryRowContext(ctx, query)
	}
	return s.live.QueryRowContext(ctx, query)
}

