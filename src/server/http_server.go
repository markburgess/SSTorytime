//******************************************************************
//
//  Web server for lookup requests and JSON interface
//
//******************************************************************

package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	SST "SSTorytime"
)

// Ugly Go directive to embed text files into the binary

//go:embed all:public
var content embed.FS

// *********************************************************************

var PSST SST.PoSST // just one persistent connection

// *********************************************************************
// Main
// *********************************************************************

func main() {

	PSST = SST.Open(true)

	// 1. Create the filesystem view rooted inside the "public" directory.

	publicFS, err := fs.Sub(content, "public")

	if err != nil {
		log.Fatal("failed to create sub-filesystem:", err)
	}

	// 2. Create a router (ServeMux) and register handlers.

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(publicFS))

	mux.Handle("/", fileServer)
	mux.HandleFunc("/searchN4L", SearchN4LHandler)
	mux.HandleFunc("/status", StatusHandler)

	// 3. Create an http.Server instance for graceful shutdown.

	srv := &http.Server{Addr:    "0.0.0.0:8080", Handler: EnableCORS(mux), }

	// 4. Run the server in a goroutine so it doesn't block.

	go func() {
		log.Println("Server starting on http://localhost:8080")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not start server: %s\n", err)
		}
	}()

	// 5. Wait for an interrupt signal.

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// 6. Perform a graceful shutdown with a timeout.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %s\n", err)
	}

	log.Println("Server exited properly")
}

// *********************************************************************
// Handlers
// *********************************************************************

func EnableCORS(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set the Access-Control-Allow-Origin header to the origin of the request.

		origin := r.Header.Get("Origin")

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Browsers send a pre-flight OPTIONS request for CORS. We need to handle it.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// *********************************************************************
// Handlers
// *********************************************************************

func SignalHandler() {

	signal_chan := make(chan os.Signal, 1)

	signal.Notify(signal_chan,
		syscall.SIGHUP,  // 1
		syscall.SIGINT,  // 2 ctrl-c
		syscall.SIGQUIT, // 3
		syscall.SIGTERM) // 15, CTRL-c

	sig := <-signal_chan // block until signal

	switch sig {

	case syscall.SIGHUP:
		fmt.Println("hungup")

	case syscall.SIGINT:
		fmt.Println("Warikomi, cutting in, sandoichi")

	case syscall.SIGTERM:
		fmt.Println("force stop")

	case syscall.SIGQUIT:
		fmt.Println("stop and core dump")

	default:
		fmt.Println("Unknown signal.")
	}
}

// *********************************************************************
// SEARCH
// *********************************************************************

func SearchN4LHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case "POST", "GET":
		name := r.FormValue("name")
		nclass := r.FormValue("nclass")
		ncptr := r.FormValue("ncptr")
		chapcontext := r.FormValue("chapcontext")

		if name == "\\lastnptr" {
			if chapcontext != "" && chapcontext != "any" {
				UpdateLastSawSection(w, r, chapcontext)
			}
			UpdateLastSawNPtr(w, r, nclass, ncptr, chapcontext)
			return
		}

		if name == "" && len(nclass) > 0 && len(ncptr) > 0 {
			// direct click on an item
			var a, b int
			fmt.Sscanf(nclass, "%d", &a)
			fmt.Sscanf(ncptr, "%d", &b)
			nstr := fmt.Sprintf("(%d,%d)", a, b)
			name = name + nstr
		}

		ambient, key, _ := SST.GetTimeContext()

		if len(name) == 0 || name == "\\remind" {
			name = "any \\chapter reminders \\context any, " + key + " " + ambient + " \\limit 20"
		}

		if name == "\\help" {
			name = "\\notes \\chapter \"help and search\" \\limit 40"
		}

		fmt.Println("\nReceived command:", name)

		search := SST.DecodeSearchField(name)

		HandleSearch(search, name, w, r)

	default:
		http.Error(w, "Not supported", http.StatusMethodNotAllowed)
	}
}

// *********************************************************************

func UpdateLastSawSection(w http.ResponseWriter, r *http.Request, query string) {

	// update lastseen db

	fmt.Println("UPDATING STATS FOR section", query)

	SST.UpdateLastSawSection(PSST, query)
}

// *********************************************************************

