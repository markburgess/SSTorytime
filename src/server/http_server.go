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
	"io"
	"sort"
	"strings"
	"syscall"
	"time"
	"flag"
	"errors"	
	"crypto/md5"

	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

// *********************************************************************
//  Go Embedded filesystem for HTML/CSS resources
// *********************************************************************

//This is an ugly Go directive to embed text files into the binary
//go:embed all:public

var content embed.FS
var VERBOSE bool

// *********************************************************************
// Main
// *********************************************************************

func main() {

	rootpath := Init()
	Start(rootpath)
}

//**************************************************************

func Init() string {

	flag.Usage = Usage

	verbosePtr := flag.Bool("v", false,"verbose")
	resourcePtr := flag.String("resources", "/mnt", "Root directory for serving /Resources/ files")

	flag.Parse()

	if *verbosePtr {
		VERBOSE = true
	}

	return *resourcePtr
}

//**************************************************************

func Usage() {

        // We assume that the server is run from the directory under which
	// it will store all cached files. The resources directory is extra read-only

	fmt.Printf("usage: http_server [-resources string]\n")
	flag.PrintDefaults()
	os.Exit(0)
}

// *********************************************************************

func Start(resources string) {

	// 1. Create shared filesystem view rooted inside the "public" directory.

	publicFS, err := fs.Sub(content, "public")

	if err != nil {
		log.Fatal("failed to create sub-filesystem:", err)
	}


	// 2. Create a router (ServeMux) and register various handlers.

	fileserver1 := http.FileServer(http.FS(publicFS))
	fileserver2 := http.FileServer(http.Dir(resources))
	fileserver3 := http.FileServer(http.Dir("./cacheroot"))

	mux := http.NewServeMux()

	mux.Handle("/", fileserver1)
	mux.Handle("/Resources/", http.StripPrefix("/Resources/", fileserver2))
	mux.Handle("/Assets/", http.StripPrefix("/Assets/cacheroot", fileserver3))
	mux.HandleFunc("/searchN4L", SearchN4LHandler)
	mux.HandleFunc("/Upload", UploadHandler)
	mux.HandleFunc("/SearchAssets", AssetsHandler)

	fmt.Println("\n***********************************************\n")
	fmt.Println(" *  File serving resources, set to: ",resources)
	fmt.Println("\n *  Use -resources=/a/b/c to configure")
	fmt.Println("\n***********************************************\n")


	// 3. Create server instances for graceful shutdown.

	http_srv := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shost := strings.Split(r.Host,":")[0]
		fmt.Println("Redirecting http","https://"+shost+":8443"+r.URL.String())
			http.Redirect(w, r, "https://"+shost+":8443"+r.URL.String(), http.StatusMovedPermanently)
		}),
	}

	https_srv := &http.Server{Addr: ":8443", Handler: mux}

	// Graceful Shutdown Channel

	done := make(chan struct{})
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start servers

	go func() {
		<-quit
		log.Println("Shutting down servers...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		http_srv.Shutdown(ctx)
		https_srv.Shutdown(ctx)
		close(done)
	}()

	go func() {
		if err := http_srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP Listen: %v", err)
		}
	}()

	go func() {
		if err := https_srv.ListenAndServeTLS("../server/cert.pem", "../server/key.pem"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTPS Listen: %v", err)
		}
	}()

	log.Println("Servers running on :8080 and :8443")
	<-done

	log.Println("Servers stopped gracefully")
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

func SearchN4LHandler(w http.ResponseWriter, r *http.Request) {

	sst := SST.Open(true)

	switch r.Method {

	case "POST", "GET":
		name := r.FormValue("name")
		nclass := r.FormValue("nclass")
		ncptr := r.FormValue("ncptr")
		chapcontext := r.FormValue("chapcontext")

		if name == "\\lastnptr" {
			if chapcontext != "" && chapcontext != "any" {
				UpdateLastSawSection(sst,w,r,chapcontext)
			}
			UpdateLastSawNPtr(sst,w,r,nclass,ncptr,chapcontext)
			return
		}

		name = SST.CheckNPtrQuery(name,nclass,ncptr)
		name = SST.CheckRemindQuery(name)
		name = SST.CheckHelpQuery(name)
		name = SST.CheckConceptQuery(name)

		fmt.Println("\nReceived command:", name)

		search := SST.DecodeSearchField(name)

		HandleSearch(sst,search, name, w, r)

	default:
		http.Error(w, "Not supported", http.StatusMethodNotAllowed)
	}
	
	SST.Close(sst)
}

