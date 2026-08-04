// **************************************************************************
//
// postgres_retrieval.go
//
// **************************************************************************

package sst

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

// **************************************************************************

func SolveNodePtrs(sst PoSST, nodenames []string, search SearchParameters, arr []ArrowPtr, limit int) []NodePtr {

	chap := search.Chapter
	cntx := search.Context
	seq := search.Sequence

	// This is a UI/UX wrapper for the underlying lookup, avoiding
	// duplicate results and ordering according to interest

	nodeptrs, rest := ParseLiteralNodePtrs(nodenames)

	var idempotence = make(map[NodePtr]bool)
	var result []NodePtr

	// If we give a precise reference, then that was obviously intended

	for n := range nodeptrs {
		idempotence[nodeptrs[n]] = true
	}

	for r := 0; r < len(rest); r++ {

		// Takes care of general context matching

		nptrs := GetDBNodePtrMatchingNCCS(sst, rest[r], chap, cntx, arr, seq, limit)

		for n := 0; n < len(nptrs); n++ {
			idempotence[nptrs[n]] = true
		}
	}

	// Currently disordered, sort by additional scoring by running context ..

	for uniqnptr := range idempotence {
		result = append(result, uniqnptr)
	}

	sort.Slice(result, ScoreContext)

	return result
}

//******************************************************************

func GetBookmarksFromDB(sst PoSST) []Bookmark {

	if sst.Q == nil {
		return nil
	}
	rows, err := sst.Q.ListBookmarks(sst.ctx())
	if err != nil {
		fmt.Println("QUERY BegBookmarksFromDB Failed", err)
		return nil
	}

	var chaps []string
	var sorts = make(map[string][]Bookmark)
	var retval []Bookmark

	for _, row := range rows {
		var b Bookmark
		b.Bookmark = derefStr(row.Bookmark)
		b.Query = derefStr(row.Query)

		line := strings.Split(b.Bookmark, ",")

		if len(line) > 1 {
			var bp Bookmark
			chapter := strings.TrimSpace(line[0])
			section := strings.TrimSpace(b.Bookmark[len(line[0])+1:])
			bp.Bookmark = section
			bp.Query = b.Query
			sorts[chapter] = append(sorts[chapter], bp)
		} else {
			sorts["misc"] = append(sorts["misc"], b)
		}
	}

	for key := range sorts {
		chaps = append(chaps, key)
	}

	sort.Strings(chaps)

	for _, k := range chaps {

		const dunbar_limit = 30
		var note_added string = ""

		if len(sorts[k]) > dunbar_limit {
			note_added = " ... (list exceeds Dunbar limit for learning)"
		}

		var bp Bookmark
		bp.Bookmark = k + note_added
		bp.Query = ""
		retval = append(retval, bp)

		sort.Slice(sorts[k], func(i, j int) bool {
			return sorts[k][i].Bookmark < sorts[k][j].Bookmark
		})

		for _, sorted := range sorts[k] {
			retval = append(retval, sorted)
		}
	}

	return retval
}

//******************************************************************

func GetDBNodePtrMatchingName(sst PoSST, name, chap string) []NodePtr {

	// simplified, retain for compatibility

	return GetDBNodePtrMatchingNCCS(sst, name, chap, nil, nil, false, CAUSAL_CONE_MAXLIMIT)
}

// **************************************************************************

func GetDBNodePtrMatchingNCCS(sst PoSST, nm, chap string, cn []string, arrow []ArrowPtr, seq bool, limit int) []NodePtr {

	if sst.Q == nil {
		return nil
	}

	arg := searchNodePtrsArg(sst, nodeSearchSpec{
		Name:    nm,
		Chap:    chap,
		Context: cn,
		Arrows:  arrow,
		Seq:     seq,
		Limit:   limit,
	})
	rows, err := sst.Q.SearchNodePtrs(sst.ctx(), arg)
	if err != nil {
		fmt.Println("QUERY GetNodePtrMatchingNCC Failed", err)
		return nil
	}

	retval := make([]NodePtr, 0, len(rows))
	for _, r := range rows {
		retval = append(retval, NodePtr{Class: int(r.Chan), CPtr: ClassedNodePtr(r.Cptr)})
	}
	return retval
}

