//******************************************************************
//
// Replacement for searchN4L
// single search string without complex options
//
//******************************************************************

package main

import (
	"fmt"
	"os"
	"sort"
	"flag"
	"strings"

        SST "SSTorytime"
)

//******************************************************************

var VERBOSE bool = false

var TESTS = []string{ 
	"range rover out of its depth",
	"\"range rover\" \"out of its depth\"",
	"from rover range 4",
	"head used as chinese stuff",
	"head context neuro,brain,etc",
	"leg in chapter bodyparts",
	"foot in bodyparts2",
	"visual for prince",	
	"visual of integral",	
	"notes on restaurants in chinese",	
	"notes about brains",
	"notes music writing",
	"page 2 of notes on brains", 
	"notes page 3 brain", 
	"(1,1), (1,3), (4,4) (3,3) other stuff",
	"integrate in math",	
	"arrows pe,ep, eh",
	"arrows 1,-1",
	"forward cone for (bjorvika) range 5",
	"backward sideways cone for (bjorvika)",
	"sequences about fox",	
	"stories about (bjorvika)",	
	"context \"not only\"", 
	"\"come in\"",	
	"containing / matching \"blub blub\"", 
	"chinese kinds of meat", 
	"images prince", 
	"summary chapter interference",
	"showme greetings in norwegian",
	"paths from arrows pe,ep, eh",
	"paths from start to target limit 5",
	"paths to target3",	
	"a2 to b5 distance 10",
	"to a5",
	"from start",
	"from (1,6)",
	"a1 to b6 arrows then",
	"paths a2 to b5 distance 10",
	"from dog to cat",
        }

//******************************************************************

func main() {

	args := GetArgs()

	SST.MemoryInit()

	load_arrows := false
	ctx := SST.Open(load_arrows)

	var search SST.SearchParameters

	search_string := ""

	for a := 0; a < len(args); a++ {
		if strings.Contains(args[a]," ") {
			search_string += fmt.Sprintf("\"%s\"",args[a]) + " "
		} else {
			search_string += args[a] + " "
		}
	}

	if search_string == "" {
		ambient,key,_ := SST.GetTimeContext()
		search_string = "any chapter reminders context " + key + " " + ambient
	}

	search = SST.DecodeSearchField(search_string)
	
	Search(ctx,search,search_string)
	SST.Close(ctx)
	return
}

//**************************************************************

func Usage() {
	
	fmt.Printf("usage: ByYourCommand <search request>\n\n")
	fmt.Println("searchN4L <mytopic> chapter <mychapter>\n\n")
	fmt.Println("searchN4L range rover out of its depth")
	fmt.Println("searchN4L \"range rover\" \"out of its depth\"")
	fmt.Println("searchN4L from rover range 4")
	fmt.Println("searchN4L head used as \"version control\"")
	fmt.Println("searchN4L head context neuro)brain)etc")
	fmt.Println("searchN4L notes on restaurants in chinese")	
	fmt.Println("searchN4L notes about brains")
	fmt.Println("searchN4L notes music writing")
	fmt.Println("searchN4L page 2 of notes on brains") 
	fmt.Println("searchN4L notes page 3 brain") 
	fmt.Println("searchN4L (1,1) (1,3) (4,4) (3,3) other stuff")
	fmt.Println("searchN4L arrows pe)ep) eh")
	fmt.Println("searchN4L arrows 1)-1")
	fmt.Println("searchN4L forward cone for (bjorvika) range 5")
	fmt.Println("searchN4L sequences about fox")	
	fmt.Println("searchN4L context \"not only\"") 
	fmt.Println("searchN4L \"come on down\"")	
	fmt.Println("searchN4L chinese kinds of meat") 
	fmt.Println("searchN4L summary chapter interference")
	fmt.Println("searchN4L paths from arrows pe)ep) eh")
	fmt.Println("searchN4L paths from start to target2 limit 5")
	fmt.Println("searchN4L paths to target3")	
	fmt.Println("searchN4L a2 to b5 distance 10")
	fmt.Println("searchN4L to a5")
	fmt.Println("searchN4L from start")
	fmt.Println("searchN4L from (1)6)")
	fmt.Println("searchN4L a1 to b6 arrows then")
	fmt.Println("searchN4L paths a2 to b5 distance 10")
	fmt.Println("searchN4L <b5|a2> distance 10")

	flag.PrintDefaults()

	os.Exit(2)
}

//**************************************************************

func GetArgs() []string {

	flag.Usage = Usage
	verbosePtr := flag.Bool("v", false,"verbose")
	flag.Parse()

	if *verbosePtr {
		VERBOSE = true
	}

	return flag.Args()
}

