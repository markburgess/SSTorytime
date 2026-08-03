//**************************************************************
//
// db_insertion.go
//
//**************************************************************

package sst

import (
	"fmt"
	"os"
	"strings"

	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

//**************************************************************

func FormDBNode(sst *PoSST, n Node) string {

	// Add node version setting explicit CPtr value, note different function call
	// We use this function when we ARE managing/counting CPtr values ourselves
	// Returns SQL only for batch builders; prefer InsertNodeFn for single calls.

	n.L, n.NPtr.Class = StorageClass(n.S)

	cptr := n.NPtr.CPtr
	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)
	seqstr := "false"
	if n.Seq {
		seqstr = "true"
	}

	return fmt.Sprintf("SELECT InsertNode(%d,%d,%d,'%s','%s',%s);\n", n.L, n.NPtr.Class, cptr, es, ec, seqstr)
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

func AppendDBLinkToNodeCommand(sst *PoSST, n1ptr NodePtr, lnk Link, sttype int) string {

	// Legacy string form kept for batch tools; prefer AppendDBLinkToNode.

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS, sttype)
		os.Exit(-1)
	}

	if n1ptr == lnk.Dst {
		return ""
	}

	linkval := fmt.Sprintf("(%d, %f, %d, (%d,%d)::NodePtr)", lnk.Arr, lnk.Wgt, lnk.Ctx, lnk.Dst.Class, lnk.Dst.CPtr)
	literal := fmt.Sprintf("%s::Link", linkval)
	link_table := STTypeDBChannel(sttype)

	return fmt.Sprintf("UPDATE NODE SET %s=array_append(%s,%s) WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d' AND (%s IS NULL OR NOT %s = ANY(%s));\n",
		link_table, link_table, literal, n1ptr.CPtr, n1ptr.Class, link_table, literal, link_table)
}

// **************************************************************************

func AppendDBLinkArrayToNode(sst *PoSST, nptr NodePtr, array string, sttype int) string {

	// Want to make this idempotent, because SQL is not (and not clause)

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS, sttype)
		os.Exit(-1)
	}

	if sst.Q != nil {
		cptr := int32(nptr.CPtr)
		chan_ := int32(nptr.Class)
		// FormatSQLLinkArray returns {...} without outer quotes sometimes with braces
		arr := array
		if !strings.HasPrefix(arr, "{") {
			arr = "{" + arr + "}"
		}
		var err error
		switch sttype {
		case -EXPRESS:
			err = sst.Q.SetLinkArrayIm3(sst.ctx(), sqlc.SetLinkArrayIm3Params{Column1: cptr, Column2: chan_, Column3: arr})
		case -CONTAINS:
			err = sst.Q.SetLinkArrayIm2(sst.ctx(), sqlc.SetLinkArrayIm2Params{Column1: cptr, Column2: chan_, Column3: arr})
		case -LEADSTO:
			err = sst.Q.SetLinkArrayIm1(sst.ctx(), sqlc.SetLinkArrayIm1Params{Column1: cptr, Column2: chan_, Column3: arr})
		case NEAR:
			err = sst.Q.SetLinkArrayIn0(sst.ctx(), sqlc.SetLinkArrayIn0Params{Column1: cptr, Column2: chan_, Column3: arr})
		case LEADSTO:
			err = sst.Q.SetLinkArrayIl1(sst.ctx(), sqlc.SetLinkArrayIl1Params{Column1: cptr, Column2: chan_, Column3: arr})
		case CONTAINS:
			err = sst.Q.SetLinkArrayIc2(sst.ctx(), sqlc.SetLinkArrayIc2Params{Column1: cptr, Column2: chan_, Column3: arr})
		case EXPRESS:
			err = sst.Q.SetLinkArrayIe3(sst.ctx(), sqlc.SetLinkArrayIe3Params{Column1: cptr, Column2: chan_, Column3: arr})
		}
		if err != nil {
			fmt.Println("SetLinkArray failed", err)
		}
		return ""
	}

	link_table := STTypeDBChannel(sttype)
	return fmt.Sprintf("UPDATE NODE SET %s='%s' WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d';\n",
		link_table, array, nptr.CPtr, nptr.Class)
}

//
// db_insertion.go
//