// nodeSearchSpec groups search filters for SearchNodePtrs.
type nodeSearchSpec struct {
	Name    string
	Chap    string
	Context []string
	Arrows  []ArrowPtr
	Seq     bool
	Limit   int
}

// searchNodePtrsArg builds sqlc SearchNodePtrs params from a nodeSearchSpec.
func searchNodePtrsArg(sst PoSST, spec nodeSearchSpec) sqlc.SearchNodePtrsParams {
	name := SQLEscape(spec.Name)
	chap := SQLEscape(spec.Chap)

	anyChap := chap == "any" || chap == ""
	chapUnaccent := false
	chapPat := ""
	if !anyChap {
		rm, stripped := IsBracketedSearchTerm(chap)
		chapUnaccent = rm
		if rm {
			chapPat = "%" + stripped + "%"
		} else {
			chapPat = "%" + chap + "%"
		}
	}

	outerExact, nopling := IsExactMatch(name)
	rmName, nobrack := IsBracketedSearchTerm(nopling)
	innerExact, bare := IsExactMatch(nobrack)
	exact := outerExact || innerExact

	mode := "any"
	nameArg := bare
	excludePaths := !strings.HasPrefix(bare, "/")

	if name == "any" || name == "%%" {
		mode = "any"
		nameArg = ""
		excludePaths = false
	} else if exact {
		mode = "exact"
	} else if IsStringFragment(bare) {
		mode = "like"
		if name == "any" || name == "%%" {
			nameArg = "%%"
		} else {
			nameArg = "%" + bare + "%"
		}
	} else if rmName {
		mode = "ufts"
	} else {
		mode = "fts"
	}

	_, cnStripped := IsBracketedSearchList(spec.Context)
	if cnStripped == nil {
		cnStripped = []string{}
	}

	return sqlc.SearchNodePtrsParams{
		Column1:  anyChap,
		Column2:  chapUnaccent,
		Lower:    chapPat,
		Column4:  mode,
		Lower_2:  nameArg,
		Column6:  excludePaths,
		Column7:  spec.Seq,
		Column8:  cnStripped,
		Column9:  arrowToInt32s(spec.Arrows),
		Column10: toInt32s(GetSTtypesFromArrows(sst, spec.Arrows)),
		Limit:    int32(spec.Limit),
	}
}

// **************************************************************************

func GetDBChaptersMatchingName(sst PoSST, src string) []string {

	if sst.Q == nil {
		return nil
	}

	remove_accents, stripped := IsBracketedSearchTerm(SQLEscape(src))
	var (
		rows []*string
		err  error
	)
	if remove_accents {
		rows, err = sst.Q.ListChaptersLikeUnaccent(sst.ctx(), "%"+stripped+"%")
	} else {
		rows, err = sst.Q.ListChaptersLike(sst.ctx(), "%"+src+"%")
	}
	if err != nil {
		fmt.Println("QUERY GetDBChaptersMatchingName", err)
		return nil
	}

	chapters := make(map[string]int)
	var retval []string
	for _, p := range rows {
		whole := derefStr(p)
		for _, part := range strings.Split(whole, ",") {
			chapters[part]++
		}
	}
	for c := range chapters {
		if strings.Contains(c, src) && len(c) > 0 {
			retval = append(retval, c)
		}
	}
	sort.Strings(retval)
	return retval
}

// **************************************************************************

