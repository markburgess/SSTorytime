// **************************************************************************
//
// postgres_retrieval.go
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"sort"
	_ "github.com/lib/pq"

)

// **************************************************************************

func SolveNodePtrs(sst PoSST,nodenames []string,search SearchParameters,arr []ArrowPtr,limit int) []NodePtr {

	chap := search.Chapter
	cntx := search.Context
	seq := search.Sequence

	// This is a UI/UX wrapper for the underlying lookup, avoiding
	// duplicate results and ordering according to interest

	nodeptrs,rest := ParseLiteralNodePtrs(nodenames)

	var idempotence = make(map[NodePtr]bool)
	var result []NodePtr

	// If we give a precise reference, then that was obviously intended

	for n := range nodeptrs {
		idempotence[nodeptrs[n]] = true
	}

	for r := 0; r < len(rest); r++ {

		// Takes care of general context matching

		nptrs := GetDBNodePtrMatchingNCCS(sst,rest[r],chap,cntx,arr,seq,limit)

		for n := 0; n < len(nptrs); n++ {
			idempotence[nptrs[n]] = true
		}
	}

	// Currently disordered, sort by additional scoring by running context ..

	for uniqnptr := range idempotence {
		result = append(result,uniqnptr)
	}

	sort.Slice(result, ScoreContext)

	return result
}

//******************************************************************

func GetDBNodePtrMatchingName(sst PoSST,name,chap string) []NodePtr {

	// simplified, retain for compatibility

	return GetDBNodePtrMatchingNCCS(sst,name,chap,nil,nil,false,CAUSAL_CONE_MAXLIMIT)
}

// **************************************************************************

func GetDBNodePtrMatchingNCCS(sst PoSST,nm,chap string,cn []string,arrow []ArrowPtr,seq bool,limit int) []NodePtr {

	// Order by L to favour exact matches

	nm = SQLEscape(nm)
	chap = SQLEscape(chap)

	qstr := fmt.Sprintf("SELECT NPtr FROM Node WHERE %s ORDER BY L ASC,(CARDINALITY(Ie3)+CARDINALITY(Im3)+CARDINALITY(Il1)) DESC LIMIT %d",NodeWhereString(nm,chap,cn,arrow,seq),limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY GetNodePtrMatchingNCC Failed",err,qstr)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
			retval = append(retval,n)
		}

		row.Close()
	}

	return retval
}

// **************************************************************************

func NodeWhereString(name,chap string,context []string,arrow []ArrowPtr,seq bool) string {

	var chap_col, nm_col string
	var ctx_col string
	var qstr string

	// Format a WHERE clause for a Node search satisfying constraints

	// Chapter first to limit search by block

	if chap != "any" && chap != "" {

		remove_chap_accents,chap_stripped := IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_search := "%"+chap_stripped+"%"
			chap_col = fmt.Sprintf("lower(unaccent(Chap)) LIKE lower('%s')",chap_search)
		} else {
			chap_search := "%"+chap+"%"
			chap_col = fmt.Sprintf("lower(Chap) LIKE lower('%s')",chap_search)
		}
	} else {
		chap_col = "true"
	}

	// Name search using tsquery for wildcards and additional S = exact_constraint for !exact!

	outer_exact_match,nopling := IsExactMatch(name)
	remove_name_accents,nobrack := IsBracketedSearchTerm(nopling)
	inner_exact_match,bare_name := IsExactMatch(nobrack)

	is_exact_match := outer_exact_match || inner_exact_match

	// First ignore technical references from ad hoc search results, like img paths

	if !strings.HasPrefix(bare_name,"/") {
		nm_col = "AND S NOT LIKE '/%'"
	}

	if is_exact_match {

		nm_col += fmt.Sprintf(" AND lower(S) = '%s'",bare_name)

	} else if IsStringFragment(bare_name) {

		if name == "any" || name == "%%" {
			nm_col = fmt.Sprintf(" AND lower(S) LIKE '%%%%'")
		} else {
			nm_col = fmt.Sprintf(" AND lower(S) LIKE '%%%s%%'",bare_name)
		}
	} else {

		if name == "any" || name == "%%" {
			nm_col = ""
		} else {
			if remove_name_accents {
				nm_col = fmt.Sprintf(" AND Unsearch @@ to_tsquery('english', '%s')",bare_name)
			} else {
				nm_col = fmt.Sprintf(" AND Search @@ to_tsquery('english', '%s')",bare_name)
			}
		}
	}

        var seq_col string
        
        if seq {
                seq_col = "AND Seq=true"
        }

	// context and arrows

	_,cn_stripped := IsBracketedSearchList(context)
	ctx_col = FormatSQLStringArray(cn_stripped)

	arrows := FormatSQLIntArray(Arrow2Int(arrow))
	sttypes := FormatSQLIntArray(GetSTtypesFromArrows(arrow))

	dbcols := I_MEXPR+","+I_MCONT+","+I_MLEAD+","+I_NEAR +","+I_PLEAD+","+I_PCONT+","+I_PEXPR

	qstr = fmt.Sprintf("%s %s %s AND NCC_match(NPtr,%s,%s,%s,%s)",
		chap_col,nm_col,seq_col, ctx_col,arrows,sttypes,dbcols)

	return qstr
}

