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

func AppendTextToDirectory(event Node,ErrFunc func(string)) NodePtr {

	var cnode_slot ClassedNodePtr = -1
	var ok bool = false
	var node_alloc_ptr NodePtr

	cnode_slot,ok = CheckExistingOrAltCaps(event,ErrFunc)

	node_alloc_ptr.Class = event.NPtr.Class

	if ok {
		node_alloc_ptr.CPtr = cnode_slot
		IdempAddChapterSeqToNode(node_alloc_ptr.Class,node_alloc_ptr.CPtr,event.Chap,event.Seq)
		return node_alloc_ptr
	}

	switch event.NPtr.Class {
	case N1GRAM:
		cnode_slot = NODE_DIRECTORY.N1_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.N1directory = append(NODE_DIRECTORY.N1directory,event)
		NODE_DIRECTORY.N1grams[event.S] = cnode_slot
		NODE_DIRECTORY.N1_top++ 
		return node_alloc_ptr
	case N2GRAM:
		cnode_slot = NODE_DIRECTORY.N2_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.N2directory = append(NODE_DIRECTORY.N2directory,event)
		NODE_DIRECTORY.N2grams[event.S] = cnode_slot
		NODE_DIRECTORY.N2_top++
		return node_alloc_ptr
	case N3GRAM:
		cnode_slot = NODE_DIRECTORY.N3_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.N3directory = append(NODE_DIRECTORY.N3directory,event)
		NODE_DIRECTORY.N3grams[event.S] = cnode_slot
		NODE_DIRECTORY.N3_top++
		return node_alloc_ptr
	case LT128:
		cnode_slot = NODE_DIRECTORY.LT128_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.LT128 = append(NODE_DIRECTORY.LT128,event)
		NODE_DIRECTORY.LT128_top++
		return node_alloc_ptr
	case LT1024:
		cnode_slot = NODE_DIRECTORY.LT1024_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.LT1024 = append(NODE_DIRECTORY.LT1024,event)
		NODE_DIRECTORY.LT1024_top++
		return node_alloc_ptr
	case GT1024:
		cnode_slot = NODE_DIRECTORY.GT1024_top
		node_alloc_ptr.CPtr = cnode_slot
		event.NPtr = node_alloc_ptr
		NODE_DIRECTORY.GT1024 = append(NODE_DIRECTORY.GT1024,event)
		NODE_DIRECTORY.GT1024_top++
		return node_alloc_ptr
	}

	return NO_NODE_PTR
}

//**************************************************************

func CheckExistingOrAltCaps(event Node,ErrFunc func(string)) (ClassedNodePtr,bool) {

	var cnode_slot ClassedNodePtr = -1
	var ok bool = false
	ignore_caps := false

	switch event.NPtr.Class {
	case N1GRAM:
		cnode_slot,ok = NODE_DIRECTORY.N1grams[event.S]
	case N2GRAM:
		cnode_slot,ok = NODE_DIRECTORY.N2grams[event.S]
	case N3GRAM:
		cnode_slot,ok = NODE_DIRECTORY.N3grams[event.S]
	case LT128:
		cnode_slot,ok = LinearFindText(NODE_DIRECTORY.LT128,event,ignore_caps)
	case LT1024:
		cnode_slot,ok = LinearFindText(NODE_DIRECTORY.LT1024,event,ignore_caps)
	case GT1024:
		cnode_slot,ok = LinearFindText(NODE_DIRECTORY.GT1024,event,ignore_caps)
	}

	if ok {
		return cnode_slot,ok
	} else {
		// Check for alternative caps

		ignore_caps = true
		alternative_caps := false
		
		switch event.NPtr.Class {
		case N1GRAM:
			for key := range NODE_DIRECTORY.N1grams {
				if strings.ToLower(key) == strings.ToLower(event.S) {
					alternative_caps = true
				}
			}
		case N2GRAM:
			for key := range NODE_DIRECTORY.N2grams {
				if strings.ToLower(key) == strings.ToLower(event.S) {
					alternative_caps = true
				}
			}
		case N3GRAM:
			for key := range NODE_DIRECTORY.N3grams {
				if strings.ToLower(key) == strings.ToLower(event.S) {
					alternative_caps = true
				}
			}

		case LT128:
			_,alternative_caps = LinearFindText(NODE_DIRECTORY.LT128,event,ignore_caps)
		case LT1024:
			_,alternative_caps = LinearFindText(NODE_DIRECTORY.LT1024,event,ignore_caps)
		case GT1024:
			_,alternative_caps = LinearFindText(NODE_DIRECTORY.GT1024,event,ignore_caps)
		}

		if alternative_caps {
			ErrFunc(WARN_DIFFERENT_CAPITALS+" ("+event.S+")")
		}

	}
	return cnode_slot,ok
}

