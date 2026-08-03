package sst

import (
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

// IdempDBAddNode inserts a node if text is new, else returns existing pointer.
func IdempDBAddNode(sst *PoSST, n Node) Node {
	n.L, n.NPtr.Class = StorageClass(n.S)
	ctx := sst.Ctx
	if ctx == nil {
		ctx = sst.Ctx
	}

	existing, err := sst.Q.GetNodeByText(ctx, n.S)
	if err == nil {
		n.NPtr.Class = int(existing.Class)
		n.NPtr.CPtr = ClassedNodePtr(existing.Cptr)
		n.L = int(existing.L)
		n.Chap = existing.Chap
		n.Seq = existing.Seq
		return n
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		// try insert anyway if table empty / other
	}

	maxC, err := sst.Q.MaxCPtrForClass(ctx, int32(n.NPtr.Class))
	if err != nil {
		fmt.Println("MaxCPtrForClass:", err)
		return n
	}
	n.NPtr.CPtr = ClassedNodePtr(maxC + 1)

	err = sst.Q.InsertNode(ctx, sqlc.InsertNodeParams{
		Class: int32(n.NPtr.Class),
		Cptr:  int32(n.NPtr.CPtr),
		L:     int32(n.L),
		S:     n.S,
		Chap:  n.Chap,
		Seq:   n.Seq,
	})
	if err != nil {
		// race: fetch existing
		if existing, err2 := sst.Q.GetNodeByText(ctx, n.S); err2 == nil {
			n.NPtr.Class = int(existing.Class)
			n.NPtr.CPtr = ClassedNodePtr(existing.Cptr)
			return n
		}
		fmt.Println("InsertNode failed:", err)
	}
	return n
}

// FormDBNode kept for API shape; prefer InsertNode via GraphToDB.
func FormDBNode(sst *PoSST, n Node) string {
	return ""
}

func IdempDBAddLink(sst *PoSST, from Node, link Link, to Node) {
	frptr := from.NPtr
	toptr := to.NPtr
	link.Dst = toptr

	if frptr == toptr {
		fmt.Println("Self-loops are not allowed", from.S, from, link, to)
		os.Exit(-1)
	}
	if link.Arr < 0 || len(sst.ARROW_DIRECTORY) == 0 {
		fmt.Println("No arrows have yet been defined, so you can't rely on the arrow names")
		os.Exit(-1)
	}
	if link.Wgt == 0 {
		fmt.Println("Attempt to register a link with zero weight is pointless")
		os.Exit(-1)
	}

	sttype := STIndexToSTType(sst.ARROW_DIRECTORY[link.Arr].STAindex)
	AppendDBLinkToNode(sst, frptr, link, sttype)

	var invlink Link
	invlink.Arr = sst.INVERSE_ARROWS[link.Arr]
	invlink.Wgt = link.Wgt
	invlink.Dst = frptr
	invlink.Ctx = link.Ctx
	AppendDBLinkToNode(sst, toptr, invlink, -sttype)
}

func AppendDBLinkToNode(sst *PoSST, n1ptr NodePtr, lnk Link, sttype int) bool {
	ctx := sst.Ctx
	err := sst.Q.InsertEdge(ctx, sqlc.InsertEdgeParams{
		SrcClass: int32(n1ptr.Class),
		SrcCptr:  int32(n1ptr.CPtr),
		DstClass: int32(lnk.Dst.Class),
		DstCptr:  int32(lnk.Dst.CPtr),
		Arr:      int32(lnk.Arr),
		Wgt:      lnk.Wgt,
		Ctx:      int32(lnk.Ctx),
		St:       StTypeFromInt(sttype),
	})
	if err != nil {
		fmt.Println("InsertEdge failed:", err)
		return false
	}
	return true
}

func AppendDBLinkToNodeCommand(sst *PoSST, n1ptr NodePtr, lnk Link, sttype int) string {
	return ""
}

func AppendDBLinkArrayToNode(sst *PoSST, nptr NodePtr, array string, sttype int) string {
	return ""
}