// **************************************************************************

func GetDBChaptersMatchingName(sst PoSST,src string) []string {

	var qstr string

	remove_accents,stripped := IsBracketedSearchTerm(SQLEscape(src))

	if remove_accents {
		search := "%"+stripped+"%"
		qstr = fmt.Sprintf("SELECT DISTINCT Chap FROM Node WHERE lower(unaccent(Chap)) LIKE lower('%s')",search)
	} else {
		search := "%"+src+"%"
		qstr = fmt.Sprintf("SELECT DISTINCT Chap FROM Node WHERE lower(Chap) LIKE lower('%s')",search)
	}

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBChaptersMatchingName",err)
	}

	var whole string
	var chapters = make(map[string]int)
	var retval []string

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			several := strings.Split(whole,",")
			
			for s := range several {
				chapters[several[s]]++
			}
		}

		for c := range chapters {
			if strings.Contains(c,src) {
				if len(c) > 0 {
					retval = append(retval,c)
				}
			}
		}

		sort.Strings(retval)
		row.Close()
	}

	return retval
}

// **************************************************************************

func GetDBContextByName(sst PoSST,src string) (string,ContextPtr) {

	var qstr string

	remove_accents,stripped := IsBracketedSearchTerm(src)

	if remove_accents {
		search := stripped
		qstr = fmt.Sprintf("SELECT DISTINCT Context,CtxPtr FROM ContextDirectory WHERE unaccent(Context)='%s'",search)
	} else {
		search := src
		qstr = fmt.Sprintf("SELECT DISTINCT Context,CtxPtr FROM ContextDirectory WHERE Context='%s'",search)
	}

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY GetDBContextByName",err)
	}

	var whole string
	var ptr int

	// Assume unique match for this, to be fixed elsewhere

	if row != nil {
		for row.Next() {
			err = row.Scan(&whole,&ptr)
		}
		row.Close()
	}

	return whole,ContextPtr(ptr)

}

// **************************************************************************

func GetDBContextByPtr(sst PoSST,ptr ContextPtr) (string,ContextPtr) {

	qstr := fmt.Sprintf("SELECT DISTINCT Context,CtxPtr FROM ContextDirectory WHERE CtxPtr=%d",ptr)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBContextssByPtr",err)
	}

	var retctx string
	var retptr int

	// Assume unique match for this, to be fixed elsewhere

	if row != nil {
		for row.Next() {
			err = row.Scan(&retctx,&retptr)
		}

		row.Close()
	}

	return retctx,ContextPtr(retptr)
}

// **************************************************************************

func GetSTtypesFromArrows(arrows []ArrowPtr) []int {

	var sttypes []int

	for a := range arrows {
		sta := ARROW_DIRECTORY[arrows[a]].STAindex
		st := STIndexToSTType(sta)
		sttypes = append(sttypes,st)
	}

	return sttypes
}

// **************************************************************************