//**************************************************************

func IdempAddChapterSeqToNode(class int,cptr ClassedNodePtr,chap string,seq bool) {

	/* In the DB version, we have handle chapter collisions
           we want all similar names to have a single node for lateral
           association, but we need to be able to search by chapter too,
           so merge the chapters as an attribute list */

	var node Node

	node = UpdateSeqStatus(class,cptr,seq)

	if strings.Contains(node.Chap,chap) {
		return
	}

	newchap := node.Chap + "," + chap

	switch class {
	case N1GRAM:
		NODE_DIRECTORY.N1directory[cptr].Chap = newchap
	case N2GRAM:
		NODE_DIRECTORY.N2directory[cptr].Chap = newchap
	case N3GRAM:
		NODE_DIRECTORY.N3directory[cptr].Chap = newchap
	case LT128:
		NODE_DIRECTORY.LT128[cptr].Chap = newchap
	case LT1024:
		NODE_DIRECTORY.LT1024[cptr].Chap = newchap
	case GT1024:
		NODE_DIRECTORY.GT1024[cptr].Chap = newchap
	}
}

//**************************************************************

func UpdateSeqStatus(class int,cptr ClassedNodePtr,seq bool) Node {

	switch class {
	case N1GRAM:
		NODE_DIRECTORY.N1directory[cptr].Seq = NODE_DIRECTORY.N1directory[cptr].Seq || seq
		return NODE_DIRECTORY.N1directory[cptr]
	case N2GRAM:
		NODE_DIRECTORY.N2directory[cptr].Seq = NODE_DIRECTORY.N2directory[cptr].Seq || seq
		return NODE_DIRECTORY.N2directory[cptr]
	case N3GRAM:
		NODE_DIRECTORY.N3directory[cptr].Seq = NODE_DIRECTORY.N3directory[cptr].Seq || seq
		return NODE_DIRECTORY.N3directory[cptr]
	case LT128:
		NODE_DIRECTORY.LT128[cptr].Seq = NODE_DIRECTORY.LT128[cptr].Seq || seq
		return NODE_DIRECTORY.LT128[cptr]
	case LT1024:
		NODE_DIRECTORY.LT1024[cptr].Seq = NODE_DIRECTORY.LT1024[cptr].Seq || seq
		return NODE_DIRECTORY.LT1024[cptr]
	case GT1024:
		NODE_DIRECTORY.GT1024[cptr].Seq = NODE_DIRECTORY.GT1024[cptr].Seq || seq
		return NODE_DIRECTORY.GT1024[cptr]
	}

	fmt.Println("Non existent node class (shouldn't happen)")
	os.Exit(-1)
	var dummy Node
	return dummy
}

//**************************************************************