//******************************************************************

func Search(ctx SST.PoSST, search SST.SearchParameters,line string) {

	if VERBOSE {
		fmt.Println("Your starting expression generated this set: ",line,"\n")
		fmt.Println(" - start set:",SL(search.Name))
		fmt.Println(" -      from:",SL(search.From))
		fmt.Println(" -        to:",SL(search.To))
		fmt.Println(" -   chapter:",search.Chapter)
		fmt.Println(" -   context:",SL(search.Context))
		fmt.Println(" -    arrows:",SL(search.Arrows))
		fmt.Println(" -    pagenr:",search.PageNr)
		fmt.Println(" - sequence/story:",search.Sequence)
		fmt.Println(" - limit/range/depth:",search.Range)
		fmt.Println()
	}

	// OPTIONS *********************************************

	name := search.Name != nil
	from := search.From != nil
	to := search.To != nil
	context := search.Context != nil
	chapter := search.Chapter != ""
	pagenr := search.PageNr > 0
	sequence := search.Sequence

	// Now convert strings into NodePointers

	arrowptrs,sttype := SST.ArrowPtrFromArrowsNames(ctx,search.Arrows)

	arrows := arrowptrs != nil
	sttypes := sttype != nil
	limit := 0

	if search.Range > 0 {
		limit = search.Range
	} else {
		limit = 10
	}

	nodeptrs := SST.SolveNodePtrs(ctx,search.Name,search.Chapter,search.Context,arrowptrs,limit)
	leftptrs := SST.SolveNodePtrs(ctx,search.From,search.Chapter,search.Context,arrowptrs,limit)
	rightptrs := SST.SolveNodePtrs(ctx,search.To,search.Chapter,search.Context,arrowptrs,limit)

	// SEARCH SELECTION *********************************************

	fmt.Println("------------------------------------------------------------------")
	fmt.Println(" Limiting to maximum of",limit,"results")

	// Table of contents

	if (context || chapter) && !name && !sequence && !pagenr && !(from || to) {

		ShowMatchingChapter(ctx,search.Chapter,search.Context,limit)
		ShowTime(ctx,search)
		return
	}

	// if we have name, (maybe with context, chapter, arrows)

	if name && ! sequence && !pagenr {

		fmt.Println("------------------------------------------------------------------")
		FindOrbits(ctx, nodeptrs, limit)
		ShowTime(ctx,search)
		return
	}

	if (name && from) || (name && to) {
		fmt.Printf("\nSearch \"%s\" has conflicting parts <to|from> and match strings\n",line)
		os.Exit(-1)
	}

	// Closed path solving, two sets of nodeptrs
	// if we have BOTH from/to (maybe with chapter/context) then we are looking for paths

	if from && to {

		fmt.Println("------------------------------------------------------------------")
		PathSolve(ctx,leftptrs,rightptrs,search.Chapter,search.Context,arrowptrs,sttype,limit)
		ShowTime(ctx,search)
		return
	}

	// Open causal cones, from one of these three

	if (name || from || to) && !pagenr && !sequence {

		// from or to or name
		
		if nodeptrs != nil {
			fmt.Println("------------------------------------------------------------------")
			CausalCones(ctx,nodeptrs,search.Chapter,search.Context,arrowptrs,sttype,limit)
			ShowTime(ctx,search)
			return
		}
		if leftptrs != nil {
			fmt.Println("------------------------------------------------------------------")
			CausalCones(ctx,leftptrs,search.Chapter,search.Context,arrowptrs,sttype,limit)
			ShowTime(ctx,search)
			return
		}
		if rightptrs != nil {
			fmt.Println("------------------------------------------------------------------")
			CausalCones(ctx,rightptrs,search.Chapter,search.Context,arrowptrs,sttype,limit)
			ShowTime(ctx,search)
			return
		}
	}
	
	// if we have page number then we are looking for notes by pagemap

	if (name || chapter) && pagenr {

		var notes []SST.PageMap

		if chapter {
			notes = SST.GetDBPageMap(ctx,search.Chapter,search.Context,search.PageNr)
			ShowNotes(ctx,notes)
			ShowTime(ctx,search)
			return
		} else {
			for n := range search.Name {
				notes = SST.GetDBPageMap(ctx,search.Name[n],search.Context,search.PageNr)
				ShowNotes(ctx,notes)
				ShowTime(ctx,search)
			}
			return
		}
	}

	// Look for axial trails following a particular arrow, like _sequence_ 

	if name && sequence || sequence && arrows {
		ShowStories(ctx,nodeptrs,arrowptrs,sttype,limit)
		ShowTime(ctx,search)
		return
	}

	// if we have sequence with arrows, then we are looking for sequence context or stories

	if arrows || sttypes {
		ShowMatchingArrows(ctx,arrowptrs,sttype)
		ShowTime(ctx,search)
		return
	}

	if VERBOSE {
		fmt.Println("Didn't find a solver")
	}

	ShowTime(ctx,search)

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

//******************************************************************
// SEARCH
//******************************************************************

func FindOrbits(ctx SST.PoSST, nptrs []SST.NodePtr, limit int) {
	
	var count int

	if VERBOSE {
		fmt.Println("Solver/handler: PrintNodeOrbit()")
	}

	for nptr := range nptrs {
		count++
		if count > limit {
			return
		}
		fmt.Print("\n",nptr,": ")
		SST.PrintNodeOrbit(ctx,nptrs[nptr],limit)
	}
}

//******************************************************************

func CausalCones(ctx SST.PoSST,nptrs []SST.NodePtr, chap string, context []string,arrows []SST.ArrowPtr, sttype []int,limit int) {

	var total int = 1

	if len(sttype) == 0 {
		sttype = []int{0,1,2,3}
	}

	if VERBOSE {
		fmt.Println("Solver/handler: GetFwdPathsAsLinks()")
	}

	for n := range nptrs {
		for st := range sttype {

			const maxlimit = SST.CAUSAL_CONE_MAXLIMIT
			fcone,_ := SST.GetFwdPathsAsLinks(ctx,nptrs[n],sttype[st],limit, maxlimit)

			if fcone != nil {
				fmt.Printf("%d. ",total)
				total += ShowCone(ctx,fcone,chap,context,limit)
			}

			if total > limit {
				return
			}

			if sttype[st] != 0 {
				bcone,_ := SST.GetFwdPathsAsLinks(ctx,nptrs[n],-sttype[st],limit, maxlimit)
				
				if bcone != nil {
					fmt.Printf("%d. ",total)
					total += ShowCone(ctx,bcone,chap,context,limit)
				}

				if total > limit {
					return
				}
			}
		}
	}

}

//******************************************************************

func PathSolve(ctx SST.PoSST,leftptrs,rightptrs []SST.NodePtr,chapter string,context []string,arrowptrs []SST.ArrowPtr,sttype []int,maxdepth int) {

	var Lnum,Rnum int
	var count int
	var left_paths, right_paths [][]SST.Link

	if leftptrs == nil || rightptrs == nil {
		return
	}

	// Find the path matrix

	if VERBOSE {
		fmt.Println("Solver/handler: GetEntireNCSuperConeAsLinks()")
	}

	var solutions [][]SST.Link
	var ldepth,rdepth int = 2,2

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		left_paths,Lnum = SST.GetEntireNCSuperConePathsAsLinks(ctx,"fwd",leftptrs,ldepth,chapter,context,maxdepth)
		right_paths,Rnum = SST.GetEntireNCSuperConePathsAsLinks(ctx,"bwd",rightptrs,rdepth,chapter,context,maxdepth)

		// try the reverse

		if Lnum == 0 || Rnum == 0 {
			left_paths,Lnum = SST.GetEntireNCSuperConePathsAsLinks(ctx,"bwd",leftptrs,ldepth,chapter,context,maxdepth)
			right_paths,Rnum = SST.GetEntireNCSuperConePathsAsLinks(ctx,"fwd",rightptrs,rdepth,chapter,context,maxdepth)
		}

		solutions,_ = SST.WaveFrontsOverlap(ctx,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > 0 {

			for s := 0; s < len(solutions); s++ {
				prefix := fmt.Sprintf(" - story path: ")
				PrintConstrainedLinkPath(ctx,solutions,s,prefix,chapter,context,arrowptrs,sttype)
			}
			count++
			break
		}

		if turn % 2 == 0 {
			ldepth++
		} else {
			rdepth++
		}
	}
}

//******************************************************************

func ShowMatchingArrows(ctx SST.PoSST,arrowptrs []SST.ArrowPtr,sttype []int) {

	if VERBOSE {
		fmt.Println("Solver/handler: GetDBArrowByPtr()/GetDBArrowBySTType")
	}

	for a := range arrowptrs {
		adir := SST.GetDBArrowByPtr(ctx,arrowptrs[a])
		inv := SST.GetDBArrowByPtr(ctx,SST.INVERSE_ARROWS[arrowptrs[a]])
		fmt.Printf("%3d. (st %d) %s -> %s,  with inverse = %3d. (st %d) %s -> %s\n",arrowptrs[a],SST.STIndexToSTType(adir.STAindex),adir.Short,adir.Long,inv.Ptr,SST.STIndexToSTType(inv.STAindex),inv.Short,inv.Long)
	}

	for st := range sttype {
		adirs := SST.GetDBArrowBySTType(ctx,sttype[st])
		for adir := range adirs {
			inv := SST.GetDBArrowByPtr(ctx,SST.INVERSE_ARROWS[adirs[adir].Ptr])
			fmt.Printf("%3d. (st %d) %s -> %s,  with inverse = %3d. (st %d) %s -> %s\n",adirs[adir].Ptr,SST.STIndexToSTType(adirs[adir].STAindex),adirs[adir].Short,adirs[adir].Long,inv.Ptr,SST.STIndexToSTType(inv.STAindex),inv.Short,inv.Long)
		}
	}
}

//******************************************************************

func ShowMatchingChapter(ctx SST.PoSST,chap string,context []string,limit int) {

	// This displays chapters and the unbroken context clusters within
        // them, with overlaps noted.

	if VERBOSE {
		fmt.Println("Solver/handler: ShowMatchingChapter()")
	}

	toc := SST.GetChaptersByChapContext(ctx,chap,context,limit)

	var chap_list []string

	for chaps := range toc {
		chap_list = append(chap_list,chaps)
	}

	sort.Strings(chap_list)

	for c := 0; c < len(chap_list); c++ {

		fmt.Printf("\n%d. Chapter: %s\n",c,chap_list[c])

		dim,clist,adj := SST.IntersectContextParts(toc[chap_list[c]])

		ShowContextFractions(dim,clist,adj)
	}
}

//******************************************************************

func ShowContextFractions(dim int,clist []string,adj [][]int) {
	
	for c := 0; c < len(adj); c++ {

		fmt.Printf("\n     %d.",c)

		for cp := 0; cp < len(adj[c]); cp++ {
			if adj[c][cp] > 0 {
				fmt.Printf(" relto ")
				break
			}
		}
		for cp := 0; cp < len(adj[c]); cp++ {
			if adj[c][cp] > 0 {
				fmt.Printf("%d,",cp)
			}
		}

		fmt.Printf(") %s\n",clist[c])
	}
}

//******************************************************************

func ShowChapterContexts(ctx SST.PoSST,chap string,context []string,limit int) {

	// This displays chapters and the fractionated context clusters within
        // them, emphasizing the atomic decomposition of context. Repeated/shared
	// context refers to the overlaps in the chapter search.

	if VERBOSE {
		fmt.Println("Solver/handler: ShowChapterContexts()")
	}

	toc := SST.GetChaptersByChapContext(ctx,chap,context,limit)

	// toc is a map by chapter with a list of list of context strings

	for c := range toc {

		fmt.Println("------------------------------------------------------------------")
		fmt.Printf("\n   Chapter context: %s\n",c)

		spectrum := SST.GetContextTokenFrequencies(toc[c])
		intent,ambient := SST.ContextIntentAnalysis(spectrum,toc[c])

		var intended string
		var common string

		for f := 0; f < len(intent); f++ {
			intended += fmt.Sprintf("\"%s\"",strings.TrimSpace(intent[f]))
			if f < len(intent)-1 {
				intended += ", "
			}
		}
		fmt.Print("\n   Exceptional context terms: ")
		SST.ShowText(intended,SST.SCREENWIDTH/2)
		fmt.Println()

		for f := 0; f < len(ambient); f++ {
			common += fmt.Sprintf("\"%s\"",ambient[f])
			if f < len(ambient)-1 {
				common += ", "
			}
		}
		fmt.Print("\n   Common context terms: ")
		SST.ShowText(common,SST.SCREENWIDTH/2)
		fmt.Println()
	}
	fmt.Println("\n")
	
}

//******************************************************************

func ShowStories(ctx SST.PoSST,nodeptrs []SST.NodePtr,arrowptrs []SST.ArrowPtr,sttypes []int,limit int) {

	fmt.Println("Solver/handler: HandleStories()")

	if arrowptrs == nil {
		arrowptrs,sttypes = SST.ArrowPtrFromArrowsNames(ctx,[]string{"!then!"})
	}
	
	stories := SST.GetSequenceContainers(ctx,nodeptrs,arrowptrs,sttypes,limit)

	for s := range stories {
		// if there is no unique match, the data contain a list of alternatives
		if stories[s].Axis == nil {
			fmt.Printf("%3d. %s\n",s,stories[s].Chapter)
		} else {
			fmt.Printf("The following story/sequence \"%s\"\n\n",stories[s].Chapter)
			for ev := range stories[s].Axis {
				fmt.Printf("\n%3d. %s\n",ev,stories[s].Axis[ev].Text)
				
				SST.PrintLinkOrbit(stories[s].Axis[ev].Orbits,SST.EXPRESS,1)
				SST.PrintLinkOrbit(stories[s].Axis[ev].Orbits,-SST.EXPRESS,1)
				SST.PrintLinkOrbit(stories[s].Axis[ev].Orbits,-SST.CONTAINS,1)
				SST.PrintLinkOrbit(stories[s].Axis[ev].Orbits,SST.LEADSTO,1)
				SST.PrintLinkOrbit(stories[s].Axis[ev].Orbits,-SST.LEADSTO,1)
				SST.PrintLinkOrbit(stories[s].Axis[ev].Orbits,SST.NEAR,1)
			}
		}
	}
}

//******************************************************************
// OUTPUT
//******************************************************************

func ShowCone(ctx SST.PoSST,cone [][]SST.Link,chap string,context []string,limit int) int {

	if len(cone) < 1 {
		return 0
	}

	if limit <= 0 {
		return 0
	}

	count := 0

	for s := 0; s < len(cone) && s < limit; s++ {
		SST.PrintSomeLinkPath(ctx,cone,s," - ",chap,context,limit)
		count++
	}

	return count
}

// **********************************************************

func ShowNode(ctx SST.PoSST,nptr []SST.NodePtr) string {

	var ret string

	for n := range nptr {
		node := SST.GetDBNodeByNodePtr(ctx,nptr[n])
		ret += fmt.Sprintf("\n    %.30s, ",node.S)
	}

	return ret
}

// **********************************************************

func PrintConstrainedLinkPath(ctx SST.PoSST, cone [][]SST.Link, p int, prefix string,chapter string,context []string,arrows []SST.ArrowPtr,sttype []int) {

	for l := 1; l < len(cone[p]); l++ {
		link := cone[p][l]

		if !ArrowAllowed(ctx,link.Arr,arrows,sttype) {
			return
		}
	}

	SST.PrintLinkPath(ctx,cone,p,prefix,chapter,context)
}

// **********************************************************

func ArrowAllowed(ctx SST.PoSST,arr SST.ArrowPtr, arrlist []SST.ArrowPtr, stlist []int) bool {

	st_ok := false
	arr_ok := false

	staidx := SST.GetDBArrowByPtr(ctx,arr).STAindex
	st := SST.STIndexToSTType(staidx)

	if arrlist != nil {
		for a := range arrlist {
			if arr == arrlist[a] {
				arr_ok = true
				break
			}
		}
	} else {
		arr_ok = true
	}

	if stlist != nil {
		for i := range stlist {
			if stlist[i] == st {
				st_ok = true
				break
			}
		}
	} else {
		st_ok = true
	}

	if st_ok || arr_ok {
		return true
	}

	return false
}

// **********************************************************

func ShowNotes(ctx SST.PoSST,notes []SST.PageMap) {

	var last string
	var lastc string

	for n := 0; n < len(notes); n++ {

		txtctx := SST.CONTEXT_DIRECTORY[notes[n].Context].Context
		
		if last != notes[n].Chapter || lastc != txtctx {

			fmt.Println("\n---------------------------------------------")
			fmt.Println("\nTitle:", notes[n].Chapter)
			fmt.Println("Context:", txtctx)
			fmt.Println("---------------------------------------------\n")

			last = notes[n].Chapter
			lastc = txtctx
		}

		for lnk := 0; lnk < len(notes[n].Path); lnk++ {
			
			text := SST.GetDBNodeByNodePtr(ctx,notes[n].Path[lnk].Dst)
			
			if lnk == 0 {
				fmt.Printf("\n [line %d]: ",notes[n].Line)
				fmt.Print(text.S," ")
			} else {
				arr := SST.GetDBArrowByPtr(ctx,notes[n].Path[lnk].Arr)
				fmt.Printf("(%s) %s ",arr.Long,text.S)
			}
		}
	}
}

// **********************************************************

func ShowTime(ctx SST.PoSST,search SST.SearchParameters) {

	ambient,key,now := SST.GetTimeContext()
	now_ctx := SST.UpdateSTMContext(ctx,ambient,key,now,search)
	SST.ShowContext(ambient,now_ctx,key)

}