func GetDBNodeByNodePtr(sst PoSST,db_nptr NodePtr) Node {

	im_nptr,cached := NODE_CACHE[db_nptr]

	if cached {
		return GetMemoryNodeFromPtr(sst,im_nptr)
	}

	// This ony works if we insert non-null arrays like '[]' during initialization
	cols := I_MEXPR+","+I_MCONT+","+I_MLEAD+","+I_NEAR +","+I_PLEAD+","+I_PCONT+","+I_PEXPR
	qstr := fmt.Sprintf("select L,S,Chap,%s from Node where NPtr='(%d,%d)'::NodePtr AND NOT L=0",cols,db_nptr.Class,db_nptr.CPtr)

	row, err := sst.DB.Query(qstr)

	var n Node
	var count int = 0

	if err != nil {
		fmt.Println("GetDBNodeByNodePointer Failed:",err)
		return n
	}

	var whole [ST_TOP]string

	// NB, there seems to be a "bug" in the SQL package, which cannot always populate the links, so try not to
	//     rely on this and work around when needed using GetEntireCone(any,2..) separately

	if row != nil {
		for row.Next() {
			err = row.Scan(&n.L,&n.S,&n.Chap,&whole[0],&whole[1],&whole[2],&whole[3],&whole[4],&whole[5],&whole[6])

			for i := 0; i < ST_TOP; i++ {
				n.I[i] = ParseLinkArray(whole[i])
			}
			count++
		}

		if count > 1 {
			fmt.Println("GetDBNodeByNodePtr returned too many matches (multi-model conflict?):",count,"for ptr",db_nptr)
			os.Exit(-1)
		}

		// Expand any dynamic inbuilt functions

		if strings.HasPrefix(n.S,"Dynamic: ") {
			n.S = ExpandDynamicFunctions(n.S)
		}

		row.Close()

		if !cached {
			CacheNode(sst,n)
		}
	}

	n.NPtr = db_nptr
	return n
}

// **************************************************************************

func GetDBSingletonBySTType(sst PoSST,sttypes []int,chap string,cn []string) ([]NodePtr,[]NodePtr) {

	// Used in graph report, analysis

	var qstr,qwhere string
	var dim = len(sttypes)

	context := FormatSQLStringArray(cn)
	chapter := "%"+SQLEscape(chap)+"%"

	if dim == 0 || dim > 4 {
		fmt.Println("Maximum 4 sttypes in GetDBSingletonBySTType")
		return nil,nil
	}

	for st := 0; st < len(sttypes); st++ {

		if sttypes[st] < 0 {
			fmt.Println("WARNING! Only give positive STType arguments to GetDBSingletonBySTType as both signs are returned as sources (+) and sinks (-)")
			return nil,nil
		}

		stname := STTypeDBChannel(sttypes[st])
		stinv := STTypeDBChannel(-sttypes[st])
		qwhere += fmt.Sprintf("(array_length(%s::text[],1) IS NOT NULL AND array_length(%s::text[],1) IS NULL AND match_context((%s)[0].Ctx,%s))",stname,stinv,stname,context)
		
		if st != dim-1 {
			qwhere += " OR "
		}
	}

	qstr = fmt.Sprintf("SELECT NPtr FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",chapter,qwhere)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType Failed",err,"IN",qstr)
		return nil,nil
	}

	var src_nptrs,snk_nptrs []NodePtr

	if row != nil {
		for row.Next() {		
		
			var n NodePtr
			var nstr string
			
			err = row.Scan(&nstr)
		
			if err != nil {
				fmt.Println("Error scanning sql data case",dim,"gave error",err,qstr)
				row.Close()
				return nil,nil
			}
		
			fmt.Sscanf(nstr,"(%d,%d)",&n.Class,&n.CPtr)
		
			src_nptrs = append(src_nptrs,n)
		}
		row.Close()
	}

	// and sinks  -> -

	qwhere = ""

	for st := 0; st < len(sttypes); st++ {

		stname := STTypeDBChannel(-sttypes[st])
		stinv := STTypeDBChannel(sttypes[st])
		qwhere += fmt.Sprintf("(array_length(%s::text[],1) IS NOT NULL AND array_length(%s::text[],1) IS NULL AND match_context((%s)[0].Ctx,%s))",stname,stinv,stname,context)
		
		if st != dim-1 {
			qwhere += " OR "
		}
	}

	qstr = fmt.Sprintf("SELECT NPtr FROM Node WHERE lower(Chap) LIKE lower('%s') AND (%s)",chapter,qwhere)

	row, err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType 2 Failed",err,"IN",qstr)
		return nil,nil
	}

	if row != nil {
		for row.Next() {		
		
			var n NodePtr
			var nstr string
			
			err = row.Scan(&nstr)
		
			if err != nil {
				fmt.Println("Error scanning sql data case",dim,"gave error",err,qstr)
				row.Close()
				return nil,nil
			}
			
			fmt.Sscanf(nstr,"(%d,%d)",&n.Class,&n.CPtr)
			
			snk_nptrs = append(snk_nptrs,n)
		}
		row.Close()
	}

	return src_nptrs,snk_nptrs
	
}

