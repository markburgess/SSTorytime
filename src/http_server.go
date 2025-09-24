//
// Simple web server lookup
//

package main

import (
	"fmt"
	"net/http"
	"strings"
	"os"
	"os/signal"
	"syscall"
	"sort"
	"context"
	"encoding/json"

        SST "SSTorytime"
)

// *********************************************************************

var CTX SST.PoSST  // just one persistent connection

// *********************************************************************

func main() {
	
	CTX = SST.Open(true)	

	server := &http.Server{	Addr: ":8080", }

	// The server structure in Go is weirdly opinionated ..

	http.HandleFunc("/",PageHandler)
	http.HandleFunc("/searchN4L",SearchN4LHandler)

	go func() {
		fmt.Println("Listening at http://localhost:8080")
		err := server.ListenAndServe()
		fmt.Println("Stop serving connections",err)
	}()

	SignalHandler()

	// We have to have this ugly context business...
	halt, shutdownRelease := context.WithTimeout(context.Background(),10)
	defer shutdownRelease()
	
	server.Shutdown(halt)

	fmt.Println("http_server shutdown complete.")
}

// *********************************************************************
// Handlers
// *********************************************************************

func SignalHandler() {

	signal_chan := make(chan os.Signal,1)

	signal.Notify(signal_chan, 
		syscall.SIGHUP,  // 1
		syscall.SIGINT,  // 2 ctrl-c
		syscall.SIGQUIT, // 3
		syscall.SIGTERM) // 15, CTRL-c 

	sig := <-signal_chan  // block until signal
	
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

func PageHandler(w http.ResponseWriter, r *http.Request) {

	GenHeader(w,r)

	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "text/html")
		page,err := os.ReadFile("./page.html")

		localAddr := r.Context().Value(http.LocalAddrContextKey) 
		ipaddr := fmt.Sprintf("%s",localAddr)
		page = []byte(strings.Replace(string(page),"localhost:8080",ipaddr,-1))

		if err != nil {
			fmt.Println("Can't find ./page.html")
			os.Exit(-1)
		}
		w.Write(page)

	default:
		http.Error(w, "Not supported", http.StatusMethodNotAllowed)
	}
}

// *********************************************************************
// SEARCH
// *********************************************************************

func SearchN4LHandler(w http.ResponseWriter, r *http.Request) {
	
	GenHeader(w,r)

	switch r.Method {

	case "POST","GET":
		name := r.FormValue("name")
		nclass := r.FormValue("nclass")
		ncptr := r.FormValue("ncptr")
		chapcontext := r.FormValue("chapcontext")
		
		if name == "\\lastnptr" {
			if chapcontext != "" && chapcontext != "any" {
				UpdateLastSawSection(w,r,chapcontext)
			}
			UpdateLastSawNPtr(w,r,nclass,ncptr,chapcontext)
			return
		}

		if name == "" && len(nclass) > 0 && len(ncptr) > 0 {
			// direct click on an item
			var a,b int
			fmt.Sscanf(nclass,"%d",&a)
			fmt.Sscanf(ncptr,"%d",&b)
			nstr := fmt.Sprintf("(%d,%d)",a,b)
			name = name + nstr
		}
		
		fmt.Println("\nReceived command:",name)
		
		ambient,key,_ := SST.GetTimeContext()
		
		if len(name) == 0 || name == "\\remind" {
			name = "any \\chapter reminders \\context " + key + " " + ambient
		}
		
		search := SST.DecodeSearchField(name)

		HandleSearch(search,name,w,r)
		
	default:
		http.Error(w, "Not supported", http.StatusMethodNotAllowed)
	}
}

// *********************************************************************

func UpdateLastSawSection(w http.ResponseWriter, r *http.Request,query string) {

	// update lastseen db

	fmt.Println("UPDATING STATS FOR section",query)

	SST.UpdateLastSawSection(CTX,query)
}

// *********************************************************************

func UpdateLastSawNPtr(w http.ResponseWriter, r *http.Request,class,cptr string,classifier string) {

	// update lastseen db

	var nptr SST.NodePtr
	var nclass int
	var ncptr int
	fmt.Sscanf(class,"%d",&nclass)
	fmt.Sscanf(cptr,"%d",&ncptr)
	nptr.Class = nclass
	nptr.CPtr = SST.ClassedNodePtr(ncptr)

	SST.UpdateLastSawNPtr(CTX,nclass,ncptr,classifier)

	fmt.Println("UPDATING STATS FOR",nclass,ncptr,"WITHIN",classifier)

	SST.UpdateLastSawSection(CTX,classifier)

	response := fmt.Sprintf("{ \"Response\" : \"LastSaw\",\n \"Content\" : \"ack(%s,%s)\" }",class,cptr)
	w.Write([]byte(response))

}

