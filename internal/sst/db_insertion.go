//**************************************************************
//
// db_insertion.go
//
//**************************************************************

package sst

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

// Package-level error table.
var (
	// ErrNoQuerier is returned when a PoSST has no sqlc querier bound.
	ErrNoQuerier         = errors.New("no querier")
	ErrIllegalLinkClass  = errors.New(ERR_ILLEGAL_LINK_CLASS)
)

//**************************************************************

// FormDBNode inserts a node with an explicit CPtr via the InsertNode PL/pgSQL function.
func FormDBNode(sst *PoSST, n Node) error {
	n.L, n.NPtr.Class = StorageClass(n.S)
	if sst.Q == nil {
		return ErrNoQuerier
	}
	return sst.Q.InsertNodeFn(sst.ctx(), sqlc.InsertNodeFnParams{
		Column1: int32(n.L),
		Column2: int32(n.NPtr.Class),
		Column3: int32(n.NPtr.CPtr),
		Isi:     n.S,
		Ichapi:  n.Chap,
		Column6: n.Seq,
	})
}

// **************************************************************************

func IdempDBAddNode(sst *PoSST, n Node) Node {

	// We use this function when we aren't counting CPtr values
	// This functon may be deprecated in future

	n.L, n.NPtr.Class = StorageClass(n.S)

	if sst.Q == nil {
		return n
	}

	row, err := sst.Q.IdempAppendNode(sst.ctx(), sqlc.IdempAppendNodeParams{
		Column1: int32(n.L),
		Column2: int32(n.NPtr.Class),
		Isi:     n.S,
		Ichapi:  n.Chap,
	})
	if err != nil {
		s := fmt.Sprint("Failed to add node", err)
		if !strings.Contains(s, "duplicate key") {
			fmt.Println(s, "FAILED", err)
		}
		return n
	}

	// ret_cptr column is filled from Chan; ret_channel from CPtr (see migration).
	n.NPtr.Class = int(row.Chan)
	n.NPtr.CPtr = ClassedNodePtr(row.Cptr)
	return n
}

// **************************************************************************

func IdempDBAddLink(sst *PoSST, from Node, link Link, to Node) {

	// API Entry point for registering links

	frptr := from.NPtr
	toptr := to.NPtr

	link.Dst = toptr // it might have changed, so override

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

	// Double up the reverse definition for easy indexing of both in/out arrows
	// But be careful not the make the graph undirected by mistake

	var invlink Link
	invlink.Arr = sst.INVERSE_ARROWS[link.Arr]
	invlink.Wgt = link.Wgt
	invlink.Dst = frptr
	AppendDBLinkToNode(sst, toptr, invlink, -sttype)
}

// **************************************************************************

func AppendDBLinkToNode(sst *PoSST, n1ptr NodePtr, lnk Link, sttype int) bool {

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS, sttype)
		os.Exit(-1)
	}

	if n1ptr == lnk.Dst {
		return true
	}

	if sst.Q == nil {
		return false
	}

	// Params: cptr, chan, arr, wgt, ctx, dst.chan, dst.cptr
	cptr := int32(n1ptr.CPtr)
	chan_ := int32(n1ptr.Class)
	arr := int32(lnk.Arr)
	wgt := float32(lnk.Wgt)
	ctx := int32(lnk.Ctx)
	dch := int32(lnk.Dst.Class)
	dcp := int32(lnk.Dst.CPtr)

	var err error
	switch sttype {
	case -EXPRESS:
		err = sst.Q.AppendLinkIm3(sst.ctx(), sqlc.AppendLinkIm3Params{Column1: cptr, Column2: chan_, Column3: arr, Column4: wgt, Column5: ctx, Column6: dch, Column7: dcp})
	case -CONTAINS:
		err = sst.Q.AppendLinkIm2(sst.ctx(), sqlc.AppendLinkIm2Params{Column1: cptr, Column2: chan_, Column3: arr, Column4: wgt, Column5: ctx, Column6: dch, Column7: dcp})
	case -LEADSTO:
		err = sst.Q.AppendLinkIm1(sst.ctx(), sqlc.AppendLinkIm1Params{Column1: cptr, Column2: chan_, Column3: arr, Column4: wgt, Column5: ctx, Column6: dch, Column7: dcp})
	case NEAR:
		err = sst.Q.AppendLinkIn0(sst.ctx(), sqlc.AppendLinkIn0Params{Column1: cptr, Column2: chan_, Column3: arr, Column4: wgt, Column5: ctx, Column6: dch, Column7: dcp})
	case LEADSTO:
		err = sst.Q.AppendLinkIl1(sst.ctx(), sqlc.AppendLinkIl1Params{Column1: cptr, Column2: chan_, Column3: arr, Column4: wgt, Column5: ctx, Column6: dch, Column7: dcp})
	case CONTAINS:
		err = sst.Q.AppendLinkIc2(sst.ctx(), sqlc.AppendLinkIc2Params{Column1: cptr, Column2: chan_, Column3: arr, Column4: wgt, Column5: ctx, Column6: dch, Column7: dcp})
	case EXPRESS:
		err = sst.Q.AppendLinkIe3(sst.ctx(), sqlc.AppendLinkIe3Params{Column1: cptr, Column2: chan_, Column3: arr, Column4: wgt, Column5: ctx, Column6: dch, Column7: dcp})
	default:
		fmt.Println(ERR_ILLEGAL_LINK_CLASS, sttype)
		os.Exit(-1)
	}
	if err != nil {
		fmt.Println("Failed to append", err)
		return false
	}
	return true
}

// **************************************************************************

// AppendDBLinkArrayToNode replaces one ST-channel link array on a node (sqlc).
func AppendDBLinkArrayToNode(sst *PoSST, nptr NodePtr, array string, sttype int) error {

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS, sttype)
		os.Exit(-1)
	}

	if sst.Q == nil {
		return ErrNoQuerier
	}

	cptr := int32(nptr.CPtr)
	chan_ := int32(nptr.Class)
	arr := array
	if !strings.HasPrefix(arr, "{") {
		arr = "{" + arr + "}"
	}

	switch sttype {
	case -EXPRESS:
		return sst.Q.SetLinkArrayIm3(sst.ctx(), sqlc.SetLinkArrayIm3Params{Column1: cptr, Column2: chan_, Column3: arr})
	case -CONTAINS:
		return sst.Q.SetLinkArrayIm2(sst.ctx(), sqlc.SetLinkArrayIm2Params{Column1: cptr, Column2: chan_, Column3: arr})
	case -LEADSTO:
		return sst.Q.SetLinkArrayIm1(sst.ctx(), sqlc.SetLinkArrayIm1Params{Column1: cptr, Column2: chan_, Column3: arr})
	case NEAR:
		return sst.Q.SetLinkArrayIn0(sst.ctx(), sqlc.SetLinkArrayIn0Params{Column1: cptr, Column2: chan_, Column3: arr})
	case LEADSTO:
		return sst.Q.SetLinkArrayIl1(sst.ctx(), sqlc.SetLinkArrayIl1Params{Column1: cptr, Column2: chan_, Column3: arr})
	case CONTAINS:
		return sst.Q.SetLinkArrayIc2(sst.ctx(), sqlc.SetLinkArrayIc2Params{Column1: cptr, Column2: chan_, Column3: arr})
	case EXPRESS:
		return sst.Q.SetLinkArrayIe3(sst.ctx(), sqlc.SetLinkArrayIe3Params{Column1: cptr, Column2: chan_, Column3: arr})
	default:
		return fmt.Errorf("%w: %d", ErrIllegalLinkClass, sttype)
	}
}

//
// db_insertion.go
//