// **************************************************************************

func SelectStoriesByArrow(sst PoSST,nodeptrs []NodePtr, arrowptrs []ArrowPtr, sttypes []int, limit int) []NodePtr {

	var matches []NodePtr

	// Need to take each arrow type at a time. We can't possibly know if an
	// intentionally promised sequence start (in Node) refers to one arrow or another,
	// but, the chance of being a start for several different independent stories is unlikely.

	// We can always search for ad hoc cases with dream/post-processing if not from N4L
	// Thus a valid story is defined from a start node. It is normally a node with an out-arrow
	// |- NODE --ARROW-->, i.e. no in-arrow entering, but this may be false if the story has
	// loops, like a repeated line in a song chorus.

	for _,n := range nodeptrs {

		// After changes, all these nodes should have Seq = true already from "SolveNodePtrs()"
		// So all the searching is finished, we just need to match the requested arrow

		node := GetDBNodeByNodePtr(sst,n)  // we are now caching this for later
		matches = append(matches,node.NPtr)
	}

	return matches
}

// **************************************************************************

func GetSequenceContainers(sst PoSST,nodeptrs []NodePtr, arrowptrs []ArrowPtr, sttypes []int, limit int) []Story {

	// Story search

	var stories []Story

	openings := SelectStoriesByArrow(sst,nodeptrs,arrowptrs,sttypes,limit)

	arrname := ""
	count := 0

	var already = make(map[NodePtr]bool)

	for nth := range openings {

		var story Story

		node := GetDBNodeByNodePtr(sst,openings[nth])

		story.Chapter = node.Chap

		axis := GetLongestAxialPath(sst,openings[nth],arrowptrs[0],limit)

		directory := AssignStoryCoordinates(axis,nth,len(openings),limit,already)

		for lnk := 0; lnk < len(axis); lnk++ {
			
			// Now add the orbit at this node, not including the axis

			var ne NodeEvent

			nd := GetDBNodeByNodePtr(sst,axis[lnk].Dst)

			ne.Text = nd.S
			ne.L = nd.L
			ne.Chap = nd.Chap
			ne.Context = GetContext(axis[lnk].Ctx)
			ne.NPtr = axis[lnk].Dst
			ne.XYZ = directory[ne.NPtr]
			ne.Orbits = GetNodeOrbit(sst,axis[lnk].Dst,arrname,limit)
			ne.Orbits = SetOrbitCoords(ne.XYZ,ne.Orbits)

			if lnk > limit {
				break
			}

			story.Axis = append(story.Axis,ne)
		}

		if story.Axis != nil {
			stories = append(stories,story)
			count ++
		}

		count++

		if count > limit {
			return stories
		}
		
	}

	return stories
}

// **************************************************************************

func GetDBArrowsWithArrowName(sst PoSST,s string) (ArrowPtr,int) {

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	s = strings.Trim(s,"!")

	if s == "" {
		fmt.Println("No such arrow found in database:",s)
		return 0,0
	}

	for a := range ARROW_DIRECTORY {
		if s == ARROW_DIRECTORY[a].Long || s == ARROW_DIRECTORY[a].Short {
			sttype := STIndexToSTType(ARROW_DIRECTORY[a].STAindex)
			return ARROW_DIRECTORY[a].Ptr,sttype
		}
	}

	fmt.Println("No such arrow found in database:",s)
	return 0,0
}

// **************************************************************************

func GetDBArrowsMatchingArrowName(sst PoSST,s string) []ArrowPtr {

	var list []ArrowPtr

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	trimmed := strings.Trim(s,"!")

	if trimmed == "" {
		return list
	}

	if trimmed != s {
		for a := range ARROW_DIRECTORY {
			if ARROW_DIRECTORY[a].Long==trimmed || ARROW_DIRECTORY[a].Short==trimmed {
				list = append(list,ARROW_DIRECTORY[a].Ptr)
			}
		}
	} else {
		for a := range ARROW_DIRECTORY {
			if SimilarString(ARROW_DIRECTORY[a].Long,s) || SimilarString(ARROW_DIRECTORY[a].Short,s) {
				list = append(list,ARROW_DIRECTORY[a].Ptr)
			}
		}
	}

	return list
}

// **************************************************************************

