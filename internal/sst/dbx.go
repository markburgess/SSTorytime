package sst

import (
	"context"

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

// query runs SQL via pgx pool (no lib/pq). Prefer sst.Q sqlc methods when possible.
func (sst *PoSST) query(sql string, args ...any) (pgx.Rows, error) {
	if sst.Pool == nil {
		return nil, pgx.ErrNoRows
	}
	return sst.Pool.Query(sst.ctx(), sql, args...)
}

// queryRow runs a single-row query via pgx pool.
func (sst *PoSST) queryRow(sql string, args ...any) pgx.Row {
	return sst.Pool.QueryRow(sst.ctx(), sql, args...)
}

// exec runs a statement via pgx pool.
func (sst *PoSST) exec(sql string, args ...any) (pgconn.CommandTag, error) {
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