func GetDBContextByName(sst *PoSST, src string) (string, ContextPtr) {

	if sst.Q == nil {
		return "", 0
	}

	remove_accents, stripped := IsBracketedSearchTerm(src)
	var (
		row sqlc.Contextdirectory
		err error
	)
	if remove_accents {
		row, err = sst.Q.GetContextByTextUnaccent(sst.ctx(), stripped)
	} else {
		row, err = sst.Q.GetContextByText(sst.ctx(), strPtr(src))
	}
	if err != nil {
		return "", 0
	}
	return derefStr(row.Context), ContextPtr(row.Ctxptr)
}

// **************************************************************************

func GetDBContextByPtr(sst *PoSST, ptr ContextPtr) (string, ContextPtr) {

	if sst.Q == nil {
		return "", ptr
	}
	row, err := sst.Q.GetContextByPtr(sst.ctx(), int32(ptr))
	if err != nil {
		return "", ptr
	}
	return derefStr(row.Context), ContextPtr(row.Ctxptr)
}

// **************************************************************************

func GetSTtypesFromArrows(sst PoSST, arrows []ArrowPtr) []int {

	var sttypes []int

	for a := range arrows {
		sta := sst.ARROW_DIRECTORY[arrows[a]].STAindex
		st := STIndexToSTType(sta)
		sttypes = append(sttypes, st)
	}

	return sttypes
}

// **************************************************************************

func GetDBNodeByNodePtr(sst *PoSST, db_nptr NodePtr) Node {

	im_nptr, cached := sst.NODE_CACHE[db_nptr]

	if cached {
		return GetMemoryNodeFromPtr(sst, im_nptr)
	}

	var n Node
	if sst.Q == nil {
		return n
	}

	row, err := sst.Q.GetNodeByNPtr(sst.ctx(), sqlc.GetNodeByNPtrParams{
		Column1: int32(db_nptr.Class),
		Column2: int32(db_nptr.CPtr),
	})
	if err != nil {
		// no rows is normal for missing
		return n
	}

	if row.L != nil {
		n.L = int(*row.L)
	}
	n.S = derefStr(row.S)
	n.Chap = derefStr(row.Chap)
	whole := [ST_TOP]string{row.Im3, row.Im2, row.Im1, row.In0, row.Il1, row.Ic2, row.Ie3}
	for i := 0; i < ST_TOP; i++ {
		n.I[i] = ParseLinkArray(whole[i])
	}

	if strings.HasPrefix(n.S, "Dynamic: ") {
		n.S = ExpandDynamicFunctions(n.S)
	}

	n.NPtr = db_nptr
	return n
}

// **************************************************************************

func GetDBSingletonBySTType(sst PoSST, sttypes []int, chap string, cn []string) ([]NodePtr, []NodePtr) {

	// Used in graph report, analysis

	dim := len(sttypes)
	if dim == 0 || dim > 4 {
		fmt.Println("Maximum 4 sttypes in GetDBSingletonBySTType")
		return nil, nil
	}
	for _, st := range sttypes {
		if st < 0 {
			fmt.Println("WARNING! Only give positive STType arguments to GetDBSingletonBySTType as both signs are returned as sources (+) and sinks (-)")
			return nil, nil
		}
	}
	if sst.Q == nil {
		return nil, nil
	}

	_, cnStripped := IsBracketedSearchList(cn)
	if cnStripped == nil {
		cnStripped = []string{}
	}
	chapter := "%" + SQLEscape(chap) + "%"
	lead, cont, expr, near := stPosFlags(sttypes)

	srcRows, err := sst.Q.ListSingletonSources(sst.ctx(), sqlc.ListSingletonSourcesParams{
		Lower:   chapter,
		Column2: lead,
		Column3: cont,
		Column4: expr,
		Column5: near,
		Column6: cnStripped,
	})
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType Failed", err)
		return nil, nil
	}
	var src_nptrs, snk_nptrs []NodePtr
	for _, r := range srcRows {
		src_nptrs = append(src_nptrs, NodePtr{Class: int(r.Chan), CPtr: ClassedNodePtr(r.Cptr)})
	}

	snkRows, err := sst.Q.ListSingletonSinks(sst.ctx(), sqlc.ListSingletonSinksParams{
		Lower:   chapter,
		Column2: lead,
		Column3: cont,
		Column4: expr,
		Column5: near,
		Column6: cnStripped,
	})
	if err != nil {
		fmt.Println("QUERY GetDBSingletonBySTType 2 Failed", err)
		return nil, nil
	}
	for _, r := range snkRows {
		snk_nptrs = append(snk_nptrs, NodePtr{Class: int(r.Chan), CPtr: ClassedNodePtr(r.Cptr)})
	}
	return src_nptrs, snk_nptrs
}