func GetDBArrowByName(sst PoSST,name string) ArrowPtr {

	if ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	name = strings.Trim(name,"!")

	if name == "" {
		return 0
	}

	ptr, ok := ARROW_SHORT_DIR[name]
	
	// If not, then check longname
	
	if !ok {
		ptr, ok = ARROW_LONG_DIR[name]
		
		if !ok {
			ptr, ok = ARROW_SHORT_DIR[name]
			
			// If not, then check longname
			
			if !ok {
				ptr, ok = ARROW_LONG_DIR[name]
				fmt.Println(ERR_NO_SUCH_ARROW,"("+name+") - no arrows defined in database yet?")
				return 0
			}
		}
	}

	return ptr
}

// **************************************************************************

func GetDBArrowByPtr(sst PoSST,arrowptr ArrowPtr) ArrowDirectory {

	if int(arrowptr) > len(ARROW_DIRECTORY) {
		DownloadArrowsFromDB(sst)
	}

	if int(arrowptr) < len(ARROW_DIRECTORY) {
		a := ARROW_DIRECTORY[arrowptr]
		return a
	} else {
		return ARROW_DIRECTORY[0]
	}
		
	return ARROW_DIRECTORY[arrowptr]

}

// **************************************************************************

func GetDBArrowBySTType(sst PoSST,sttype int) []ArrowDirectory {

	var retval []ArrowDirectory

	DownloadArrowsFromDB(sst)

	for a := range ARROW_DIRECTORY {
		sta := ARROW_DIRECTORY[a].STAindex
		if STIndexToSTType(sta) == sttype {
			retval = append(retval,ARROW_DIRECTORY[a])
		}
	}

	return retval
}

//******************************************************************

func ArrowPtrFromArrowsNames(sst PoSST,arrows []string) ([]ArrowPtr,[]int) {

	// Parse input and discern arrow types, best guess

	var arr []ArrowPtr
	var stt []int

	for a := range arrows {

		// is the entry a number? sttype?

		number, err := strconv.Atoi(arrows[a])
		notnumber := err != nil

		if notnumber {
			arrs := GetDBArrowsMatchingArrowName(sst,arrows[a])
			for  ar := range arrs {
				arrowptr := arrs[ar]
				if arrowptr > 0 {
					arrdir := GetDBArrowByPtr(sst,arrowptr)
					arr = append(arr,arrdir.Ptr)
					stt = append(stt,STIndexToSTType(arrdir.STAindex))
				}
			}
		} else {
			if number < -EXPRESS {
				fmt.Println("Negative arrow value doesn't make sense",number)
			} else if number >= -EXPRESS && number <= EXPRESS {
				stt = append(stt,number)
			} else {
				// whatever remains can only be an arrowpointer
				arrdir := GetDBArrowByPtr(sst,ArrowPtr(number))
				arr = append(arr,arrdir.Ptr)
				stt = append(stt,STIndexToSTType(arrdir.STAindex))
			}
		}
	}

	return arr,stt
}

// **************************************************************************

func GetAppointedNodesByArrow(sst PoSST,arrow ArrowPtr,cn []string,chap string,size int) map[ArrowPtr][]Appointment {

	// return a map of all the nodes in chap,context that are pointed to by the same type of arrow
        // grouped by arrow

	reverse_arrow := INVERSE_ARROWS[arrow]
	arr := GetDBArrowByPtr(sst,reverse_arrow)
	sttype := STIndexToSTType(arr.STAindex)

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	var chap_col,chap_stripped string
	var remove_chap_accents bool

	if chap != "any" && chap != "" {	
		remove_chap_accents,chap_stripped = IsBracketedSearchTerm(chap)
		
		if remove_chap_accents {
			chap_col = "%"+chap_stripped+"%"
		} else {
			chap_col = "%"+chap+"%"
		}
	}

	qstr := fmt.Sprintf("SELECT unnest(GetAppointments(%d,%d,%d,'%s',%s,%v))",int(reverse_arrow),sttype,size,chap_col,context,remove_chap_accents)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesByArrow Failed",err,qstr)
	}

	var whole string

	var retval = make(map[ArrowPtr][]Appointment)
	
	if row != nil {
		for row.Next() {
			err = row.Scan(&whole) //arrint,&sttype,&rchap,&rctx,&apex,&arry)

			next := ParseAppointedNodeCluster(whole)
			retval[next.Arr] = append(retval[next.Arr],next)
		}
	
		row.Close()
	}

	return retval
}

// **************************************************************************