func UpdateLastSawNPtr(w http.ResponseWriter, r *http.Request, class, cptr string, classifier string) {

	// update lastseen db

	var nptr SST.NodePtr
	var nclass int
	var ncptr int
	fmt.Sscanf(class, "%d", &nclass)
	fmt.Sscanf(cptr, "%d", &ncptr)
	nptr.Class = nclass
	nptr.CPtr = SST.ClassedNodePtr(ncptr)

	SST.UpdateLastSawNPtr(PSST, nclass, ncptr, classifier)

	fmt.Println("UPDATING STATS FOR", nclass, ncptr, "WITHIN", classifier)

	SST.UpdateLastSawSection(PSST, classifier)

	response := fmt.Sprintf("{ \"Response\" : \"LastSaw\",\n \"Content\" : \"ack(%s,%s)\" }", class, cptr)
	w.Write([]byte(response))

}

// *********************************************************************

func HandleSearch(search SST.SearchParameters, line string, w http.ResponseWriter, r *http.Request) {

	// This is analogous to searchN4L

	// OPTIONS *********************************************

	name := search.Name != nil
	from := search.From != nil
	to := search.To != nil
	context := search.Context != nil
	chapter := search.Chapter != ""
	pagenr := search.PageNr > 0
	sequence := search.Sequence

	// Now convert strings into NodePointers

	arrowptrs, sttype := SST.ArrowPtrFromArrowsNames(PSST, search.Arrows)

	arrows := arrowptrs != nil
	sttypes := sttype != nil
	limit := 0

	if search.Range > 0 {
		limit = search.Range
	} else {
		if from || to || sequence {
			limit = 30 // many paths make hard work
		} else {
			const common_word = 5

			if SST.SearchTermLen(search.Name) < common_word {
				limit = 5
			} else {
				limit = 10
			}
		}
	}

	fmt.Println()
	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tabWriter, "start set:\t", SL(search.Name))
	fmt.Fprintln(tabWriter, "from:\t", SL(search.From))
	fmt.Fprintln(tabWriter, "to:\t", SL(search.To))
	fmt.Fprintln(tabWriter, "chapter:\t", search.Chapter)
	fmt.Fprintln(tabWriter, "context:\t", SL(search.Context))
	fmt.Fprintln(tabWriter, "arrows:\t", SL(search.Arrows))
	fmt.Fprintln(tabWriter, "pageNR:\t", search.PageNr)
	fmt.Fprintln(tabWriter, "sequence/story:\t", search.Sequence)
	fmt.Fprintln(tabWriter, "limit/range/depth:\t", limit)
	fmt.Fprintln(tabWriter, "show stats:\t", search.Stats)

	tabWriter.Flush()
	fmt.Println()

	var nodeptrs, leftptrs, rightptrs []SST.NodePtr

	if !pagenr && !sequence {
		leftptrs = SST.SolveNodePtrs(PSST, search.From, search, arrowptrs, limit)
		rightptrs = SST.SolveNodePtrs(PSST, search.To, search, arrowptrs, limit)
	}

	nodeptrs = SST.SolveNodePtrs(PSST, search.Name, search, arrowptrs, limit)

	fmt.Println("Solved search nodes ...")

	// SEARCH SELECTION *********************************************

	// Table of contents

	if search.Stats {
		ShowStats(w, r, PSST, search, nodeptrs)
		return
	}

	if (context || chapter) && !name && !sequence && !pagenr && !(from || to) {
		ShowChapterContexts(w, r, PSST, search, limit)
		return
	}

	if name && !sequence && !pagenr {
		HandleOrbit(w, r, PSST, search, nodeptrs, limit)
		return
	}

	if (name && from) || (name && to) {
		fmt.Printf("\nSearch \"%s\" has conflicting parts <to|from> and match strings\n", line)
		os.Exit(-1)
	}

	// Closed path solving, two sets of nodeptrs
	// if we have BOTH from/to (maybe with chapter/context) then we are looking for paths

	if from && to {
		HandlePathSolve(w, r, PSST, leftptrs, rightptrs, search, arrowptrs, sttype, limit)
		return
	}

	// Open causal cones, from one of these three

	if (name || from || to) && !pagenr && !sequence {

		if nodeptrs != nil {
			HandleCausalCones(w, r, PSST, nodeptrs, search, arrowptrs, sttype, limit)
			return
		}
		if leftptrs != nil {
			HandleCausalCones(w, r, PSST, leftptrs, search, arrowptrs, sttype, limit)
			return
		}
		if rightptrs != nil {
			HandleCausalCones(w, r, PSST, rightptrs, search, arrowptrs, sttype, limit)
			return
		}
	}

	// if we have page number then we are looking for notes by pagemap

	if (name || chapter) && pagenr {

		var notes []SST.PageMap

		if chapter {
			notes = SST.GetDBPageMap(PSST, search.Chapter, search.Context, search.PageNr)
			HandlePageMap(w, r, PSST, search, notes)
			return
		} else {
			for n := range search.Name {
				notes = SST.GetDBPageMap(PSST, search.Name[n], search.Context, search.PageNr)
				HandlePageMap(w, r, PSST, search, notes)
			}
			return
		}
	}

	// Look for axial trails following a particular arrow, like _sequence_

	if sequence {
		HandleStories(w, r, PSST, search, nodeptrs, arrowptrs, sttype, limit)
		return
	}

	// if we have sequence with arrows, then we are looking for sequence context or stories

	if arrows || sttypes {
		HandleMatchingArrows(w, r, PSST, search, arrowptrs, sttype)
		return
	}

	fmt.Println("Didn't find a solver")
}