func InsertArrowDirectory(stname,alias,name,pm string) ArrowPtr {

	// Insert an arrow into the forward/backward indices

	var newarrow ArrowDirectory

	// Check is already exists - harmless

	prev_alias,a_exists := ARROW_SHORT_DIR[alias]
	prev_name,n_exists := ARROW_LONG_DIR[name]

	if a_exists && n_exists {
		if prev_alias == prev_name {
			return prev_alias
		}
	}

	for a := range ARROW_DIRECTORY {
		if ARROW_DIRECTORY[a].Long == name || ARROW_DIRECTORY[a].Short == alias {
			return ArrowPtr(-1)
		}
	}

	newarrow.STAindex = GetSTIndexByName(stname,pm)
	newarrow.Long = name
	newarrow.Short = alias
	newarrow.Ptr = ARROW_DIRECTORY_TOP

	ARROW_DIRECTORY = append(ARROW_DIRECTORY,newarrow)
	ARROW_SHORT_DIR[alias] = ARROW_DIRECTORY_TOP
	ARROW_LONG_DIR[name] = ARROW_DIRECTORY_TOP
	ARROW_DIRECTORY_TOP++

	return ARROW_DIRECTORY_TOP-1
}

//**************************************************************

func InsertInverseArrowDirectory(fwd,bwd ArrowPtr) {

	if fwd == ArrowPtr(-1) || bwd == ArrowPtr(-1) {
		return
	}

	// Lookup inverse by long name, only need this in search presentation

	INVERSE_ARROWS[fwd] = bwd
	INVERSE_ARROWS[bwd] = fwd
}

//**************************************************************

func AppendLinkToNode(frptr NodePtr,link Link,toptr NodePtr) {

	frclass := frptr.Class
	frm := frptr.CPtr
	stindex := ARROW_DIRECTORY[link.Arr].STAindex

	link.Dst = toptr // fill in the last part of the reference

	// Idempotently add any new context strings to the current list
	// between from and to nodes -- stindex tells us which link type, so implicit in the arrow type
	// the empty arrow is used to record node context, which is type LEADSTO

	switch frclass {

	case N1GRAM:
		NODE_DIRECTORY.N1directory[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.N1directory[frm].I[stindex],link)
	case N2GRAM:
		NODE_DIRECTORY.N2directory[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.N2directory[frm].I[stindex],link)
	case N3GRAM:
		NODE_DIRECTORY.N3directory[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.N3directory[frm].I[stindex],link)
	case LT128:
		NODE_DIRECTORY.LT128[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.LT128[frm].I[stindex],link)
	case LT1024:
		NODE_DIRECTORY.LT1024[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.LT1024[frm].I[stindex],link)
	case GT1024:
		NODE_DIRECTORY.GT1024[frm].I[stindex] = MergeLinkLists(NODE_DIRECTORY.GT1024[frm].I[stindex],link)
	}
}

//**************************************************************

func MergeLinkLists(linklist []Link,lnk Link) []Link {

	// Ensure all arrows and contexts in lnk are in list for the appropriate arrows

	new_ctxstr := GetContext(lnk.Ctx)
	new_ctxlist := strings.Split(new_ctxstr,",")

	// Check if the arrow is already there to add to its context

	for l := range linklist {
		if linklist[l].Arr == lnk.Arr && linklist[l].Dst == lnk.Dst {

			already_ctxstr := GetContext(linklist[l].Ctx)
			already_ctxlist := strings.Split(already_ctxstr,",")

			linklist[l].Ctx = MergeContextLists(already_ctxlist,new_ctxlist)

			return linklist
		}
	}

	// if not already there, add this arrow

	linklist = append(linklist,lnk)
	return linklist
}

//**************************************************************

func MergeContextLists(one,two []string) ContextPtr {

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

	ctxptr,ok := CONTEXT_DIR[ctxstr]

	if ok {
		return ctxptr
	} else {
		var cd ContextDirectory
		cd.Context = ctxstr
		cd.Ptr = CONTEXT_TOP
		CONTEXT_DIRECTORY = append(CONTEXT_DIRECTORY,cd)
		CONTEXT_DIR[ctxstr] = CONTEXT_TOP
		ctxptr = CONTEXT_TOP
		CONTEXT_TOP++
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