func GetAppointedNodesBySTType(sst PoSST,sttype int,cn []string,chap string,size int) map[ArrowPtr][]Appointment {

	// return a map of all the nodes in chap,context that are pointed to by the same type of arrow
        // grouped by arrow

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	var chap_col,chap_stripped string
	var remove_chap_accents bool

	if chap != "any" && chap != "" {	
		remove_chap_accents,chap_stripped = IsBracketedSearchTerm(chap)
		
		if remove_chap_accents {
			chap_col = "%"+chap_stripped+"%"
		} else {
			chap_col = "%"+chap+"%"
		}
	}

	qstr := fmt.Sprintf("SELECT unnest(GetAppointments(%d,%d,%d,'%s',%s,%v))",-1,sttype,size,chap_col,context,remove_chap_accents)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesByArrow Failed",err,qstr)
	}

	var whole string

	var retval = make(map[ArrowPtr][]Appointment)
	
	if row != nil {
		for row.Next() {
			err = row.Scan(&whole) //arrint,&sttype,&rchap,&rctx,&apex,&arry)

			next := ParseAppointedNodeCluster(whole)
			retval[next.Arr] = append(retval[next.Arr],next)
		}	
		row.Close()
	}

	return retval
}

// **************************************************************************

func ParseAppointedNodeCluster(whole string) Appointment {

    //  (13,-1,maze,{},"(1,3122)","{""(1,3121)"",""(1,3138)""}")

	var next Appointment
      	var l []string

    	whole = strings.Trim(whole,"(")
    	whole = strings.Trim(whole,")")

	uni_array := []rune(whole)

	var items []string
	var item []rune
	var protected = false

	for u := range uni_array {

		if uni_array[u] == '"' {
			protected = !protected
			continue
		}

		if !protected && uni_array[u] == ',' {
			items = append(items,string(item))
			item = nil
			continue
		}

		item = append(item,uni_array[u])
	}

	if item != nil {
		items = append(items,string(item))
	}

	for i := range items {

	    s := strings.TrimSpace(items[i])

	    l = append(l,s)
	    }

	var arrp ArrowPtr
	fmt.Sscanf(l[0],"%d",&arrp)
	fmt.Sscanf(l[1],"%d",&next.STType)

	// invert arrow
	next.Arr = INVERSE_ARROWS[ArrowPtr(arrp)]
	next.STType = -next.STType

	next.Chap = l[2]
	next.Ctx = ParseSQLArrayString(l[3])

	fmt.Sscanf(l[4],"(%d,%d)",&next.NTo.Class,&next.NTo.CPtr)

	// Postgres is inconsistent in adding \" to arrays (hack)

	l[5] = strings.Replace(l[5],"(","\"(",-1)
	l[5] = strings.Replace(l[5],")",")\"",-1)
	next.NFrom = ParseSQLNPtrArray(l[5])

	return next
}

//******************************************************************

func ScoreContext(i,j int) bool {

	// the more matching items the more relevant

	return true
}

// **************************************************************************

func GetDBPageMap(sst PoSST,chap string,cn []string,page int) []PageMap {

	var qstr string

	chap = strings.Trim(chap,"\"")

	context := FormatSQLStringArray(cn)
	chapter := "%"+chap+"%"

	const hits_per_page = 60
	offset := (page-1) * hits_per_page;

	qstr = fmt.Sprintf("SELECT DISTINCT Chap,Ctx,Line,Path FROM PageMap\n"+
		"WHERE match_context(Ctx,%s)=true AND lower(Chap) LIKE lower('%s') ORDER BY Chap,Line OFFSET %d LIMIT %d",context,chapter,offset,hits_per_page)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("GetDBPageMap Failed:",err,qstr)
	}

	var path string
	var pagemap []PageMap
	var line int
	var ctxptr ContextPtr

	if row != nil {
		for row.Next() {		

			var event PageMap

			err = row.Scan(&chap,&ctxptr,&line,&path)

			if err != nil {
				fmt.Println("Error reading GetDBPageMap",err)
			}

			event.Path = ParseMapLinkArray(path)

			event.Chapter = chap
			event.Context = ctxptr
			event.Line = line;

			pagemap = append(pagemap,event)
		}

		row.Close()
	}

	return pagemap
}

// **************************************************************************

