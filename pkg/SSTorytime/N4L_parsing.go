//**************************************************************
//
// N4L_parsing.go
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

func AppendTextToDirectory(sst *PoSST,event Node,ErrFunc func(string)) NodePtr {

	var cnode_slot ClassedNodePtr = -1
	var ok bool = false
	var node_alloc_ptr NodePtr = NO_NODE_PTR

	cnode_slot,ok = CheckExisting(sst,event)

	node_alloc_ptr.Class = event.NPtr.Class

	if ok {
		// This node already exists
		node_alloc_ptr.CPtr = cnode_slot
		IdempAddChapterSeqToNode(sst,node_alloc_ptr.Class,node_alloc_ptr.CPtr,event.Chap,event.Seq)
		return node_alloc_ptr
	}

	switch event.NPtr.Class {
	case N1GRAM:
		cnode_slot = sst.NODE_DIRECTORY.N1_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		sst.NODE_DIRECTORY.N1directory = append(sst.NODE_DIRECTORY.N1directory,event)
		sst.NODE_DIRECTORY.N1grams[event.S] = cnode_slot
		sst.NODE_DIRECTORY.N1_top++ 
	case N2GRAM:
		cnode_slot = sst.NODE_DIRECTORY.N2_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		sst.NODE_DIRECTORY.N2directory = append(sst.NODE_DIRECTORY.N2directory,event)
		sst.NODE_DIRECTORY.N2grams[event.S] = cnode_slot
		sst.NODE_DIRECTORY.N2_top++
	case N3GRAM:
		cnode_slot = sst.NODE_DIRECTORY.N3_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		sst.NODE_DIRECTORY.N3directory = append(sst.NODE_DIRECTORY.N3directory,event)
		sst.NODE_DIRECTORY.N3grams[event.S] = cnode_slot
		sst.NODE_DIRECTORY.N3_top++
	case LT128:
		cnode_slot = sst.NODE_DIRECTORY.LT128_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		sst.NODE_DIRECTORY.LT128directory = append(sst.NODE_DIRECTORY.LT128directory,event)
		sst.NODE_DIRECTORY.LT128[event.S] = cnode_slot
		sst.NODE_DIRECTORY.LT128_top++
	case LT1024:
		cnode_slot = sst.NODE_DIRECTORY.LT1024_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		sst.NODE_DIRECTORY.LT1024 = append(sst.NODE_DIRECTORY.LT1024,event)
		sst.NODE_DIRECTORY.LT1024_top++
	case GT1024:
		cnode_slot = sst.NODE_DIRECTORY.GT1024_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		sst.NODE_DIRECTORY.GT1024 = append(sst.NODE_DIRECTORY.GT1024,event)
		sst.NODE_DIRECTORY.GT1024_top++
	}

	event.NPtr = node_alloc_ptr
	
	return node_alloc_ptr
}

//**************************************************************

func CheckExisting(sst *PoSST,event Node) (ClassedNodePtr,bool) {

	var cnode_slot ClassedNodePtr = -1
	var ok bool = false
	ignore_caps := false

	switch event.NPtr.Class {
	case N1GRAM:
		cnode_slot,ok = sst.NODE_DIRECTORY.N1grams[event.S]
	case N2GRAM:
		cnode_slot,ok = sst.NODE_DIRECTORY.N2grams[event.S]
	case N3GRAM:
		cnode_slot,ok = sst.NODE_DIRECTORY.N3grams[event.S]
	case LT128:
		cnode_slot,ok = sst.NODE_DIRECTORY.LT128[event.S]
	case LT1024:
		cnode_slot,ok = LinearFindText(sst.NODE_DIRECTORY.LT1024,event,ignore_caps)
	case GT1024:
		cnode_slot,ok = LinearFindText(sst.NODE_DIRECTORY.GT1024,event,ignore_caps)
	}

	return cnode_slot,ok
}

//**************************************************************