// *********************************************************************

func HandleOrbit(w http.ResponseWriter, r *http.Request, sst SST.PoSST, search SST.SearchParameters, nptrs []SST.NodePtr, limit int) {

	var count int
	var array []SST.NodeEvent

	origin := SST.Coords{X: 0.0, Y: 0.0, Z: 0.0}

	for n := 0; n < len(nptrs); n++ {

		count++

		if count > limit {
			break
		}

		fmt.Printf("Assembling Node Orbit(%v)\n", nptrs[n])

		orb := SST.GetNodeOrbit(PSST, nptrs[n], "", limit)
		// create a set of coords for len(nptrs) disconnected nodes

		xyz := SST.RelativeOrbit(origin, SST.R0, n, len(nptrs))
		orb = SST.SetOrbitCoords(xyz, orb)

		nodeevent := SST.JSONNodeEvent(PSST, nptrs[n], xyz, orb)
		array = append(array, nodeevent)
	}

	data, _ := json.Marshal(array)
	response := PackageResponse(sst, search, "Orbits", string(data))

	//fmt.Println("REPLY:\n",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Reply Orbit sent")
}

// *********************************************************************

func HandleCausalCones(w http.ResponseWriter, r *http.Request, sst SST.PoSST, nptrs []SST.NodePtr, search SST.SearchParameters, arrows []SST.ArrowPtr, sttype []int, limit int) {

	chap := search.Chapter
	context := search.Context

	fmt.Println("HandleCausalCones()", nptrs)
	var total int = 1

	if len(sttype) == 0 {
		sttype = []int{0, 1, 2, 3}
	}

	var cones []SST.WebConePaths

	for n := range nptrs {
		for st := range sttype {

			subcone, count := PackageConeFromOrigin(sst, nptrs[n], n, sttype[st], chap, context, len(nptrs), limit)
			cones = append(cones, subcone)

			total += count

			if total > limit {
				break
			}
		}

		if total > limit {
			break
		}
	}

	array, _ := json.Marshal(cones)

	response := PackageResponse(sst, search, "ConePaths", string(array))
	//fmt.Println("CasualConePath reponse",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent cone")
}

//******************************************************************

func PackageConeFromOrigin(sst SST.PoSST, nptr SST.NodePtr, nth int, sttype int, chap string, context []string, dimnptr, limit int) (SST.WebConePaths, int) {

	// Package a JSON object for the nth/dimnptr causal cone , assigning each nth the same width

	var wpaths [][]SST.WebPath

	fcone, count := SST.GetFwdPathsAsLinks(PSST, nptr, sttype, limit, limit)
	wpaths = append(wpaths, SST.LinkWebPaths(PSST, fcone, nth, chap, context, dimnptr, limit)...)

	if sttype != 0 {
		bcone, countb := SST.GetFwdPathsAsLinks(PSST, nptr, -sttype, limit, limit)
		wpaths = append(wpaths, SST.LinkWebPaths(PSST, bcone, nth, chap, context, dimnptr, limit)...)
		count += countb
	}

	var subcone SST.WebConePaths
	subcone.RootNode = nptr
	subcone.Title = SST.GetDBNodeByNodePtr(sst, nptr).S
	subcone.Paths = wpaths

	return subcone, count
}

//******************************************************************

