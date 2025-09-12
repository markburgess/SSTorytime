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
	ctx := SST.Open(load_arrows)

	nodeptrs := SST.GetDBNodePtrMatchingName(ctx,"a1","slit")

	fmt.Println("Found",nodeptrs)

	for n := range nodeptrs {

		const maxdepth = 5
		context := []string{"physics","slits"}
		chapter := "slit"

		const limit = 10
		alt_paths,path_depth := SST.GetEntireConePathsAsLinks(ctx,"fwd",nodeptrs[n],maxdepth,limit)
		
		if alt_paths != nil {
			
			for p := 0; p < path_depth; p++ {
				SST.PrintLinkPath(ctx,alt_paths,p,"\nStory:",chapter,context)
			}
		}
	}
}







