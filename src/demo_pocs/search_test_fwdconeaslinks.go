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
//	"strings"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	searchtext := "(limian)"
	chaptext := "chinese"
	context := []string{"city","exercise"}

	Search(sst,chaptext,context,searchtext)

	SST.Close(sst)
}

//******************************************************************

func Search(sst SST.PoSST, chaptext string,context []string,searchtext string) {

	start_set := SST.GetDBNodePtrMatchingName(sst,searchtext,chaptext)
	
	for sttype := -SST.EXPRESS; sttype <= SST.EXPRESS; sttype++ {
		
		name :=  SST.GetDBNodeByNodePtr(sst,start_set[0])
		
		alt_paths,path_depth := SST.GetFwdPathsAsLinks(sst,start_set[0],sttype,2)
		
		if alt_paths != nil {
			
			fmt.Println("\n-------\n",SST.STTypeName(sttype),"\n NPTR=",start_set[0]," with NAME",name,"\n-----")
			
			for p := 0; p < path_depth; p++ {

				SST.PrintLinkPath(sst,alt_paths,p,"\nStory:","",nil)
			}
		}
	}
}