func HandlePathSolve(w http.ResponseWriter, r *http.Request, sst SST.PoSST, leftptrs, rightptrs []SST.NodePtr, search SST.SearchParameters, arrowptrs []SST.ArrowPtr, sttype []int, maxdepth int) {

	chapter := search.Chapter
	context := search.Context

	fmt.Println("HandlePathSolve(", leftptrs, ",", rightptrs, ")")

	solutions := SST.GetPathsAndSymmetries(sst,leftptrs,rightptrs,chapter,context,arrowptrs,maxdepth)

	if len(solutions) > 0 {
		// format paths
		
		var pack []SST.WebConePaths
		var soln SST.WebConePaths
		
		soln.RootNode = solutions[0][0].Dst
		soln.Title = fmt.Sprintf("paths solutions from %v to %v",search.From,search.To)
		soln.BTWC = SST.BetweenNessCentrality(sst, solutions)
		soln.SuperNodes = SST.SuperNodes(sst, solutions, maxdepth)
		
		var wpaths [][]SST.WebPath
		nth := 0
		swimlanes := 1
		
		wpaths = append(wpaths, SST.LinkWebPaths(sst, solutions, nth, chapter, context, swimlanes, maxdepth)...)
		
		soln.Paths = wpaths
		pack = append(pack, soln)
		array_pack, _ := json.Marshal(pack)
		
		response := PackageResponse(sst, search, "PathSolve", string(array_pack))
		
		//fmt.Println("PATH SOLVE:",string(response))
		
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
		return
	}

	fmt.Println("No paths satisfy constraints")
	response := PackageResponse(sst, search, "PathSolve", "[]")

	//fmt.Println("PATHSOLVE NOTES",string(response))
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent path solve")
}

//******************************************************************

func HandlePageMap(w http.ResponseWriter, r *http.Request, sst SST.PoSST, search SST.SearchParameters, notes []SST.PageMap) {

	fmt.Println("Solver/handler: HandlePageMap()")

	jstr := SST.JSONPage(PSST, notes)
	response := PackageResponse(sst, search, "PageMap", jstr)

	if notes != nil {
		UpdateLastSawSection(w, r, notes[0].Chapter)
	}

	//fmt.Println("PAGEMAP NOTES",string(response))
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent pagemap")
}

//******************************************************************

func HandleStories(w http.ResponseWriter, r *http.Request, sst SST.PoSST, search SST.SearchParameters, nodeptrs []SST.NodePtr, arrowptrs []SST.ArrowPtr, sttypes []int, limit int) {

	if arrowptrs == nil {
		arrowptrs, sttypes = SST.ArrowPtrFromArrowsNames(PSST, []string{"!then!"})
	}

	fmt.Println("Solver/handler: HandleStories()")

	stories := SST.GetSequenceContainers(sst, nodeptrs, arrowptrs, sttypes, limit)

	jarray := ""

	for s := 0; s < len(stories); s++ {

		var jstory string

		for a := 0; a < len(stories[s].Axis); a++ {
			jstr := JSONStoryNodeEvent(stories[s].Axis[a])
			jstory += fmt.Sprintf("%s,", jstr)
		}

		jstory = strings.Trim(jstory, ",")
		jarray = fmt.Sprintf("[%s],", jstory)
	}

	jarray = strings.Trim(jarray, ",")

	response := PackageResponse(sst, search, "Sequence", jarray)

	//fmt.Println("Sequence...",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent sequence")

}

// *********************************************************************