func CheckAltCaps(sst *PoSST,event Node,ErrFunc func(string)) {
		
	// Check for alternative caps

	var keyNPtr NodePtr
	
	switch event.NPtr.Class {
	case N1GRAM:
		for key := range sst.NODE_DIRECTORY.N1grams {

			keyNPtr.Class = N1GRAM
			keyNPtr.CPtr = sst.NODE_DIRECTORY.N1grams[key]
			n := sst.NODE_DIRECTORY.N1directory[keyNPtr.CPtr]

			if DifferentCaps(n,event) {
				NearEquiv(sst,keyNPtr,event.NPtr,key,event.S,ErrFunc)
			}
		}
	case N2GRAM:
		for key := range sst.NODE_DIRECTORY.N2grams {
			
			keyNPtr.Class = N2GRAM
			keyNPtr.CPtr = sst.NODE_DIRECTORY.N2grams[key]
			n := sst.NODE_DIRECTORY.N2directory[keyNPtr.CPtr]

			if DifferentCaps(n,event) {
				NearEquiv(sst,keyNPtr,event.NPtr,key,event.S,ErrFunc)
			}

		}
	case N3GRAM:
		for key := range sst.NODE_DIRECTORY.N3grams {

			keyNPtr.Class = N3GRAM
			keyNPtr.CPtr = sst.NODE_DIRECTORY.N3grams[key]
			n := sst.NODE_DIRECTORY.N3directory[keyNPtr.CPtr]

			if DifferentCaps(n,event) {
				NearEquiv(sst,keyNPtr,event.NPtr,key,event.S,ErrFunc)
			}

		}

	case LT128:
		for key := range sst.NODE_DIRECTORY.LT128 {
			
			keyNPtr.Class = N3GRAM
			keyNPtr.CPtr = sst.NODE_DIRECTORY.LT128[key]
			n := sst.NODE_DIRECTORY.LT128directory[keyNPtr.CPtr]
			
			if DifferentCaps(n,event) {
				NearEquiv(sst,keyNPtr,event.NPtr,key,event.S,ErrFunc)
			}
			
		}
	}
}

//**************************************************************

func DifferentCaps(n1,n2 Node) bool {

	const margin = 2
	
	if n1.L != n2.L {
		return false
	}

	s1 := n1.S
	s2 := n2.S
	
	if (s1 != s2) && (strings.ToLower(s1) == strings.ToLower(s2)) {
		return true
	}

	return false
}

//**************************************************************

func NearEquiv(sst *PoSST,n1,n2 NodePtr, s1,s2 string,ErrFunc func(string)) {

	// Alternative capitalizations are NEAR = "capitalization" one another

	var lnk Link
	var context = []string{"ambiguous"}

	lnk.Arr = sst.ARROW_SHORT_DIR["caps"]
	lnk.Wgt = 1
	lnk.Ctx = RegisterContext(sst,nil,context)
	AppendLinkToNode(sst,n1,lnk,n2)
	AppendLinkToNode(sst,n2,lnk,n1)

	ErrFunc(WARN_DIFFERENT_CAPITALS+" ("+s1+" vs "+s2+") - linking as NEAR ")
}

//**************************************************************

func IdempAddChapterSeqToNode(sst *PoSST,class int,cptr ClassedNodePtr,chap string,seq bool) {

	/* In the DB version, we have handle chapter collisions
           we want all similar names to have a single node for lateral
           association, but we need to be able to search by chapter too,
           so merge the chapters as an attribute list */

	var node Node

	node = UpdateSeqStatus(sst,class,cptr,seq)

	if strings.Contains(node.Chap,chap) {
		return
	}

	newchap := node.Chap + "," + chap

	switch class {
	case N1GRAM:
		sst.NODE_DIRECTORY.N1directory[cptr].Chap = newchap
	case N2GRAM:
		sst.NODE_DIRECTORY.N2directory[cptr].Chap = newchap
	case N3GRAM:
		sst.NODE_DIRECTORY.N3directory[cptr].Chap = newchap
	case LT128:
		sst.NODE_DIRECTORY.LT128directory[cptr].Chap = newchap
	case LT1024:
		sst.NODE_DIRECTORY.LT1024[cptr].Chap = newchap
	case GT1024:
		sst.NODE_DIRECTORY.GT1024[cptr].Chap = newchap
	}
}

//**************************************************************