// *********************************************************************

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(32 << 20)
	
	uri := r.FormValue("uri")
	
	if uri != "none" {
		UploadURI(w,r)
	} else {
		UploadInline(w,r)
	}
	
}

// *********************************************************************

func UploadURI(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	chapter := r.FormValue("chapter")
	context := r.FormValue("context")
	ext := r.FormValue("extension")
	
	uri := r.FormValue("uri")
	dir := FileCacheLocation(name,ext,chapter,context)
	
	file,err := SST.GetURIFile(uri)
	
	if file == "" || err != nil {
		fmt.Println("1. Upload failed making",dir,"\n", err,"\n")
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))
		return
	}

        err = os.MkdirAll(dir, 0755)

	if err != nil {
		fmt.Println("2. Upload failed making",dir,"\n", err,"\n")
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))
		return	
	}
	
	// cache node = 3 words + date pattern

	c1,c2,_ := SST.GetTimeContext()

	target := fmt.Sprintf("%s%s",SST.SanitizePath(c1),SST.SanitizePath(c2))
	target = strings.ReplaceAll(target,"__","_")
        location := dir + target + "." + ext
	
	dst, err := os.Create(location)

	if err != nil {
		fmt.Println("3. Upload failed making",dir,"\n", err,"\n")
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))
		return
	}
	
	defer dst.Close()

	fmt.Println("Uploaded cache file",location)	

	// Write the file
	bytes := []byte(file)
	size, err := dst.Write(bytes)
	
	if err != nil {
		fmt.Println("Upload failed writing",location, err)
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))
		return
	}

	response := fmt.Sprintf("{ \"Response\" : \"Uploaded\",\n \"Content\" : \"wrote %d bytes to %s\" }",size,location)
	w.Write([]byte(response))

}

// *********************************************************************

func UploadInline(w http.ResponseWriter, r *http.Request) {

	fmt.Println("HANDLE INLINE UPLOAD\n")
	var err error		

	name := r.FormValue("name")
	chapter := r.FormValue("chapter")
	context := r.FormValue("context")
	ext := r.FormValue("extension")

	file, header, err := r.FormFile("filedata")

	fmt.Sprintf("File inline upload: %s --> %s\n",header,err)
	
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}

	defer file.Close()

	dir := FileCacheLocation(name,ext,chapter,context)

        err = os.MkdirAll(dir, 0755)

	if err != nil {
		fmt.Println("2. Upload failed making",dir,"\n", err,"\n")
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))
		return	
	}
	
	// cache node = 3 words + date pattern

	c1,c2,_ := SST.GetTimeContext()

	target := fmt.Sprintf("%s%s",SST.SanitizePath(c1),SST.SanitizePath(c2))
	target = strings.ReplaceAll(target,"__","_")
        location := dir + target + "." + ext
	
	dst, err := os.Create(location)

	if err != nil {
		fmt.Println("3. Upload failed making",dir,"\n", err,"\n")
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))
		return
	}
	
	defer dst.Close()

	fmt.Println("Uploaded cache file",location)	

	_, err = io.Copy(dst,file)
		
	if err != nil {
		fmt.Println("Upload failed writing",location, err)
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))
		return
	}

	if CheckFile(location) {
		response := fmt.Sprintf("{ \"Response\" : \"Uploaded\",\n \"Content\" : \"wrote %s\" }",location)
		w.Write([]byte(response))
	} else {
		os.Remove(location)
		fmt.Println("Upload filetype rejected",location, err)
		w.WriteHeader(http.StatusInternalServerError)
		response := fmt.Sprintf("{ \"Response\" : \"Failed\",\n \"Content\" : \"Error: %s\" }",err)
		w.Write([]byte(response))		
	}
	
}

// *********************************************************************

func AssetsHandler(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	chapter := r.FormValue("chapter")
	context := r.FormValue("context")
	ext := "any"

	dir := FileCacheLocation(name,ext,chapter,context)
	
	w.Header().Set("Content-Type", "application/json")

	response := ListCacheAssets(dir)

	w.Write([]byte(response))
}

