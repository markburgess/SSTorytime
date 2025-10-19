//******************************************************************
//
// Try out neighbour search for all ST stypes together
//
// Prepare:
// cd examples
// ../src/N4L-db -u chinese.n4l
//
//******************************************************************

package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	nodeptrs := SST.GetDBNodePtrMatchingName(sst,"a1","slit")

	fmt.Println("Found",nodeptrs)

	for n := range nodeptrs {

		const maxdepth = 5
		context := []string{"physics","slits"}
		chapter := "slit"

		const limit = 10
		alt_paths,path_depth := SST.GetEntireConePathsAsLinks(sst,"fwd",nodeptrs[n],maxdepth,limit)
		
		if alt_paths != nil {
			
			for p := 0; p < path_depth; p++ {
				SST.PrintLinkPath(sst,alt_paths,p,"\nStory:",chapter,context)
			}
		}
	}
}