func HandleMatchingArrows(w http.ResponseWriter, r *http.Request, sst SST.PoSST, search SST.SearchParameters, arrowptrs []SST.ArrowPtr, sttype []int) {

	fmt.Println("Solver/handler: HandleMatchingArrows()")

	type ArrowList struct {
		ArrPtr  SST.ArrowPtr
		ASTtype int
		Short   string
		Long    string
		InvPtr  SST.ArrowPtr
		ISTtype int
		InvS    string
		InvL    string
	}

	var arrows []ArrowList

	for a := range arrowptrs {
		adir := SST.GetDBArrowByPtr(sst, arrowptrs[a])
		inv := SST.GetDBArrowByPtr(sst, SST.INVERSE_ARROWS[arrowptrs[a]])

		var al ArrowList
		al.ArrPtr = arrowptrs[a]
		al.ASTtype = SST.STIndexToSTType(adir.STAindex)
		al.Short = adir.Short
		al.Long = adir.Long
		al.InvPtr = inv.Ptr
		al.ISTtype = SST.STIndexToSTType(inv.STAindex)
		al.InvS = inv.Short
		al.InvL = inv.Long
		arrows = append(arrows, al)
	}

	if arrowptrs == nil {
		for st := range sttype {
			adirs := SST.GetDBArrowBySTType(sst, sttype[st])
			for adir := range adirs {
				inv := SST.GetDBArrowByPtr(sst, SST.INVERSE_ARROWS[adirs[adir].Ptr])

				var al ArrowList
				al.ArrPtr = adirs[adir].Ptr
				al.ASTtype = SST.STIndexToSTType(adirs[adir].STAindex)
				al.Short = adirs[adir].Short
				al.Long = adirs[adir].Long
				al.InvPtr = inv.Ptr
				al.ISTtype = SST.STIndexToSTType(inv.STAindex)
				al.InvS = inv.Short
				al.InvL = inv.Long
				arrows = append(arrows, al)
			}
		}
	}

	data, _ := json.Marshal(arrows)
	response := PackageResponse(sst, search, "Arrows", string(data))

	fmt.Println("Arrows...", string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent arrows")
}

// *********************************************************************

func ShowStats(w http.ResponseWriter, r *http.Request, sst SST.PoSST, search SST.SearchParameters, nptrs []SST.NodePtr) {

	var retval []SST.LastSeen

	if nptrs == nil {
		retval = SST.GetLastSawSection(sst)
	} else {

		for n := range nptrs {
			nptr := SST.GetLastSawNPtr(sst, nptrs[n])
			retval = append(retval, nptr)
		}
	}

	data, _ := json.Marshal(retval)

	response := PackageResponse(sst, search, "STAT", string(data))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent stat")

}

// *********************************************************************

func ShowChapterContexts(w http.ResponseWriter, r *http.Request, sst SST.PoSST, search SST.SearchParameters, limit int) {

	chap := search.Chapter
	context := search.Context

	fmt.Println("Solver/handler: ShowChapterContexts()")

	var chapters []SST.ChCtx
	var chap_list []string

	toc := SST.GetChaptersByChapContext(sst, chap, context, limit)

	for chaps := range toc {
		chap_list = append(chap_list, chaps)
	}

	sort.Strings(chap_list)

	for c := 0; c < len(chap_list); c++ {

		var chap_anchor SST.ChCtx

		chap_anchor.Chapter = chap_list[c]
		chap_anchor.XYZ = SST.AssignChapterCoordinates(c, len(chap_list))

		// Fractionate the (chapter,context) information

		dim, clist, adj := SST.IntersectContextParts(toc[chap_list[c]])
		spectrum := SST.GetContextTokenFrequencies(toc[chap_list[c]])
		intent, ambient := SST.ContextIntentAnalysis(spectrum, toc[chap_list[c]])

		chap_anchor.Context = GetContextSets(dim, clist, adj, chap_anchor.XYZ)
		chap_anchor.Single = GetContextFragments(intent, chap_anchor.XYZ)
		chap_anchor.Common = GetContextFragments(ambient, chap_anchor.XYZ)

		chapters = append(chapters, chap_anchor)
	}

	data, _ := json.Marshal(chapters)
	response := PackageResponse(sst, search, "TOC", string(data))

	//fmt.Println("Chap/context...", string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent content")
}

//******************************************************************

func GetContextSets(dim int, clist []string, adj [][]int, xyz SST.Coords) []SST.Loc {

	var retvar []SST.Loc

	for c := 0; c < len(adj); c++ {

		var contextgroup SST.Loc

		contextgroup.Text = clist[c]

		for cp := 0; cp < len(adj[c]); cp++ {
			if adj[c][cp] > 0 {
				contextgroup.Reln = append(contextgroup.Reln, cp)
			}
		}

		contextgroup.XYZ = SST.AssignContextSetCoordinates(xyz, c, len(adj))

		retvar = append(retvar, contextgroup)
	}
	return retvar
}

//******************************************************************

func GetContextFragments(clist []string, ooo SST.Coords) []SST.Loc {

	var retvar []SST.Loc

	for c := 0; c < len(clist); c++ {

		var contextgroup SST.Loc

		contextgroup.Text = clist[c]
		contextgroup.XYZ = SST.AssignFragmentCoordinates(ooo, c, len(clist))

		retvar = append(retvar, contextgroup)
	}
	return retvar
}

// *********************************************************************
// Misc
// *********************************************************************

func JSONStoryNodeEvent(en SST.NodeEvent) string {

	var jstr string

	//	j,_ := json.Marshal(en)

	//	jstr = string(j)

	if len(en.Text) == 0 {
		return ""
	}

	t, _ := json.Marshal(en.Text)
	text := SST.EscapeString(string(t))
	text = SST.SQLEscape(text)

	jstr += fmt.Sprintf("{\"Text\": \"%s\",\n", text)
	jstr += fmt.Sprintf("\"L\": \"%d\",\n", en.L)

	c, _ := json.Marshal(en.Chap)
	chap := SST.EscapeString(string(c))
	chap = SST.SQLEscape(chap)

	jstr += fmt.Sprintf("\"Chap\": \"%s\",\n", chap)

	jstr += fmt.Sprintf("\"Context\": \"%s\",\n", SST.EscapeString(en.Context))
	jstr += fmt.Sprintf("\"NPtr\": { \"Class\": \"%d\", \"CPtr\" : \"%d\"},\n", en.NPtr.Class, en.NPtr.CPtr)
	jxyz, _ := json.Marshal(en.XYZ)
	jstr += fmt.Sprintf("\"XYZ\": %s,\n", jxyz)

	var arrays string

	for sti := 0; sti < SST.ST_TOP; sti++ {
		var arr string
		if en.Orbits[sti] != nil {
			js, _ := json.Marshal(en.Orbits[sti])
			arr = fmt.Sprintf("%s,", string(js))
		} else {
			arr = "[],"
		}
		arrays += arr
	}

	arrays = strings.Trim(arrays, ",")

	jstr += fmt.Sprintf("\"Orbits\": [%s] }", arrays)
	return jstr
}

// *********************************************************************

func GenHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Add("Vary", "Origin")
}