// *********************************************************************

func UpdateLastSawSection(sst SST.PoSST,w http.ResponseWriter, r *http.Request, query string) {

	// update lastseen db

	fmt.Println("UPDATING STATS FOR section", query)

	SST.UpdateLastSawSection(sst, query)
}

// *********************************************************************

func UpdateLastSawNPtr(sst SST.PoSST,w http.ResponseWriter, r *http.Request, class, cptr string, classifier string) {

	// update lastseen db

	var nptr SST.NodePtr
	var nclass int
	var ncptr int
	fmt.Sscanf(class, "%d", &nclass)
	fmt.Sscanf(cptr, "%d", &ncptr)
	nptr.Class = nclass
	nptr.CPtr = SST.ClassedNodePtr(ncptr)

	SST.UpdateLastSawNPtr(sst, nclass, ncptr, classifier)

	fmt.Println("UPDATING STATS FOR", nclass, ncptr, "WITHIN", classifier)

	SST.UpdateLastSawSection(sst, classifier)

	response := fmt.Sprintf("{ \"Response\" : \"LastSaw\",\n \"Content\" : \"ack(%s,%s)\" }", class, cptr)
	w.Write([]byte(response))

}

// *********************************************************************

func HandleSearch(sst SST.PoSST,search SST.SearchParameters, line string, w http.ResponseWriter, r *http.Request) {

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

	arrowptrs, sttype := SST.ArrowPtrFromArrowsNames(&sst, search.Arrows)

	arrows := arrowptrs != nil
	sttypes := sttype != nil

	minlimit,maxlimit := SST.MinMaxPolicy(search)

	fmt.Println()
	fmt.Println("        start set:", SL(search.Name))
	fmt.Println("          finding:", SL(search.Finds))
	fmt.Println("             from:", SL(search.From))
	fmt.Println("               to:", SL(search.To))
	fmt.Println("          chapter:", search.Chapter)
	fmt.Println("          context:", SL(search.Context))
	fmt.Println("           arrows:", SL(search.Arrows))
	fmt.Println("           pageNR:", search.PageNr)
	fmt.Println("   sequence/story:", search.Sequence)
	fmt.Println("limit/range/depth:", maxlimit)
	fmt.Println(" at least/minimum:", minlimit)
	fmt.Println("       show stats:", search.Stats)
	fmt.Println("   not seen hours:", search.Horizon)
	fmt.Println()

	var nodeptrs, leftptrs, rightptrs []SST.NodePtr

	if (from || to) && !pagenr && !sequence {
		leftptrs = SST.SolveNodePtrs(sst, search.From, search, arrowptrs, maxlimit)
		rightptrs = SST.SolveNodePtrs(sst, search.To, search, arrowptrs, maxlimit)
	}

	if search.Sequence && len(search.Name) == 0 {
		search.Name = append(search.Name,"any")
	}

	nodeptrs = SST.SolveNodePtrs(sst, search.Name, search, arrowptrs, maxlimit)

	fmt.Println("Solved search nodes ...")

	// SEARCH SELECTION *********************************************

	// Table of contents

	if search.Stats {
		ShowStats(w,r,sst,search,nodeptrs)
		return
	}

	if (context || chapter) && !name && !sequence && !pagenr && !(from || to) {
		ShowChapterContexts(w,r,sst,search,maxlimit)
		return
	}

	if name && !sequence && !pagenr {
		HandleOrbit(w,r,sst,search,nodeptrs,maxlimit)
		return
	}

	if (name && from) || (name && to) {
		fmt.Printf("\nSearch \"%s\" has conflicting parts <to|from> and match strings\n", line)
		os.Exit(-1)
	}

	// Closed path solving, two sets of nodeptrs
	// if we have BOTH from/to (maybe with chapter/context) then we are looking for paths

	if from && to {
		HandlePathSolve(w,r,sst,leftptrs,rightptrs,search,arrowptrs,sttype,minlimit,maxlimit)
		return
	}

	// Open causal cones, from one of these three

	if (name || from || to) && !pagenr && !sequence {

		if nodeptrs != nil {
			HandleCausalCones(w,r,sst,nodeptrs,search,arrowptrs,sttype,maxlimit)
			return
		}
		if leftptrs != nil {
			HandleCausalCones(w,r,sst,leftptrs,search,arrowptrs,sttype,maxlimit)
			return
		}
		if rightptrs != nil {
			HandleCausalCones(w,r,sst,rightptrs,search,arrowptrs,sttype,maxlimit)
			return
		}
	}

	// if we have page number then we are looking for notes by pagemap

	if (name || chapter || context) && pagenr {

		var notes []SST.PageMap

		if !(name || chapter) {
			search.Chapter = "%%"
			chapter = true
		}

		if chapter {
			notes = SST.GetDBPageMap(sst,search.Chapter,search.Context,search.PageNr)
			HandlePageMap(w,r,sst,search,notes)
			return
		} else {
			for n := range search.Name {
				notes = SST.GetDBPageMap(sst,search.Name[n],search.Context,search.PageNr)
				HandlePageMap(w,r,sst,search,notes)
			}
			return
		}
	}

	// Look for axial trails following a particular arrow, like _sequence_

	if sequence {
		HandleStories(w,r,sst,search,nodeptrs,arrowptrs,sttype,maxlimit)
		return
	}

	// if we have sequence with arrows, then we are looking for sequence context or stories

	if arrows || sttypes {
		HandleMatchingArrows(w,r,sst,search,arrowptrs,sttype)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	data,_ := json.Marshal("No solver matched this search")
	response := PackageResponse(sst, search, "Error", string(data))

	w.Write(response)

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

		fmt.Printf("Assembling Node Orbit(%v)\n",nptrs[n])

		orb := SST.GetNodeOrbit(&sst,nptrs[n],"",limit)
		// create a set of coords for len(nptrs) disconnected nodes

		fmt.Printf("...Setting coordinates\n")
		xyz := SST.RelativeOrbit(origin,SST.R0,n,len(nptrs))
		orb = SST.SetOrbitCoords(xyz, orb)

		nodeevent := SST.JSONNodeEvent(sst,nptrs[n],xyz,orb)
		array = append(array, nodeevent)
	}

	data, _ := json.Marshal(array)
	response := PackageResponse(sst,search,"Orbits",string(data))

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

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent cone")
}

