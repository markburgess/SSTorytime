//******************************************************************
//
// Exploring how to present a search text, with API
//
// Prepare:
// cd examples
// ../src/N4L-db -u Mary.n4l, e.g. try type Mary example, type 1
//
//******************************************************************

package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	for goes := 0; goes < 10; goes ++ {

		fmt.Println("\n\nEnter some text:")
		
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		
		Search(sst,text)
	}

	SST.Close(sst)
}

//******************************************************************

func Search(sst SST.PoSST, text string) {

	text = strings.TrimSpace(text)

	const maxdepth = 5

	sttype := SST.LEADSTO
	fmt.Print("Choose a search type: ")

	for t := SST.NEAR; t <= SST.EXPRESS; t++ {
		fmt.Print(t,"=",SST.STTypeName(t),", ")
	}

	fmt.Scanf("%d",&sttype)

	var start_set []SST.NodePtr

	search_items := strings.Split(text," ")

	for w := range search_items {
		start_set = append(start_set,SST.GetDBNodePtrMatchingName(sst,search_items[w],"poet")...)
	}

	for start := range start_set {

		name :=  SST.GetDBNodeByNodePtr(sst,start_set[start])

		fmt.Println()
		fmt.Println("-------------------------------------------")
		fmt.Printf(" SEARCH MATCH %d: (%s -> %s)\n",start,text,name.S)
		fmt.Println("-------------------------------------------")

		allnodes := SST.GetFwdConeAsNodes(sst,start_set[start],sttype,maxdepth)
		
		for l := range allnodes {
			fullnode := SST.GetDBNodeByNodePtr(sst,allnodes[l])
			fmt.Println("   - Fwd ",SST.STTypeName(sttype)," cone item: ",fullnode.S,", found in",fullnode.Chap)
		}

		alt_paths,path_depth := SST.GetFwdPathsAsLinks(sst,start_set[start],sttype,maxdepth)
			
		if alt_paths != nil {
			
			fmt.Printf("\n-- Forward",SST.STTypeName(sttype),"cone stories ----------------------------------\n")
			
			for p := 0; p < path_depth; p++ {
				SST.PrintLinkPath(sst,alt_paths,p,"\nStory:","",nil)
			}
		}
		fmt.Printf("     (END %d)\n",start)
	}

	// Now look at the arrow content

	fmt.Println("\nLooking at relations...\n")

	matching_arrows := SST.GetDBArrowsMatchingArrowName(sst,text)

	relns := SST.GetDBNodeArrowNodeMatchingArrowPtrs(sst,"poet",nil,matching_arrows)

	for r := range relns {

		from := SST.GetDBNodeByNodePtr(sst,relns[r].NFrom)
		to := SST.GetDBNodeByNodePtr(sst,relns[r].NFrom)
		//st := relns[r].STType
		arr := SST.ARROW_DIRECTORY[relns[r].Arr].Long
		wgt := relns[r].Wgt
		actx := relns[r].Ctx
		fmt.Println("See also: ",from.S,"--(",arr,")->",to.S,"\n       (... wgt",wgt,"in the contexts",actx,")\n")

	}
}

//******************************************************************

func IsNew(nptr SST.NodePtr,levels [][]SST.NodePtr) bool {

	for l := range levels {
		for e := range levels[l] {
			if levels[l][e] == nptr {
				return false
			}
		}
	}
	return true
}