// **************************************************************************

func SelectStoriesByArrow(sst *PoSST, nodeptrs []NodePtr, arrowptrs []ArrowPtr, sttypes []int, limit int) []NodePtr {

	var matches []NodePtr

	// Need to take each arrow type at a time. We can't possibly know if an
	// intentionally promised sequence start (in Node) refers to one arrow or another,
	// but, the chance of being a start for several different independent stories is unlikely.

	// We can always search for ad hoc cases with dream/post-processing if not from N4L
	// Thus a valid story is defined from a start node. It is normally a node with an out-arrow
	// |- NODE --ARROW-->, i.e. no in-arrow entering, but this may be false if the story has
	// loops, like a repeated line in a song chorus.

	for _, n := range nodeptrs {

		// After changes, all these nodes should have Seq = true already from "SolveNodePtrs()"
		// So all the searching is finished, we just need to match the requested arrow

		node := GetDBNodeByNodePtr(sst, n) // we are now caching this for later
		matches = append(matches, node.NPtr)
	}

	return matches
}

// **************************************************************************

func GetSequenceContainers(sst *PoSST, nodeptrs []NodePtr, arrowptrs []ArrowPtr, sttypes []int, limit int) []Story {

	// Story search

	var stories []Story

	openings := SelectStoriesByArrow(sst, nodeptrs, arrowptrs, sttypes, limit)

	arrname := ""
	count := 0

	var already = make(map[NodePtr]bool)

	for nth := range openings {

		var story Story

		node := GetDBNodeByNodePtr(sst, openings[nth])

		story.Chapter = node.Chap

		axis := GetLongestAxialPath(sst, openings[nth], arrowptrs[0], limit)

		directory := AssignStoryCoordinates(axis, nth, len(openings), limit, already)

		for lnk := 0; lnk < len(axis); lnk++ {

			// Now add the orbit at this node, not including the axis

			var ne NodeEvent

			nd := GetDBNodeByNodePtr(sst, axis[lnk].Dst)

			ne.Text = nd.S
			ne.L = nd.L
			ne.Chap = nd.Chap
			ne.Context = GetContext(sst, axis[lnk].Ctx)
			ne.NPtr = axis[lnk].Dst
			ne.XYZ = directory[ne.NPtr]
			ne.Orbits = GetNodeOrbit(sst, axis[lnk].Dst, arrname, limit)
			ne.Orbits = SetOrbitCoords(ne.XYZ, ne.Orbits)

			if lnk > limit {
				break
			}

			story.Axis = append(story.Axis, ne)
		}

		if story.Axis != nil {
			stories = append(stories, story)
			count++
		}

		count++

		if count > limit {
			return stories
		}

	}

	return stories
}

// **************************************************************************

func GetDBArrowsWithArrowName(sst *PoSST, s string) (ArrowPtr, int) {

	if sst.ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	s = strings.Trim(s, "!")

	if s == "" {
		fmt.Println("No such arrow found in database:", s)
		return 0, 0
	}

	for a := range sst.ARROW_DIRECTORY {
		if s == sst.ARROW_DIRECTORY[a].Long || s == sst.ARROW_DIRECTORY[a].Short {
			sttype := STIndexToSTType(sst.ARROW_DIRECTORY[a].STAindex)
			return sst.ARROW_DIRECTORY[a].Ptr, sttype
		}
	}

	fmt.Println("No such arrow found in database:", s)
	return 0, 0
}