//******************************************************************

func PackageConeFromOrigin(sst SST.PoSST, nptr SST.NodePtr, nth int, sttype int, chap string, context []string, dimnptr, limit int) (SST.WebConePaths, int) {

	// Package a JSON object for the nth/dimnptr causal cone , assigning each nth the same width

	var wpaths [][]SST.WebPath

	fcone, count := SST.GetFwdPathsAsLinks(&sst, nptr, sttype, limit, limit)
	wpaths = append(wpaths, SST.LinkWebPaths(&sst, fcone, nth, chap, context, dimnptr, limit)...)

	if sttype != 0 {
		bcone, countb := SST.GetFwdPathsAsLinks(&sst, nptr, -sttype, limit, limit)
		wpaths = append(wpaths, SST.LinkWebPaths(&sst, bcone, nth, chap, context, dimnptr, limit)...)
		count += countb
	}

	var subcone SST.WebConePaths
	subcone.RootNode = nptr
	subcone.Title = SST.GetDBNodeByNodePtr(&sst, nptr).S
	subcone.Paths = wpaths

	return subcone, count
}

//******************************************************************

func HandlePathSolve(w http.ResponseWriter, r *http.Request, sst SST.PoSST, leftptrs, rightptrs []SST.NodePtr, search SST.SearchParameters, arrowptrs []SST.ArrowPtr, sttype []int, mindepth,maxdepth int) {

	chapter := search.Chapter
	context := search.Context

	fmt.Println("HandlePathSolve(", leftptrs, ",", rightptrs, ")")

	solutions := SST.GetPathsAndSymmetries(&sst,leftptrs,rightptrs,chapter,context,arrowptrs,sttype,mindepth,maxdepth)

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

		wpaths = append(wpaths, SST.LinkWebPaths(&sst, solutions, nth, chapter, context, swimlanes, maxdepth)...)

		soln.Paths = wpaths
		pack = append(pack, soln)
		array_pack, _ := json.Marshal(pack)

		response := PackageResponse(sst, search, "PathSolve", string(array_pack))

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

	displayset := FilterSeen(sst,notes,search)

	jstr := SST.JSONPage(sst,displayset)
	response := PackageResponse(sst, search, "PageMap", jstr)

	if notes != nil {
		UpdateLastSawSection(sst,w,r,notes[0].Chapter)
	}

	//fmt.Println("PAGEMAP NOTES",string(response))
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	fmt.Println("Done/sent pagemap")
}

