//******************************************************************
//
// Exploring how to present a search text, with API
//
// Prepare:
// cd examples
// ../src/N4L-db -u chinese.n4l
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

	reader := bufio.NewReader(os.Stdin)

	var context []string

	fmt.Println("\n\nEnter chapter text (e.g. poetry,chinese, etc):")
	chaptext, _ := reader.ReadString('\n')

	for goes := 0; goes < 10; goes ++ {

		fmt.Println("\nCurrent context:",context)
		fmt.Println("Enter newcontext text:")

		cntext, _ := reader.ReadString('\n')

		if cntext != "" {
			w := strings.Split(cntext," ")
			for c := range w {
				context = append(context,strings.TrimSpace(w[c]))
			}
		}
		fmt.Println("\n\nEnter search text:")
		searchtext, _ := reader.ReadString('\n')

		context := []string{"poem"}
		
		Search(sst,chaptext,context,searchtext)
	}

	SST.Close(sst)
}

//******************************************************************

func Search(sst SST.PoSST, chaptext string,context []string,searchtext string) {

	chaptext = strings.TrimSpace(chaptext)
	searchtext = strings.TrimSpace(searchtext)

	fmt.Println("--------------------------------------------------")
	fmt.Println("Looking for relevant nodes by",searchtext)
	fmt.Println("--------------------------------------------------")

	const maxdepth = 3
	
	var start_set []SST.NodePtr
	
	search_items := strings.Split(searchtext," ")
	
	for w := range search_items {
		fmt.Print("Looking for nodes like ",search_items[w],"...")
		start_set = append(start_set,SST.GetDBNodePtrMatchingName(sst,search_items[w],chaptext)...)
	}

	fmt.Println("   Found possible relevant nodes:",start_set)

	for start := range start_set {

		for sttype := -SST.EXPRESS; sttype <= SST.EXPRESS; sttype++ {

			name :=  SST.GetDBNodeByNodePtr(sst,start_set[start])

			allnodes := SST.GetFwdConeAsNodes(sst,start_set[start],sttype,maxdepth)
			
			if len(allnodes) > 1 {
				fmt.Println()
				fmt.Println("    -------------------------------------------")
				fmt.Printf("     Search text MATCH #%d via %s connection\n",start+1,SST.STTypeName(sttype))
				fmt.Printf("     (search %s => hit %s)\n",searchtext,name.S)
				fmt.Println("    -------------------------------------------")

				for l := range allnodes {
					fullnode := SST.GetDBNodeByNodePtr(sst,allnodes[l])
					fmt.Println("     - SSType",SST.STTypeName(sttype)," cone item: ",fullnode.S,", found in",fullnode.Chap)
				}

				// Conic proper time paths
			
				alt_paths,path_depth := SST.GetFwdPathsAsLinks(sst,start_set[start],sttype,maxdepth)
				
				if alt_paths != nil {
					
					fmt.Println("\n-- Forward (",SST.STTypeName(sttype),") cone stories ----------------------------------\n")
					
					for p := 0; p < path_depth; p++ {
						SST.PrintLinkPath(sst,alt_paths,p,"\nStory:","",nil)
					}
				}
				fmt.Printf("     (END %d)\n",start+1)
			}
		}
	}
	
	
	// Now look at the arrow content
	
	fmt.Println()
	fmt.Println("--------------------------------------------------")
	fmt.Println("checking whether any arrows also match search",searchtext,"(in any context)")
	fmt.Println("--------------------------------------------------")
	
	matching_arrows := SST.GetDBArrowsMatchingArrowName(sst,searchtext)
	
	relns := SST.GetDBNodeArrowNodeMatchingArrowPtrs(sst,chaptext,context,matching_arrows)
	
	for r := range relns {
		
		from := SST.GetDBNodeByNodePtr(sst,relns[r].NFrom)
		to := SST.GetDBNodeByNodePtr(sst,relns[r].NTo)
		arr := SST.ARROW_DIRECTORY[relns[r].Arr].Long
		wgt := relns[r].Wgt
		actx := relns[r].Ctx
		fmt.Println("   See also: ",from.S,"--(",arr,")->",to.S,"\n       (... wgt",wgt,"in the contexts",actx,")\n")
		
	}
}