func GetFwdConeAsNodes(sst PoSST, start NodePtr, sttype,depth int,limit int) []NodePtr {

	qstr := fmt.Sprintf("select unnest(fwdconeasnodes) from FwdConeAsNodes('(%d,%d)',%d,%d,%d);",start.Class,start.CPtr,sttype,depth,limit)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdConeAsNodes Failed",err)
	}

	var whole string
	var n NodePtr
	var retval []NodePtr

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			fmt.Sscanf(whole,"(%d,%d)",&n.Class,&n.CPtr)
			retval = append(retval,n)
		}

		row.Close()
	}

	return retval
}

// **************************************************************************

func GetFwdConeAsLinks(sst PoSST, start NodePtr, sttype,depth int) []Link {

	// This function may be misleading as it doesn't respect paths, may be deprecated in future

	qstr := fmt.Sprintf("select unnest(fwdconeaslinks) from FwdConeAsLinks('(%d,%d)',%d,%d);",start.Class,start.CPtr,sttype,depth)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdConeAsLinks Failed",err)
	}

	var whole string
	var retval []Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			l := ParseSQLLinkString(whole)
			retval = append(retval,l)
		}

		row.Close()
	}

	return retval
}

// **************************************************************************

func GetFwdPathsAsLinks(sst PoSST, start NodePtr, sttype,depth int, maxlimit int) ([][]Link,int) {

	qstr := fmt.Sprintf("SELECT FwdPathsAsLinks from FwdPathsAsLinks('(%d,%d)',%d,%d,%d);",start.Class,start.CPtr,sttype,depth,maxlimit)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY to FwdPathsAsLinks Failed",err)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
		}

		row.Close()
	}

	return retval,len(retval)
}

// **************************************************************************

func GetEntireConePathsAsLinks(sst PoSST,orientation string,start NodePtr,depth int,limit int) ([][]Link,int) {

	// orientation should be "fwd" or "bwd" else "both"

	// Todo: how to limit path search? Usually solutions are small..?

	qstr := fmt.Sprintf("select AllPathsAsLinks from AllPathsAsLinks('(%d,%d)','%s',%d, %d);",
		start.Class,start.CPtr,orientation,depth,limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllPathsAsLinks Failed",err,qstr)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
		}

		row.Close()
	}

	sort.Slice(retval, func(i,j int) bool {
		return len(retval[i]) < len(retval[j])
	})

	return retval,len(retval)
}

// **************************************************************************

func GetEntireNCConePathsAsLinks(sst PoSST,orientation string,start []NodePtr,depth int,chapter string,context []string,limit int) ([][]Link,int) {

	// See also GetConstraintConePathsAsLinks for an interface with arrow matching
	// orientation should be "fwd" or "bwd" else "both"

	remove_accents,stripped := IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	qstr := fmt.Sprintf("select AllNCPathsAsLinks(%s,'%s',%s,%s,'%s',%d,%d);",FormatSQLNodePtrArray(start),chapter,rm_acc,FormatSQLStringArray(context),orientation,depth,limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to AllNCPathsAsLinks Failed",err,qstr)
		os.Exit(-1)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
		}

		row.Close()
	}

	return retval,len(retval)
}

// **************************************************************************

func GetConstraintConePathsAsLinks(sst PoSST,start []NodePtr,depth int,chapter string,context []string,arrowptrs []ArrowPtr,sttypes []int,limit int) ([][]Link,int) {

	// See also GetEntireNCConePathsAsLinks() for a differently optimized interface
	// orientation should be "fwd" or "bwd" else "both"

	remove_accents,stripped := IsBracketedSearchTerm(chapter)
	chapter = "%"+stripped+"%"
	rm_acc := "false"

	if remove_accents {
		rm_acc = "true"
	}

	nod := FormatSQLNodePtrArray(start)
	arr := FormatSQLIntArray(Arrow2Int(arrowptrs))
	stt := FormatSQLIntArray(sttypes)
	cnt := FormatSQLStringArray(context)

	qstr := fmt.Sprintf("select ConstraintPathsAsLinks(%s,'%s',%s,%s,%s,%s,%d,%d);",nod,chapter,rm_acc,cnt,arr,stt,depth,limit)

	row, err := sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("QUERY to ConstraintPathsAsLinks Failed",err,qstr)
		os.Exit(-1)
	}

	var whole string
	var retval [][]Link

	if row != nil {
		for row.Next() {		
			err = row.Scan(&whole)
			retval = ParseLinkPath(whole)
			break
		}

		row.Close()
	}

	return retval,len(retval)
}



//
// postgres_retrieval.go
//