// *********************************************************************

func HandleSearch(search SST.SearchParameters,line string,w http.ResponseWriter, r *http.Request) {

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

	arrowptrs,sttype := SST.ArrowPtrFromArrowsNames(CTX,search.Arrows)

	arrows := arrowptrs != nil
	sttypes := sttype != nil
	limit := 0

	if search.Range > 0 {
		limit = search.Range
	} else {
		if from || to {
			limit = 30 // many paths make hard work
		} else {
			limit = 10
		}
	}

	fmt.Println("Your starting expression generated this set: ",line,"\n")
	fmt.Println(" - start set:",SL(search.Name))
	fmt.Println(" -      from:",SL(search.From))
	fmt.Println(" -        to:",SL(search.To))
	fmt.Println(" -   chapter:",search.Chapter)
	fmt.Println(" -   context:",SL(search.Context))
	fmt.Println(" -    arrows:",SL(search.Arrows))
	fmt.Println(" -    pagenr:",search.PageNr)
	fmt.Println(" - sequence/story:",search.Sequence)
	fmt.Println(" - limit/range/depth:",limit)
	fmt.Println(" - show stats:",search.Stats)
	fmt.Println()

	var nodeptrs,leftptrs,rightptrs []SST.NodePtr

	if !pagenr && !sequence {
		leftptrs = SST.SolveNodePtrs(CTX,search.From,search,arrowptrs,limit)
		rightptrs = SST.SolveNodePtrs(CTX,search.To,search,arrowptrs,limit)
	}

	nodeptrs = SST.SolveNodePtrs(CTX,search.Name,search,arrowptrs,limit)


	fmt.Println("Solved search nodes ...")

	// SEARCH SELECTION *********************************************

	// Table of contents

	if (search.Stats) {
		ShowStats(w,r,CTX,search,nodeptrs)
		return
	}

	if (context || chapter) && !name && !sequence && !pagenr && !(from || to) {
		ShowChapterContexts(w,r,CTX,search,limit)
		return
	}

	if name && ! sequence && !pagenr {
		HandleOrbit(w,r,CTX,search,nodeptrs,limit)
		return
	}

	if (name && from) || (name && to) {
		fmt.Printf("\nSearch \"%s\" has conflicting parts <to|from> and match strings\n",line)
		os.Exit(-1)
	}

	// Closed path solving, two sets of nodeptrs
	// if we have BOTH from/to (maybe with chapter/context) then we are looking for paths

	if from && to {
		HandlePathSolve(w,r,CTX,leftptrs,rightptrs,search,arrowptrs,sttype,limit)
		return
	}

	// Open causal cones, from one of these three

	if (name || from || to) && !pagenr && !sequence {

		if nodeptrs != nil {
			HandleCausalCones(w,r,CTX,nodeptrs,search,arrowptrs,sttype,limit)
			return
		}
		if leftptrs != nil {
			HandleCausalCones(w,r,CTX,leftptrs,search,arrowptrs,sttype,limit)
			return
		}
		if rightptrs != nil {
			HandleCausalCones(w,r,CTX,rightptrs,search,arrowptrs,sttype,limit)
			return
		}
	}

	// if we have page number then we are looking for notes by pagemap

	if (name || chapter) && pagenr {

		var notes []SST.PageMap

		if chapter {
			notes = SST.GetDBPageMap(CTX,search.Chapter,search.Context,search.PageNr)
			HandlePageMap(w,r,CTX,search,notes)
			return
		} else {
			for n := range search.Name {
				notes = SST.GetDBPageMap(CTX,search.Name[n],search.Context,search.PageNr)
				HandlePageMap(w,r,CTX,search,notes)
			}
			return
		}
	}

	// Look for axial trails following a particular arrow, like _sequence_ 

	if sequence {
		HandleStories(w,r,CTX,search,nodeptrs,arrowptrs,sttype,limit)
		return
	}

	// if we have sequence with arrows, then we are looking for sequence context or stories

	if arrows || sttypes {
		HandleMatchingArrows(w,r,CTX,search,arrowptrs,sttype)
		return
	}

	fmt.Println("Didn't find a solver")
}