//******************************************************************

func FilterSeen(sst SST.PoSST,notes []SST.PageMap,search SST.SearchParameters) []SST.PageMap {

	if search.Horizon == 0 {
		return notes
	}

	excluded_nptrs := SST.GetNewlySeenNPtrs(sst,search)

	var filtered []SST.PageMap

	for _,note := range notes {

		var newline SST.PageMap

		for _,l := range note.Path {
			if excluded_nptrs[l.Dst] {
				continue
			}
			newline.Path = append(newline.Path,l)
		}

		if newline.Path == nil {
			continue
		}
		newline.Chapter = note.Chapter
		newline.Alias  = note.Alias
		newline.Context  = note.Context
		newline.Line = note.Line

		filtered = append(filtered,newline)
	}

	return filtered
}

//******************************************************************

func HandleStories(w http.ResponseWriter, r *http.Request, sst SST.PoSST, search SST.SearchParameters, nodeptrs []SST.NodePtr, arrowptrs []SST.ArrowPtr, sttypes []int, limit int) {

	if arrowptrs == nil {
		arrowptrs, sttypes = SST.ArrowPtrFromArrowsNames(&sst, []string{"!then!"})
	}

	fmt.Println("Solver/handler: HandleStories()")

	stories := SST.GetSequenceContainers(&sst, nodeptrs, arrowptrs, sttypes, limit)

	var node_events []SST.NodeEvent

	for s := 0; s < len(stories); s++ {

		for _, ne := range stories[s].Axis {
			node_events = append(node_events,ne)
		}
	}

	jarray,_ := json.Marshal(node_events)

	response := PackageResponse(sst, search, "Sequence", string(jarray))

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
		adir := SST.GetDBArrowByPtr(&sst, arrowptrs[a])
		inv := SST.GetDBArrowByPtr(&sst, sst.INVERSE_ARROWS[arrowptrs[a]])

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
				inv := SST.GetDBArrowByPtr(&sst, sst.INVERSE_ARROWS[adirs[adir].Ptr])

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

func ShowNode(sst *SST.PoSST, nptr []SST.NodePtr) string {

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
	now_ctx := SST.UpdateSTMContext(&sst, ambien, key, now, search)

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

func FileCacheLocation(name,ext,chapter,context string) string {

        // We assume that the server is run from the directory under which
	// it will store all cached files

	chap := SST.SanitizePath(chapter)
        context = SST.SanitizePath(context)
	item := ""

        // For long texts we need to hash something, hoping the text will not change

	_,class := SST.StorageClass(name)
	
        switch class {

        case SST.N1GRAM, SST.N2GRAM, SST.N3GRAM:

		item = SST.SanitizePath(name)
        default:
		words := strings.Split(name," ")
		data := []byte(name)
		hash := md5.Sum(data)
		item = fmt.Sprintf("%s_%s_%s_%x",words[0],words[1],words[2],hash)
        }

        dir := fmt.Sprintf("./cacheroot/%s/%s/%s/",chap,context,item)

	return dir
}

//******************************************************************

func ListCacheAssets(path string) []byte {

	var response string
	
	files, err := os.ReadDir(path)

	if err != nil {
		response = fmt.Sprintf("{ \"Response\" : \"Assets\",\n \"Content\" : [] }")
		return []byte(response)
	}

	var array []string
	
	for _, file := range files {
		array = append(array,"/Assets/"+path+file.Name())
	}
	
	data, _ := json.Marshal(array)
	response = fmt.Sprintf("{ \"Response\" : \"Assets\",\n \"Content\" : %s }",data)

	return []byte(response)

}

//******************************************************************

func CheckFile(location string) bool {

	f, err := os.Open(location)

	if err == nil {

		buf := make([]byte, 64)

		_, err := io.ReadFull(f, buf)
		
		if err != nil {
			return false
		}

		contenttype := http.DetectContentType(buf)

		fmt.Println("TYPE",contenttype)
		f.Close();
		return true
	}

	return false
}