func UpdateSeqStatus(sst *PoSST,class int,cptr ClassedNodePtr,seq bool) Node {

	switch class {
	case N1GRAM:
		sst.NODE_DIRECTORY.N1directory[cptr].Seq = sst.NODE_DIRECTORY.N1directory[cptr].Seq || seq
		return sst.NODE_DIRECTORY.N1directory[cptr]
	case N2GRAM:
		sst.NODE_DIRECTORY.N2directory[cptr].Seq = sst.NODE_DIRECTORY.N2directory[cptr].Seq || seq
		return sst.NODE_DIRECTORY.N2directory[cptr]
	case N3GRAM:
		sst.NODE_DIRECTORY.N3directory[cptr].Seq = sst.NODE_DIRECTORY.N3directory[cptr].Seq || seq
		return sst.NODE_DIRECTORY.N3directory[cptr]
	case LT128:
		sst.NODE_DIRECTORY.LT128directory[cptr].Seq = sst.NODE_DIRECTORY.LT128directory[cptr].Seq || seq
		return sst.NODE_DIRECTORY.LT128directory[cptr]
	case LT1024:
		sst.NODE_DIRECTORY.LT1024[cptr].Seq = sst.NODE_DIRECTORY.LT1024[cptr].Seq || seq
		return sst.NODE_DIRECTORY.LT1024[cptr]
	case GT1024:
		sst.NODE_DIRECTORY.GT1024[cptr].Seq = sst.NODE_DIRECTORY.GT1024[cptr].Seq || seq
		return sst.NODE_DIRECTORY.GT1024[cptr]
	}

	fmt.Println("Non existent node class (shouldn't happen)")
	os.Exit(-1)
	var dummy Node
	return dummy
}

//**************************************************************

func InsertArrowDirectory(sst *PoSST,stname,alias,name,pm string) ArrowPtr {

	// Insert an arrow into the forward/backward indices

	var newarrow ArrowDirectory

	// Check is already exists - harmless

	prev_alias,a_exists := sst.ARROW_SHORT_DIR[alias]
	prev_name,n_exists := sst.ARROW_LONG_DIR[name]

	// long and short versions the same
	
	if a_exists && n_exists {
		if prev_alias == prev_name {
			return prev_alias
		}
	}

	// Already defined, no need to do it again, warning
	for a := range sst.ARROW_DIRECTORY {
		if sst.ARROW_DIRECTORY[a].Long == name {
			fmt.Printf(" !! Info, long name (%s) is previously found with short name: %s\n",sst.ARROW_DIRECTORY[a].Long,sst.ARROW_DIRECTORY[a].Short)
			fmt.Println(" !! You might need to wipe and recompile if an old definition is cached")
			return ArrowPtr(-1)
		}

		if sst.ARROW_DIRECTORY[a].Short == alias {
			fmt.Println(" !! Info, short name (%s) is previously found with long name: %s\n",sst.ARROW_DIRECTORY[a].Long,sst.ARROW_DIRECTORY[a].Short)
			fmt.Println(" !! You might need to wipe and recompile if an old definition is cached")
			return ArrowPtr(-1)
		}
	}

	newarrow.STAindex = GetSTIndexByName(stname,pm)
	newarrow.Long = name
	newarrow.Short = alias
	newarrow.Ptr = sst.ARROW_DIRECTORY_TOP

	sst.ARROW_DIRECTORY = append(sst.ARROW_DIRECTORY,newarrow)
	sst.ARROW_SHORT_DIR[alias] = sst.ARROW_DIRECTORY_TOP
	sst.ARROW_LONG_DIR[name] = sst.ARROW_DIRECTORY_TOP
	sst.ARROW_DIRECTORY_TOP++

	return sst.ARROW_DIRECTORY_TOP-1
}

//**************************************************************

func InsertInverseArrowDirectory(sst *PoSST,fwd,bwd ArrowPtr) {

	if fwd == ArrowPtr(-1) || bwd == ArrowPtr(-1) {
		return
	}

	// Lookup inverse by long name, only need this in search presentation

	sst.INVERSE_ARROWS[fwd] = bwd
	sst.INVERSE_ARROWS[bwd] = fwd
}

//**************************************************************

