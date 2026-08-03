package sst

import (
	"context"
	"fmt"
	"strings"

	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

// ctx returns the session context or background.
func (sst *PoSST) ctx() context.Context {
	if sst.Ctx != nil {
		return sst.Ctx
	}
	return context.Background()
}

// ensureQ returns sqlc querier (shared).
func (sst *PoSST) ensureQ() *sqlc.Queries {
	return sst.Q
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