// **************************************************************************

func GetDBArrowsMatchingArrowName(sst *PoSST, s string) []ArrowPtr {

	var list []ArrowPtr

	if sst.ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	trimmed := strings.Trim(s, "!")

	if trimmed == "" {
		return list
	}

	if trimmed != s {
		for a := range sst.ARROW_DIRECTORY {
			if sst.ARROW_DIRECTORY[a].Long == trimmed || sst.ARROW_DIRECTORY[a].Short == trimmed {
				list = append(list, sst.ARROW_DIRECTORY[a].Ptr)
			}
		}
	} else {
		for a := range sst.ARROW_DIRECTORY {
			if SimilarString(sst.ARROW_DIRECTORY[a].Long, s) || SimilarString(sst.ARROW_DIRECTORY[a].Short, s) {
				list = append(list, sst.ARROW_DIRECTORY[a].Ptr)
			}
		}
	}

	return list
}

// **************************************************************************

func GetDBArrowByName(sst *PoSST, name string) ArrowPtr {

	if sst.ARROW_DIRECTORY_TOP == 0 {
		DownloadArrowsFromDB(sst)
	}

	name = strings.Trim(name, "!")

	if name == "" {
		return 0
	}

	ptr, ok := sst.ARROW_SHORT_DIR[name]

	// If not, then check longname

	if !ok {
		ptr, ok = sst.ARROW_LONG_DIR[name]

		if !ok {
			ptr, ok = sst.ARROW_SHORT_DIR[name]

			// If not, then check longname

			if !ok {
				ptr, ok = sst.ARROW_LONG_DIR[name]
				fmt.Println(ERR_NO_SUCH_ARROW, "("+name+") - no arrows defined in database yet?")
				return 0
			}
		}
	}

	return ptr
}

// **************************************************************************

func GetDBArrowByPtr(sst *PoSST, arrowptr ArrowPtr) ArrowDirectory {

	if int(arrowptr) > len(sst.ARROW_DIRECTORY) {
		DownloadArrowsFromDB(sst)
	}

	if int(arrowptr) < len(sst.ARROW_DIRECTORY) {
		a := sst.ARROW_DIRECTORY[arrowptr]
		return a
	} else {
		return sst.ARROW_DIRECTORY[0]
	}

	return sst.ARROW_DIRECTORY[arrowptr]

}

// **************************************************************************

func GetDBArrowBySTType(sst PoSST, sttype int) []ArrowDirectory {

	var retval []ArrowDirectory

	DownloadArrowsFromDB(&sst)

	for a := range sst.ARROW_DIRECTORY {
		sta := sst.ARROW_DIRECTORY[a].STAindex
		if STIndexToSTType(sta) == sttype {
			retval = append(retval, sst.ARROW_DIRECTORY[a])
		}
	}

	return retval
}

//******************************************************************

func ArrowPtrFromArrowsNames(sst *PoSST, arrows []string) ([]ArrowPtr, []int) {

	// Parse input and discern arrow types, best guess

	var arr []ArrowPtr
	var stt []int

	for a := range arrows {

		// is the entry a number? sttype?

		number, err := strconv.Atoi(arrows[a])
		notnumber := err != nil

		if notnumber {
			arrs := GetDBArrowsMatchingArrowName(sst, arrows[a])
			for ar := range arrs {
				arrowptr := arrs[ar]
				if arrowptr > 0 {
					arrdir := GetDBArrowByPtr(sst, arrowptr)
					arr = append(arr, arrdir.Ptr)
					stt = append(stt, STIndexToSTType(arrdir.STAindex))
				}
			}
		} else {
			if number < -EXPRESS {
				fmt.Println("Negative arrow value doesn't make sense", number)
			} else if number >= -EXPRESS && number <= EXPRESS {
				stt = append(stt, number)
			} else {
				// whatever remains can only be an arrowpointer
				arrdir := GetDBArrowByPtr(sst, ArrowPtr(number))
				arr = append(arr, arrdir.Ptr)
				stt = append(stt, STIndexToSTType(arrdir.STAindex))
			}
		}
	}

	return arr, stt
}