func AppendLinkToNode(sst *PoSST,frptr NodePtr,link Link,toptr NodePtr) {

	frclass := frptr.Class
	frm := frptr.CPtr
	stindex := sst.ARROW_DIRECTORY[link.Arr].STAindex

	link.Dst = toptr // fill in the last part of the reference

	// Idempotently add any new context strings to the current list
	// between from and to nodes -- stindex tells us which link type, so implicit in the arrow type
	// the empty arrow is used to record node context, which is type LEADSTO

	switch frclass {

	case N1GRAM:
		sst.NODE_DIRECTORY.N1directory[frm].I[stindex] = MergeLinkLists(sst,sst.NODE_DIRECTORY.N1directory[frm].I[stindex],link)
	case N2GRAM:
		sst.NODE_DIRECTORY.N2directory[frm].I[stindex] = MergeLinkLists(sst,sst.NODE_DIRECTORY.N2directory[frm].I[stindex],link)
	case N3GRAM:
		sst.NODE_DIRECTORY.N3directory[frm].I[stindex] = MergeLinkLists(sst,sst.NODE_DIRECTORY.N3directory[frm].I[stindex],link)
	case LT128:
		sst.NODE_DIRECTORY.LT128directory[frm].I[stindex] = MergeLinkLists(sst,sst.NODE_DIRECTORY.LT128directory[frm].I[stindex],link)
	case LT1024:
		sst.NODE_DIRECTORY.LT1024[frm].I[stindex] = MergeLinkLists(sst,sst.NODE_DIRECTORY.LT1024[frm].I[stindex],link)
	case GT1024:
		sst.NODE_DIRECTORY.GT1024[frm].I[stindex] = MergeLinkLists(sst,sst.NODE_DIRECTORY.GT1024[frm].I[stindex],link)
	}
}

//**************************************************************

func MergeLinkLists(sst *PoSST,linklist []Link,lnk Link) []Link {

	// Ensure all arrows and contexts in lnk are in list for the appropriate arrows

	new_ctxstr := GetContext(sst,lnk.Ctx)
	new_ctxlist := strings.Split(new_ctxstr,",")

	// Check if the arrow is already there to add to its context

	for l := range linklist {
		if linklist[l].Arr == lnk.Arr && linklist[l].Dst == lnk.Dst {

			already_ctxstr := GetContext(sst,linklist[l].Ctx)
			already_ctxlist := strings.Split(already_ctxstr,",")

			linklist[l].Ctx = MergeContextLists(sst,already_ctxlist,new_ctxlist)

			return linklist
		}
	}

	// if not already there, add this arrow

	linklist = append(linklist,lnk)
	return linklist
}

//**************************************************************

func MergeContextLists(sst *PoSST,one,two []string) ContextPtr {

	var merging = make(map[string]bool)
	var merged []string

	for s := range one {
		merging[one[s]] = true
	}

	for s := range two {
		merging[two[s]] = true
	}

	for s := range merging {
		if s != "_sequence_" {
			merged = append(merged,s)
		}
	}

	ctxstr := List2String(merged)

	// Register the merger of contexts

	ctxptr,ok := sst.CONTEXT_DIR[ctxstr]

	if ok {
		return ctxptr
	} else {
		var cd ContextDirectory
		cd.Context = ctxstr
		cd.Ptr = sst.CONTEXT_TOP
		sst.CONTEXT_DIRECTORY = append(sst.CONTEXT_DIRECTORY,cd)
		sst.CONTEXT_DIR[ctxstr] = sst.CONTEXT_TOP
		ctxptr = sst.CONTEXT_TOP
		sst.CONTEXT_TOP++
	}

	return ctxptr
}

//**************************************************************

func LinearFindText(in []Node,event Node,ignore_caps bool) (ClassedNodePtr,bool) {

	for i := 0; i < len(in); i++ {

		if event.L != in[i].L {
			continue
		}

		if ignore_caps {
			if strings.ToLower(in[i].S) == strings.ToLower(event.S) {
				return ClassedNodePtr(i),true
			}
		} else {
			if in[i].S == event.S {
				return ClassedNodePtr(i),true
			}
		}
	}

	return -1,false
}


//
// end N4L_parsing.go
//

