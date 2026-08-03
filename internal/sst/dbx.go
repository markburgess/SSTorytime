package sst

import (
	"context"
	"fmt"
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

func anyToInt(v any) int {
	switch x := v.(type) {
	case int32:
		return int(x)
	case int64:
		return int(x)
	case int:
		return x
	case float64:
		return int(x)
	case nil:
		return 0
	default:
		return 0
	}
}

func toInt32s(a []int) []int32 {
	out := make([]int32, len(a))
	for i, v := range a {
		out[i] = int32(v)
	}
	return out
}

func arrowToInt32s(a []ArrowPtr) []int32 {
	out := make([]int32, len(a))
	for i, v := range a {
		out[i] = int32(v)
	}
	return out
}

// nodePtrArrayLiteral is a Postgres text form for nodeptr[] (no SQL quotes).
func nodePtrArrayLiteral(array []NodePtr) string {
	if len(array) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, n := range array {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "\"(%d,%d)\"", n.Class, n.CPtr)
	}
	b.WriteByte('}')
	return b.String()
}

func strPtr(s string) *string { return &s }

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefInt32(p *int32) int {
	if p == nil {
		return 0
	}
	return int(*p)
}

// stPosFlags maps positive ST types into (leads, contains, express, near) flags.
func stPosFlags(sttypes []int) (lead, cont, expr, near bool) {
	for _, st := range sttypes {
		if st < 0 {
			st = -st
		}
		switch st {
		case LEADSTO:
			lead = true
		case CONTAINS:
			cont = true
		case EXPRESS:
			expr = true
		case NEAR:
			near = true
		}
	}
	return
}

// stChannelFlags enables im3..ie3 for adjacency (signed ST types).
func stChannelFlags(sttypes []int) (im3, im2, im1, in0, il1, ic2, ie3 bool) {
	for _, st := range sttypes {
		switch st {
		case -EXPRESS:
			im3 = true
		case -CONTAINS:
			im2 = true
		case -LEADSTO:
			im1 = true
		case NEAR:
			in0 = true
		case LEADSTO:
			il1 = true
		case CONTAINS:
			ic2 = true
		case EXPRESS:
			ie3 = true
		}
	}
	return
}