// **************************************************************************

func GetAppointedNodesByArrow(sst *PoSST, arrow ArrowPtr, cn []string, chap string, size int) map[ArrowPtr][]Appointment {

	// return a map of all the nodes in chap,context that are pointed to by the same type of arrow
	// grouped by arrow

	if sst.Q == nil {
		return nil
	}

	reverse_arrow := sst.INVERSE_ARROWS[arrow]
	arr := GetDBArrowByPtr(sst, reverse_arrow)
	sttype := STIndexToSTType(arr.STAindex)

	_, cn_stripped := IsBracketedSearchList(cn)
	if cn_stripped == nil {
		cn_stripped = []string{}
	}

	var chap_col, chap_stripped string
	var remove_chap_accents bool

	if chap != "any" && chap != "" {
		remove_chap_accents, chap_stripped = IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_col = "%" + chap_stripped + "%"
		} else {
			chap_col = "%" + chap + "%"
		}
	}

	rows, err := sst.Q.GetAppointments(sst.ctx(), sqlc.GetAppointmentsParams{
		Column1: int32(reverse_arrow),
		Column2: int32(sttype),
		Column3: int32(size),
		Column4: chap_col,
		Column5: cn_stripped,
		Column6: remove_chap_accents,
	})
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesByArrow Failed", err)
		return nil
	}
	retval := make(map[ArrowPtr][]Appointment)
	for _, whole := range rows {
		next := ParseAppointedNodeCluster(sst, whole)
		retval[next.Arr] = append(retval[next.Arr], next)
	}
	return retval
}

// **************************************************************************

func GetAppointedNodesBySTType(sst *PoSST, sttype int, cn []string, chap string, size int) map[ArrowPtr][]Appointment {

	// return a map of all the nodes in chap,context that are pointed to by the same type of arrow
	// grouped by arrow

	if sst.Q == nil {
		return nil
	}

	_, cn_stripped := IsBracketedSearchList(cn)
	if cn_stripped == nil {
		cn_stripped = []string{}
	}

	var chap_col, chap_stripped string
	var remove_chap_accents bool

	if chap != "any" && chap != "" {
		remove_chap_accents, chap_stripped = IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_col = "%" + chap_stripped + "%"
		} else {
			chap_col = "%" + chap + "%"
		}
	}

	rows, err := sst.Q.GetAppointments(sst.ctx(), sqlc.GetAppointmentsParams{
		Column1: -1,
		Column2: int32(sttype),
		Column3: int32(size),
		Column4: chap_col,
		Column5: cn_stripped,
		Column6: remove_chap_accents,
	})
	if err != nil {
		fmt.Println("QUERY GetAppointedNodesBySTType Failed", err)
		return nil
	}
	retval := make(map[ArrowPtr][]Appointment)
	for _, whole := range rows {
		next := ParseAppointedNodeCluster(sst, whole)
		retval[next.Arr] = append(retval[next.Arr], next)
	}
	return retval
}

// **************************************************************************

