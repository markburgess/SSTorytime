//******************************************************************
//
// A path solving method, progressing from DAG to loop corrected
// paths containing cycles (like a quantum loop expansion)
//
// (This relies on the data from the double.n4l notes)
//
//******************************************************************

package main

import (
	"fmt"
        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := true
	ctx := SST.Open(load_arrows)

	// Contra colliding wavefronts as path integral solver

	const branching_limit = 2;
	const maxdepth = 7
	var ldepth,rdepth int = 2,2
	var Lnum,Rnum int
	var count int
	var left_paths, right_paths [][]SST.Link

	start_bc := "A1"
	end_bc := "B6"

	leftptrs := SST.GetDBNodePtrMatchingName(ctx,start_bc,"")
	rightptrs := SST.GetDBNodePtrMatchingName(ctx,end_bc,"")

	if leftptrs == nil || rightptrs == nil {
		fmt.Println("No paths available from end points")
		return
	}

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		left_paths,Lnum = SST.GetEntireConePathsAsLinks(ctx,"any",leftptrs[0],ldepth,branching_limit)
		right_paths,Rnum = SST.GetEntireConePathsAsLinks(ctx,"any",rightptrs[0],rdepth,branching_limit)
		
		solutions,loop_corrections := SST.WaveFrontsOverlap(ctx,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

		if len(solutions) > 0 {
			fmt.Println("-- T R E E ----------------------------------")
			fmt.Println("Path solution",count,"from",start_bc,"to",end_bc,"with lengths",ldepth,-rdepth)

			for s := 0; s < len(solutions); s++ {
				prefix := fmt.Sprintf(" - story %d: ",s)
				SST.PrintLinkPath(ctx,solutions,s,prefix,"",nil)
			}
			count++
			fmt.Println("-------------------------------------------")
		}

		if len(loop_corrections) > 0 {
			fmt.Println("++ L O O P S +++++++++++++++++++++++++++++++")
			fmt.Println("Path solution",count,"from",start_bc,"to",end_bc,"with lengths",ldepth,-rdepth)

			for s := 0; s < len(loop_corrections); s++ {
				prefix := fmt.Sprintf(" - story %d: ",s)
				SST.PrintLinkPath(ctx,loop_corrections,s,prefix,"",nil)
			}
			count++
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++")
		}

		if turn % 2 == 0 {
			ldepth++
		} else {
			rdepth++
		}
	}
}