// *********************************************************************

func CleanText(c string) string {

	c = strings.Replace(c, "{", "", -1)
	c = strings.Replace(c, "}", "", -1)
	c = strings.Replace(c, ",", " ", -1)
	c = strings.Replace(c, "\"", "\\\"", -1)
	return c
}

// **********************************************************

func ShowNode(sst SST.PoSST, nptr []SST.NodePtr) string {

	var ret string

	for n := 0; n < len(nptr); n++ {
		node := SST.GetDBNodeByNodePtr(sst, nptr[n])
		ret += fmt.Sprintf("%.30s", node.S)
		if n < len(nptr)-1 {
			ret += ","
		}
	}

	return ret
}

// **********************************************************

func PackageResponse(sst SST.PoSST, search SST.SearchParameters, kind string, jstr string) []byte {

	ambien, key, now := SST.GetTimeContext()
	now_ctx := SST.UpdateSTMContext(PSST, ambien, key, now, search)

	intent, _ := json.Marshal(now_ctx)
	ambient, _ := json.Marshal(ambien)

	response := fmt.Sprintf("{ \"Response\" : \"%s\",\n \"Content\" : %s,\n \"Time\" : \"%s\", \"Intent\" : %s, \"Ambient\" : %s }", kind, jstr, key, intent, ambient)

	return []byte(response)
}

//******************************************************************

func SL(list []string) string {

	var s string

	s += fmt.Sprint(" [")
	for i := 0; i < len(list); i++ {
		s += fmt.Sprint(list[i], ", ")
	}

	s += fmt.Sprint(" ]")

	return s
}

//******************************************************************

// StatusResponse defines the structure for our JSON response.

type StatusResponse struct {
	ServerStatus    string    `json:"server_status"`
	DatabaseStatus  string    `json:"database_status"`
	AvailableTopics []string  `json:"available_topics"`
	Timestamp       time.Time `json:"timestamp"`
}

//******************************************************************

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	toc := SST.GetChaptersByChapContext(PSST, "", nil, 1000) // "" for chapter and nil for context should get all

	var topics []string
	for chapter := range toc {
		topics = append(topics, chapter)
	}
	sort.Strings(topics)

	// Create the response object.
	status := StatusResponse{
		ServerStatus:    "OK",
		DatabaseStatus:  "OK",
		AvailableTopics: topics,
		Timestamp:       time.Now(),
	}

	// Marshal the struct into JSON.
	responseJSON, err := json.Marshal(status)
	if err != nil {
		http.Error(w, "Failed to generate status response", http.StatusInternalServerError)
		return
	}

	// Set the content type and send the response.
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