// *********************************************************************

func HandleOrbit(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,search SST.SearchParameters,nptrs []SST.NodePtr,limit int) {

	var count int
	var array string

	origin := SST.Coords{X : 0.0, Y : 0.0, Z : 0.0}

	for n := 0; n < len(nptrs); n++ {

		count++

		if count > limit {
			break
		}

		fmt.Printf("Assembling Node Orbit(%v)\n",nptrs[n])

		orb := SST.GetNodeOrbit(CTX,nptrs[n],"",limit)
		// create a set of coords for len(nptrs) disconnected nodes

		xyz := SST.RelativeOrbit(origin,SST.R0,n,len(nptrs))
		orb = SST.SetOrbitCoords(xyz,orb)

		array += SST.JSONNodeEvent(CTX,nptrs[n],xyz,orb)
		array += ","
	}

	array = strings.Trim(array,",")
	content := fmt.Sprintf("[ %s ]",array)
	response := PackageResponse(ctx,search,"Orbits",content)
	
	//fmt.Println("REPLY:\n",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Reply Orbit sent")
}

// *********************************************************************

func HandleCausalCones(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,nptrs []SST.NodePtr,search SST.SearchParameters,arrows []SST.ArrowPtr, sttype []int,limit int) {

	chap := search.Chapter
	context := search.Context

	fmt.Println("HandleCausalCones()",nptrs)
	var total int = 1
	var data string

	if len(sttype) == 0 {
		sttype = []int{0,1,2,3}
	}

	for n := range nptrs {
		for st := range sttype {

			jstr,count := PackageConeFromOrigin(ctx,nptrs[n],n,sttype[st],chap,context,len(nptrs),limit)

			if count > 0 {
				total += count
				data += jstr
				data += ","
			}

			if total > limit {
				break
			}
		}

		if total > limit {
			break
		}
	}

	data = strings.Trim(data,",")
	array := fmt.Sprintf("[%s]",data)

	response := PackageResponse(ctx,search,"ConePaths",array)
	//fmt.Println("CasualConePath reponse",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent cone")
}

//******************************************************************

func PackageConeFromOrigin(ctx SST.PoSST,nptr SST.NodePtr,nth int,sttype int,chap string,context []string,dimnptr,limit int) (string,int) {

	// Package a JSON object for the nth/dimnptr causal cone , assigning each nth the same width

	var wpaths [][]SST.WebPath

	fcone,count := SST.GetFwdPathsAsLinks(CTX,nptr,sttype,limit,limit)
	wpaths = append(wpaths,SST.LinkWebPaths(CTX,fcone,nth,chap,context,dimnptr,limit)...)

	if sttype != 0 {
		bcone,countb := SST.GetFwdPathsAsLinks(CTX,nptr,-sttype,limit,limit)
		wpaths = append(wpaths,SST.LinkWebPaths(CTX,bcone,nth,chap,context,dimnptr,limit)...)
		count += countb
	}

	wstr,err := json.Marshal(wpaths)

	if wpaths == nil {
		return "",0
	}

	if err != nil {
		fmt.Println("Error in PackageConeFromOrigin",err)
		os.Exit(-1)
	}

	title := SST.EscapeString(SST.GetDBNodeByNodePtr(ctx,nptr).S)

	jstr := fmt.Sprintf(" { \"NClass\" : %d,\n",nptr.Class)
	jstr += fmt.Sprintf("   \"NCPtr\" : %d,\n",nptr.CPtr)
	jstr += fmt.Sprintf("   \"Title\" : \"%s\",\n",title)  // tbd
	jstr += fmt.Sprintf("   \"Paths\" : %s\n}",string(wstr))	

	return jstr,count
}

//******************************************************************