func ParseAppointedNodeCluster(sst *PoSST, whole string) Appointment {

	//  (13,-1,maze,{},"(1,3122)","{""(1,3121)"",""(1,3138)""}")

	var next Appointment
	var l []string

	whole = strings.Trim(whole, "(")
	whole = strings.Trim(whole, ")")

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
			items = append(items, string(item))
			item = nil
			continue
		}

		item = append(item, uni_array[u])
	}

	if item != nil {
		items = append(items, string(item))
	}

	for i := range items {

		s := strings.TrimSpace(items[i])

		l = append(l, s)
	}

	var arrp ArrowPtr
	fmt.Sscanf(l[0], "%d", &arrp)
	fmt.Sscanf(l[1], "%d", &next.STType)

	// invert arrow
	next.Arr = sst.INVERSE_ARROWS[ArrowPtr(arrp)]
	next.STType = -next.STType

	next.Chap = l[2]
	next.Ctx = ParseSQLArrayString(l[3])

	fmt.Sscanf(l[4], "(%d,%d)", &next.NTo.Class, &next.NTo.CPtr)

	// Postgres is inconsistent in adding \" to arrays (hack)

	l[5] = strings.Replace(l[5], "(", "\"(", -1)
	l[5] = strings.Replace(l[5], ")", ")\"", -1)
	next.NFrom = ParseSQLNPtrArray(l[5])

	return next
}

//******************************************************************

func ScoreContext(i, j int) bool {

	// the more matching items the more relevant

	return true
}

// **************************************************************************

func GetDBPageMap(sst PoSST, chap string, cn []string, page int, limit int) []PageMap {

	if sst.Q == nil {
		return nil
	}

	chap = strings.Trim(chap, "\"")
	_, cnStripped := IsBracketedSearchList(cn)
	if cnStripped == nil {
		cnStripped = []string{}
	}
	chapter := "%" + chap + "%"
	hits_per_page := limit
	offset := (page - 1) * hits_per_page
	if offset < 0 {
		offset = 0
	}

	rows, err := sst.Q.ListPageMap(sst.ctx(), sqlc.ListPageMapParams{
		Column1: cnStripped,
		Lower:   chapter,
		Offset:  int32(offset),
		Limit:   int32(hits_per_page),
	})
	if err != nil {
		fmt.Println("GetDBPageMap Failed:", err)
		return nil
	}

	var pagemap []PageMap
	for _, row := range rows {
		var event PageMap
		event.Path = ParseMapLinkArray(row.Path)
		event.Chapter = derefStr(row.Chap)
		event.Context = ContextPtr(derefInt32(row.Ctx))
		event.Line = derefInt32(row.Line)
		pagemap = append(pagemap, event)
	}
	return pagemap
}

// **************************************************************************

func GetFwdConeAsNodes(sst *PoSST, start NodePtr, sttype, depth int, limit int) []NodePtr {

	if sst.Q == nil {
		return nil
	}
	rows, err := sst.Q.FwdConeAsNodes(sst.ctx(), sqlc.FwdConeAsNodesParams{
		Column1: int32(start.Class),
		Column2: int32(start.CPtr),
		Column3: int32(sttype),
		Column4: int32(depth),
		Column5: int32(limit),
	})
	if err != nil {
		fmt.Println("QUERY to FwdConeAsNodes Failed", err)
		return nil
	}
	retval := make([]NodePtr, 0, len(rows))
	for _, r := range rows {
		retval = append(retval, NodePtr{Class: int(r.Chan), CPtr: ClassedNodePtr(r.Cptr)})
	}
	return retval
}

// **************************************************************************

func GetFwdConeAsLinks(sst *PoSST, start NodePtr, sttype, depth int) []Link {

	// This function may be misleading as it doesn't respect paths, may be deprecated in future

	if sst.Q == nil {
		return nil
	}
	// maxlimit mirrors depth when callers omit a separate limit (upstream 3-arg form).
	rows, err := sst.Q.FwdConeAsLinks(sst.ctx(), sqlc.FwdConeAsLinksParams{
		Column1: int32(start.Class),
		Column2: int32(start.CPtr),
		Column3: int32(sttype),
		Column4: int32(depth),
		Column5: int32(depth),
	})
	if err != nil {
		fmt.Println("QUERY to FwdConeAsLinks Failed", err)
		return nil
	}
	retval := make([]Link, 0, len(rows))
	for _, whole := range rows {
		retval = append(retval, ParseSQLLinkString(whole))
	}
	return retval
}

// **************************************************************************

