package sst

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

// ctx returns the session context or background.
func (sst *PoSST) ctx() context.Context {
	if sst.Ctx != nil {
		return sst.Ctx
	}
	return context.Background()
}

// useSimpleProtocol is true for multi-statement SQL batches (N4L upload).
// pgx extended/prepared protocol rejects "multiple commands in a prepared statement".
func useSimpleProtocol(sql string, args []any) bool {
	if len(args) > 0 {
		return false
	}
	// Count statement terminators; upload builds ";\n" joined batches + optional COMMIT.
	return strings.Count(sql, ";") > 1 || strings.Contains(strings.ToUpper(sql), "COMMIT")
}

// query runs SQL via pgx pool (no lib/pq). Prefer sst.Q sqlc methods when possible.
func (sst *PoSST) query(sql string, args ...any) (pgx.Rows, error) {
	if sst.Pool == nil {
		return nil, pgx.ErrNoRows
	}
	if useSimpleProtocol(sql, args) {
		return sst.Pool.Query(sst.ctx(), sql, append([]any{pgx.QueryExecModeSimpleProtocol}, args...)...)
	}
	return sst.Pool.Query(sst.ctx(), sql, args...)
}

// queryRow runs a single-row query via pgx pool.
func (sst *PoSST) queryRow(sql string, args ...any) pgx.Row {
	if useSimpleProtocol(sql, args) {
		return sst.Pool.QueryRow(sst.ctx(), sql, append([]any{pgx.QueryExecModeSimpleProtocol}, args...)...)
	}
	return sst.Pool.QueryRow(sst.ctx(), sql, args...)
}

// exec runs a statement via pgx pool.
func (sst *PoSST) exec(sql string, args ...any) (pgconn.CommandTag, error) {
	if useSimpleProtocol(sql, args) {
		return sst.Pool.Exec(sst.ctx(), sql, append([]any{pgx.QueryExecModeSimpleProtocol}, args...)...)
	}
	return sst.Pool.Exec(sst.ctx(), sql, args...)
}

// ensureQ returns sqlc querier (shared).
func (sst *PoSST) ensureQ() *sqlc.Queries {
	return sst.Q
}

// mustPool is used where pool is required.
func (sst *PoSST) mustPool() *pgxpool.Pool {
	return sst.Pool
}
