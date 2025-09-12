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
	"strings"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	ctx := SST.Open(load_arrows)

	searchtext := "notes"
	chaptext := ""
	context := []string{""}

	Search(ctx,chaptext,context,searchtext)

	SST.Close(ctx)
}

//******************************************************************

func Search(ctx SST.PoSST, chaptext string,context []string,searchtext string) {

	chaptext = strings.TrimSpace(chaptext)
	searchtext = strings.TrimSpace(searchtext)

	fmt.Println("--------------------------------------------------")
	fmt.Println("Looking for relevant nodes by",searchtext)
	fmt.Println("--------------------------------------------------")

	const maxdepth = 5
	
	var start_set []SST.NodePtr
	
	search_items := strings.Split(searchtext," ")
	
	for w := range search_items {
		fmt.Print("Looking for nodes like ",search_items[w],"...")
		start_set = append(start_set,SST.GetDBNodePtrMatchingName(ctx,search_items[w],chaptext)...)
	}

	fmt.Println("   Found possible relevant nodes:",start_set)

	for start := range start_set {

		for sttype := -SST.EXPRESS; sttype <= SST.EXPRESS; sttype++ {

			name :=  SST.GetDBNodeByNodePtr(ctx,start_set[start])

			const limit = 10
			allnodes := SST.GetFwdConeAsNodes(ctx,start_set[start],sttype,maxdepth,limit)
			
			if len(allnodes) > 1 {
				fmt.Println()
				fmt.Println("    -------------------------------------------")
				fmt.Printf("     Search text MATCH #%d via %s connection\n",start+1,SST.STTypeName(sttype))
				fmt.Printf("     (search %s => hit %s)\n",searchtext,name.S)
				fmt.Println("    -------------------------------------------")

				for l := range allnodes {
					fullnode := SST.GetDBNodeByNodePtr(ctx,allnodes[l])
					fmt.Println("     - SSType",SST.STTypeName(sttype)," cone item: ",fullnode.S,", found in",fullnode.Chap)
				}
			
				alt_paths,path_depth := SST.GetFwdPathsAsLinks(ctx,start_set[start],sttype,maxdepth,limit)
				
				if alt_paths != nil {
					
					fmt.Println("\n-- Forward",SST.STTypeName(sttype),"cone stories ----------------------------------")
					
					for p := 0; p < path_depth; p++ {
						SST.PrintLinkPath(ctx,alt_paths,p,"\nStory:","",nil)
					}
				}
				fmt.Printf("     (END %d)\n",start+1)
			}
		}
	}	
}