func GetFwdPathsAsLinks(sst *PoSST, start NodePtr, sttype, depth int, maxlimit int) ([][]Link, int) {

	if sst.Q == nil {
		return nil, 0
	}
	whole, err := sst.Q.FwdPathsAsLinks(sst.ctx(), sqlc.FwdPathsAsLinksParams{
		Column1: int32(start.Class),
		Column2: int32(start.CPtr),
		Column3: int32(sttype),
		Column4: int32(depth),
		Column5: int32(maxlimit),
	})
	if err != nil {
		fmt.Println("QUERY to FwdPathsAsLinks Failed", err)
		return nil, 0
	}
	retval := ParseLinkPath(whole)
	return retval, len(retval)
}

// **************************************************************************

func GetEntireConePathsAsLinks(sst *PoSST, orientation string, start NodePtr, depth int, limit int) ([][]Link, int) {

	// orientation should be "fwd" or "bwd" else "both"

	if sst.Q == nil {
		return nil, 0
	}
	whole, err := sst.Q.AllPathsAsLinks(sst.ctx(), sqlc.AllPathsAsLinksParams{
		Column1: int32(start.Class),
		Column2: int32(start.CPtr),
		Column3: orientation,
		Column4: int32(depth),
		Column5: int32(limit),
	})
	if err != nil {
		fmt.Println("QUERY to AllPathsAsLinks Failed", err)
		return nil, 0
	}
	retval := ParseLinkPath(whole)

	sort.Slice(retval, func(i, j int) bool {
		return len(retval[i]) < len(retval[j])
	})

	return retval, len(retval)
}

// **************************************************************************

func GetEntireNCConePathsAsLinks(sst *PoSST, orientation string, start []NodePtr, depth int, chapter string, context []string, limit int) ([][]Link, int) {

	// See also GetConstraintConePathsAsLinks for an interface with arrow matching
	// orientation should be "fwd" or "bwd" else "both"

	if sst.Q == nil {
		return nil, 0
	}

	remove_accents, stripped := IsBracketedSearchTerm(chapter)
	chapter = "%" + stripped + "%"
	if context == nil {
		context = []string{}
	}

	whole, err := sst.Q.AllNCPathsAsLinks(sst.ctx(), sqlc.AllNCPathsAsLinksParams{
		Column1: nodePtrArrayLiteral(start),
		Column2: chapter,
		Column3: remove_accents,
		Column4: context,
		Column5: orientation,
		Column6: int32(depth),
		Column7: int32(limit),
	})
	if err != nil {
		fmt.Println("QUERY to AllNCPathsAsLinks Failed", err)
		os.Exit(-1)
	}
	retval := ParseLinkPath(whole)
	return retval, len(retval)
}

// **************************************************************************

func GetConstraintConePathsAsLinks(sst *PoSST, start []NodePtr, depth int, chapter string, context []string, arrowptrs []ArrowPtr, sttypes []int, limit int) ([][]Link, int) {

	// See also GetEntireNCConePathsAsLinks() for a differently optimized interface
	// orientation should be "fwd" or "bwd" else "both"

	if sst.Q == nil {
		return nil, 0
	}

	remove_accents, stripped := IsBracketedSearchTerm(chapter)
	chapter = "%" + stripped + "%"
	if context == nil {
		context = []string{}
	}

	whole, err := sst.Q.ConstraintPathsAsLinks(sst.ctx(), sqlc.ConstraintPathsAsLinksParams{
		Column1: nodePtrArrayLiteral(start),
		Column2: chapter,
		Column3: remove_accents,
		Column4: context,
		Column5: toInt32s(Arrow2Int(arrowptrs)),
		Column6: toInt32s(sttypes),
		Column7: int32(depth),
		Column8: int32(limit),
	})
	if err != nil {
		fmt.Println("QUERY to ConstraintPathsAsLinks Failed", err)
		os.Exit(-1)
	}
	retval := ParseLinkPath(whole)
	return retval, len(retval)
}

//
// postgres_retrieval.go
//