func HandlePathSolve(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,leftptrs,rightptrs []SST.NodePtr,search SST.SearchParameters,arrowptrs []SST.ArrowPtr,sttype []int,maxdepth int) {

	chapter := search.Chapter
	context := search.Context

	fmt.Println("HandlePathSolve(",leftptrs,",",rightptrs,")")

	var Lnum,Rnum int
	var left_paths, right_paths [][]SST.Link

	// Find the path matrix

	var solutions [][]SST.Link
	var ldepth,rdepth int = 2,2

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		left_paths,Lnum = SST.GetEntireNCSuperConePathsAsLinks(CTX,"fwd",leftptrs,ldepth,chapter,context,maxdepth)
		right_paths,Rnum = SST.GetEntireNCSuperConePathsAsLinks(CTX,"bwd",rightptrs,rdepth,chapter,context,maxdepth)

		if Lnum == 0 || Rnum == 0 {
			fmt.Println("Nothing, trying reverse")
			left_paths,Lnum = SST.GetEntireNCSuperConePathsAsLinks(CTX,"bwd",leftptrs,ldepth,chapter,context,maxdepth)
			right_paths,Rnum = SST.GetEntireNCSuperConePathsAsLinks(CTX,"fwd",rightptrs,rdepth,chapter,context,maxdepth)

			if Lnum == 0 || Rnum == 0 {
				fmt.Println("No paths")
				response := PackageResponse(ctx,search,"PathSolve","")	
				w.Header().Set("Content-Type", "application/json")
				w.Write(response)
				return
			}
		}

		solutions,_ = SST.WaveFrontsOverlap(CTX,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		fmt.Println("solutions",solutions)

		if len(solutions) > 0 {
			// format paths
			var jstr string

			jstr += fmt.Sprintf(" { \"NClass\" : %d,\n",solutions[0][0].Dst.Class)
			jstr += fmt.Sprintf("   \"NCPtr\" : %d,\n",solutions[0][0].Dst.CPtr)
			jstr += fmt.Sprintf("   \"Title\" : \"%s\",\n","path solutions")
			jstr += fmt.Sprintf("   \"BTWC\" : [ %s ],\n",SST.BetweenNessCentrality(CTX,solutions))
			jstr += fmt.Sprintf("   \"Supernodes\" : [ %s ],\n",SST.SuperNodes(CTX,solutions,maxdepth))

			var wpaths [][]SST.WebPath
			nth := 0
			swimlanes := 1

			wpaths = append(wpaths,SST.LinkWebPaths(CTX,solutions,nth,chapter,context,swimlanes,maxdepth)...)

			if wpaths == nil {
				break
			}

			wstr,_ := json.Marshal(wpaths)
			jstr += fmt.Sprintf("   \"Paths\" : %s }",string(wstr))

			array_pack := fmt.Sprintf("[%s]",jstr)
			response := PackageResponse(ctx,search,"PathSolve",array_pack)

			//fmt.Println("PATH SOLVE:",string(response))

			w.Header().Set("Content-Type", "application/json")
			w.Write(response)
			return
		}

		if turn % 2 == 0 {
			ldepth++
		} else {
			rdepth++
		}
	}
	
	fmt.Println("No paths satisfy constraints")
	response := PackageResponse(ctx,search,"PathSolve","[]")

	//fmt.Println("PATHSOLVE NOTES",string(response))
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent path solve")
}

//******************************************************************

func HandlePageMap(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,search SST.SearchParameters,notes []SST.PageMap) {

	fmt.Println("Solver/handler: HandlePageMap()")

	jstr := SST.JSONPage(CTX,notes)
	response := PackageResponse(ctx,search,"PageMap",jstr)

	if notes != nil {
		UpdateLastSawSection(w,r,notes[0].Chapter)
	}

	//fmt.Println("PAGEMAP NOTES",string(response))
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent pagemap")
}

//******************************************************************

