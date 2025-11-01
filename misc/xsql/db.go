package xsql

import (
	"context"
	"database/sql"
	"os"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TransactFunc func(ctx context.Context, s sqlx.Session) error

type beginTxKey struct{}

// 对db的封装
type DB struct {
	sc sqlx.SqlConn
}

func New(s sqlx.SqlConn) *DB {
	d := &DB{
		sc: s,
	}

	return d
}

func NewFromEnv() *DB {
	return New(sqlx.NewMysql(GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	)))
}

func (d *DB) getSess(ctx context.Context) sqlx.Session {
	if tx, ok := ctx.Value(beginTxKey{}).(sqlx.Session); ok {
		return tx
	}
	return d.sc
}

func (d *DB) Conn() sqlx.SqlConn {
	return d.sc
}

func (d *DB) Transact(ctx context.Context, fn func(ctx context.Context) error) error {
	_, already := ctx.Value(beginTxKey{}).(sqlx.Session)
	if !already {
		err := d.sc.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
			ctx = context.WithValue(ctx, beginTxKey{}, s)
			return fn(ctx)
		})

		return err
	}

	return fn(ctx)
}

func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.sc.Exec(query, args...)
}

func (d *DB) ExecCtx(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.getSess(ctx).ExecCtx(ctx, query, args...)
}

func (d *DB) Prepare(query string) (sqlx.StmtSession, error) {
	return d.sc.Prepare(query)
}

func (d *DB) PrepareCtx(ctx context.Context, query string) (sqlx.StmtSession, error) {
	return d.getSess(ctx).PrepareCtx(ctx, query)
}

func (d *DB) QueryRow(v any, query string, args ...any) error {
	return d.sc.QueryRow(v, query, args...)
}

func (d *DB) QueryRowCtx(ctx context.Context, v any, query string, args ...any) error {
	return d.getSess(ctx).QueryRowCtx(ctx, v, query, args...)
}

func (d *DB) QueryRowPartial(v any, query string, args ...any) error {
	return d.sc.QueryRowPartial(v, query, args...)
}

func (d *DB) QueryRowPartialCtx(ctx context.Context, v any, query string, args ...any) error {
	return d.getSess(ctx).QueryRowPartialCtx(ctx, v, query, args...)
}

func (d *DB) QueryRows(v any, query string, args ...any) error {
	return d.sc.QueryRows(v, query, args...)
}

func (d *DB) QueryRowsCtx(ctx context.Context, v any, query string, args ...any) error {
	return d.getSess(ctx).QueryRowsCtx(ctx, v, query, args...)
}

func (d *DB) QueryRowsPartial(v any, query string, args ...any) error {
	return d.sc.QueryRowsPartial(v, query, args...)
}

func (d *DB) QueryRowsPartialCtx(ctx context.Context, v any, query string, args ...any) error {
	return d.getSess(ctx).QueryRowsPartialCtx(ctx, v, query, args...)
}
