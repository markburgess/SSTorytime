//**************************************************************
//
// db_insertion.go
//
//**************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	"strings"
	_ "github.com/lib/pq"

)

//**************************************************************

func FormDBNode(sst PoSST, n Node) string {

	// Add node version setting explicit CPtr value, note different function call
	// We use this function when we ARE managing/counting CPtr values ourselves

	var qstr,seqstr string

        n.L,n.NPtr.Class = StorageClass(n.S)
	
	cptr := n.NPtr.CPtr

	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)

	if n.Seq {
		seqstr = "true"
	} else {
		seqstr = "false"
	}

	qstr = fmt.Sprintf("SELECT InsertNode(%d,%d,%d,'%s','%s',%s);\n",n.L,n.NPtr.Class,cptr,es,ec,seqstr)
	return qstr
}



// **************************************************************************

func IdempDBAddNode(sst PoSST,n Node) Node {

	// We use this function when we aren't counting CPtr values
	// This functon may be deprecated in future

	var qstr string

	// No need to trust the values, ignore/overwrite CPtr

        n.L,n.NPtr.Class = StorageClass(n.S)

	es := SQLEscape(n.S)
	ec := SQLEscape(n.Chap)

	// Wrap BEGIN/END a single transaction

	qstr = fmt.Sprintf("SELECT IdempAppendNode(%d,%d,'%s','%s')",n.L,n.NPtr.Class,es,ec)

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		s := fmt.Sprint("Failed to add node",err)
		
		if strings.Contains(s,"duplicate key") {
		} else {
			fmt.Println(s,"FAILED \n",qstr,err)
		}
		return n
	}

	var whole string
	var cl,ch int

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&cl,&ch)
		}
		
		n.NPtr.Class = cl
		n.NPtr.CPtr = ClassedNodePtr(ch)
		
		row.Close()
	}

	return n
}

// **************************************************************************

func IdempDBAddLink(sst PoSST,from Node,link Link,to Node) {

	// API Entry point for registering links

	frptr := from.NPtr
	toptr := to.NPtr

	link.Dst = toptr // it might have changed, so override

	if frptr == toptr {
		fmt.Println("Self-loops are not allowed",from.S,from,link,to)
		os.Exit(-1)
	}

	if link.Arr < 0 || len(ARROW_DIRECTORY) == 0 {
		fmt.Println("No arrows have yet been defined, so you can't rely on the arrow names")
		os.Exit(-1)
	}

	if link.Wgt == 0 {
		fmt.Println("Attempt to register a link with zero weight is pointless")
		os.Exit(-1)
	}

	sttype := STIndexToSTType(ARROW_DIRECTORY[link.Arr].STAindex)

	AppendDBLinkToNode(sst,frptr,link,sttype)

	// Double up the reverse definition for easy indexing of both in/out arrows
	// But be careful not the make the graph undirected by mistake

	var invlink Link
	invlink.Arr = INVERSE_ARROWS[link.Arr]
	invlink.Wgt = link.Wgt
	invlink.Dst = frptr
	AppendDBLinkToNode(sst,toptr,invlink,-sttype)
}

// **************************************************************************

func AppendDBLinkToNode(sst PoSST, n1ptr NodePtr, lnk Link, sttype int) bool {

	qstr := AppendDBLinkToNodeCommand(sst,n1ptr,lnk,sttype)

	row,err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Failed to append",err,qstr)
	       return false
	}

	row.Close()
	return true
}

// **************************************************************************

func AppendDBLinkToNodeCommand(sst PoSST, n1ptr NodePtr, lnk Link, sttype int) string {

	// Want to make this idempotent, because SQL is not (and not clause)

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS,sttype)
		os.Exit(-1)
	}

	if n1ptr == lnk.Dst {
		return ""
	}

	//                       Arr,Wgt,Ctx,  Dst
	linkval := fmt.Sprintf("(%d, %f, %d, (%d,%d)::NodePtr)",lnk.Arr,lnk.Wgt,lnk.Ctx,lnk.Dst.Class,lnk.Dst.CPtr)

	literal := fmt.Sprintf("%s::Link",linkval)

	link_table := STTypeDBChannel(sttype)

	qstr := fmt.Sprintf("UPDATE NODE SET %s=array_append(%s,%s) WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d' AND (%s IS NULL OR NOT %s = ANY(%s));\n",
		link_table,
		link_table,
		literal,
		n1ptr.CPtr,
		n1ptr.Class,
		link_table,
		literal,
		link_table)

	return qstr
}

// **************************************************************************

func AppendDBLinkArrayToNode(sst PoSST, nptr NodePtr, array string, sttype int) string {

	// Want to make this idempotent, because SQL is not (and not clause)

	if sttype < -EXPRESS || sttype > EXPRESS {
		fmt.Println(ERR_ST_OUT_OF_BOUNDS,sttype)
		os.Exit(-1)
	}

	link_table := STTypeDBChannel(sttype)

	qstr := fmt.Sprintf("UPDATE NODE SET %s='%s' WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d';\n",
		link_table,
		array,
		nptr.CPtr,
		nptr.Class)

	return qstr
}


//
// db_insertion.go
//