func HandleStories(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,search SST.SearchParameters,nodeptrs []SST.NodePtr,arrowptrs []SST.ArrowPtr,sttypes []int,limit int) {

	if arrowptrs == nil {
		arrowptrs,sttypes = SST.ArrowPtrFromArrowsNames(CTX,[]string{"!then!"})
	}

	fmt.Println("Solver/handler: HandleStories()")

	stories := SST.GetSequenceContainers(ctx,nodeptrs,arrowptrs,sttypes,limit)

	fmt.Println("STORIES",stories,"from node set",len(nodeptrs),arrowptrs)
	jarray := ""

	for s := range stories {
		var jstory string
		
		for a := 0; a < len(stories[s].Axis); a++ {
			jstr := JSONStoryNodeEvent(stories[s].Axis[a])
			jstory += fmt.Sprintf("%s,",jstr)
		}
		jstory = strings.Trim(jstory,",")
		jarray += fmt.Sprintf("[%s],",jstory)
	}

	if jarray == "" {
		jarray = "[]"
	}

	data := strings.Trim(jarray,",")
	response := PackageResponse(ctx,search,"Sequence",data)

	fmt.Println("Sequence...",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent sequence")

}

// *********************************************************************

func HandleMatchingArrows(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,search SST.SearchParameters,arrowptrs []SST.ArrowPtr,sttype []int) {

	fmt.Println("Solver/handler: HandleMatchingArrows()")

	type ArrowList struct {
		ArrPtr SST.ArrowPtr
		ASTtype int
		Short string
		Long string
		InvPtr SST.ArrowPtr
		ISTtype int
		InvS string
		InvL string
	}

	var arrows []ArrowList

	for a := range arrowptrs {
		adir := SST.GetDBArrowByPtr(ctx,arrowptrs[a])
		inv := SST.GetDBArrowByPtr(ctx,SST.INVERSE_ARROWS[arrowptrs[a]])

		var al ArrowList		
		al.ArrPtr = arrowptrs[a]
		al.ASTtype = SST.STIndexToSTType(adir.STAindex)
		al.Short = adir.Short
		al.Long = adir.Long
		al.InvPtr = inv.Ptr
		al. ISTtype = SST.STIndexToSTType(inv.STAindex)
		al.InvS = inv.Short
		al.InvL = inv.Long
		arrows = append(arrows,al)
	}

	if arrowptrs == nil {
		for st := range sttype {
			adirs := SST.GetDBArrowBySTType(ctx,sttype[st])
			for adir := range adirs {
				inv := SST.GetDBArrowByPtr(ctx,SST.INVERSE_ARROWS[adirs[adir].Ptr])
				
				var al ArrowList
				al.ArrPtr = adirs[adir].Ptr
				al.ASTtype = SST.STIndexToSTType(adirs[adir].STAindex)
				al.Short = adirs[adir].Short
				al.Long = adirs[adir].Long
				al.InvPtr = inv.Ptr
				al.ISTtype = SST.STIndexToSTType(inv.STAindex)
				al.InvS = inv.Short
				al.InvL = inv.Long
				arrows = append(arrows,al)
			}
		}
	}

	data,_ := json.Marshal(arrows)
	response := PackageResponse(ctx,search,"Arrows",string(data))

	fmt.Println("Arrows...",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent arrows")
}

// *********************************************************************

func ShowStats(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,search SST.SearchParameters,nptrs []SST.NodePtr) {

	var retval []SST.LastSeen

	if nptrs == nil {
		retval = SST.GetLastSawSection(ctx)
	} else {

		for  n := range nptrs {
			nptr := SST.GetLastSawNPtr(ctx,nptrs[n])
			retval = append(retval,nptr)
		}
	}

	data,_ := json.Marshal(retval)

	response := PackageResponse(ctx,search,"STAT",string(data))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent stat")

}

// *********************************************************************

func ShowChapterContexts(w http.ResponseWriter, r *http.Request,ctx SST.PoSST,search SST.SearchParameters,limit int) {

	chap := search.Chapter
	context := search.Context

	fmt.Println("Solver/handler: ShowChapterContexts()")

	var chapters []SST.ChCtx
	var chap_list []string

	toc := SST.GetChaptersByChapContext(ctx,chap,context,limit)

	for chaps := range toc {
		chap_list = append(chap_list,chaps)
	}

	sort.Strings(chap_list)

	for c := 0; c < len(chap_list); c++ {

		var chap_anchor SST.ChCtx
		
		chap_anchor.Chapter = chap_list[c]
		chap_anchor.XYZ = SST.AssignChapterCoordinates(c,len(chap_list))

		// Fractionate the (chapter,context) information

		dim,clist,adj := SST.IntersectContextParts(toc[chap_list[c]])
		spectrum := SST.GetContextTokenFrequencies(toc[chap_list[c]])
		intent,ambient := SST.ContextIntentAnalysis(spectrum,toc[chap_list[c]])		

		chap_anchor.Context = GetContextSets(dim,clist,adj,chap_anchor.XYZ)
		chap_anchor.Single = GetContextFragments(intent,chap_anchor.XYZ)
		chap_anchor.Common = GetContextFragments(ambient,chap_anchor.XYZ)

		chapters = append(chapters,chap_anchor)
	}

	data,_ := json.Marshal(chapters)
	response := PackageResponse(ctx,search,"TOC",string(data))

	//fmt.Println("Chap/context...",string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent content")
}

//******************************************************************

func GetContextSets(dim int,clist []string,adj [][]int, xyz SST.Coords) []SST.Loc {

	var retvar []SST.Loc

	for c := 0; c < len(adj); c++ {

		var contextgroup SST.Loc

		contextgroup.Text = clist[c]

		for cp := 0; cp < len(adj[c]); cp++ {
			if adj[c][cp] > 0 {
				contextgroup.Reln = append(contextgroup.Reln,cp)
			}
		}

		contextgroup.XYZ = SST.AssignContextSetCoordinates(xyz,c,len(adj))

		retvar = append(retvar,contextgroup)
	}
	return retvar
}

//******************************************************************

func GetContextFragments(clist []string, ooo SST.Coords) []SST.Loc {

	var retvar []SST.Loc

	for c := 0; c < len(clist); c++ {

		var contextgroup SST.Loc

		contextgroup.Text = clist[c]
		contextgroup.XYZ = SST.AssignFragmentCoordinates(ooo,c,len(clist))

		retvar = append(retvar,contextgroup)
	}
	return retvar
}

// *********************************************************************
// Misc
// *********************************************************************

func JSONStoryNodeEvent(en SST.NodeEvent) string {

	var jstr string

	if len(en.Text) == 0 {
		return ""
	}

	t,_ := json.Marshal(en.Text)
	text := SST.EscapeString(string(t))
	jstr += fmt.Sprintf("{\"Text\": \"%s\",\n",text)
	jstr += fmt.Sprintf("\"L\": \"%d\",\n",en.L)
	c,_ := json.Marshal(en.Chap)
	chap := SST.EscapeString(string(c))
	jstr += fmt.Sprintf("\"Chap\": \"%s\",\n",chap)
	jstr += fmt.Sprintf("\"Context\": \"%s\",\n",SST.EscapeString(en.Context))
	jstr += fmt.Sprintf("\"NPtr\": { \"Class\": \"%d\", \"CPtr\" : \"%d\"},\n",en.NPtr.Class,en.NPtr.CPtr)
	jxyz,_ := json.Marshal(en.XYZ)
	jstr += fmt.Sprintf("\"XYZ\": %s,\n",jxyz)

	var arrays string

	for sti := 0; sti < SST.ST_TOP; sti++ {
		var arr string
		if en.Orbits[sti] != nil {
			js,_ := json.Marshal(en.Orbits[sti])
			arr = fmt.Sprintf("%s,",string(js))
		} else {
			arr = "[],"
		}
		arrays += arr
	}
	arrays = strings.Trim(arrays,",")
	jstr += fmt.Sprintf("\"Orbits\": [%s] }",arrays)
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

	c = strings.Replace(c,"{","",-1)
	c = strings.Replace(c,"}","",-1)
	c = strings.Replace(c,","," ",-1)
	c = strings.Replace(c,"\"","\\\"",-1)
	return c
}

// **********************************************************

func ShowNode(ctx SST.PoSST,nptr []SST.NodePtr) string {

	var ret string

	for n := 0; n < len(nptr); n++ {
		node := SST.GetDBNodeByNodePtr(ctx,nptr[n])
		ret += fmt.Sprintf("%.30s",node.S)
		if n < len(nptr)-1 {
			ret += ","
		}
	}

	return ret
}

// **********************************************************

func PackageResponse(ctx SST.PoSST,search SST.SearchParameters,kind string, jstr string) []byte {

	ambien,key,now := SST.GetTimeContext()
	now_ctx := SST.UpdateSTMContext(CTX,ambien,key,now,search)

	intent,_ := json.Marshal(now_ctx)
	ambient,_ := json.Marshal(ambien)

	response := fmt.Sprintf("{ \"Response\" : \"%s\",\n \"Content\" : %s,\n \"Time\" : \"%s\", \"Intent\" : %s, \"Ambient\" : %s }",kind,jstr,key,intent,ambient)

	return []byte(response)
}

//******************************************************************

func SL(list []string) string {

	var s string

	s += fmt.Sprint(" [")
	for i := 0; i < len(list); i++ {
		s += fmt.Sprint(list[i],", ")
	}

	s += fmt.Sprint(" ]")

	return s
}














