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
	"os"
        SST "SSTorytime"
)

var path [8][]string

//******************************************************************

func main() {

	load_arrows := true
	ctx := SST.Open(load_arrows)

	Solve(ctx)

	SST.Close(ctx)
}

//******************************************************************

func Solve(ctx SST.PoSST) {

	// Contra colliding wavefronts as path integral solver

	const maxdepth = 16
	var ldepth,rdepth int = 1,1
	var Lnum,Rnum int
	var left_paths, right_paths [][]SST.Link

	start_bc := "a7"
	end_bc := "i6"

	leftptrs := SST.GetDBNodePtrMatchingName(ctx,start_bc,"")
	rightptrs := SST.GetDBNodePtrMatchingName(ctx,end_bc,"")

	if leftptrs == nil || rightptrs == nil {
		fmt.Println("No paths available from end points")
		return
	}

	cntx := []string{""}

	// Note, because retrieval is non-deterministic in order, if we make this small
	// the probability of getting the same set decreases. So grab everything.

	const limit = 30

	for turn := 0; ldepth < maxdepth && rdepth < maxdepth; turn++ {

		left_paths,Lnum = SST.GetEntireNCConePathsAsLinks(ctx,"fwd",leftptrs[0],ldepth,"",cntx,limit)
		xleft_paths,Lnumx := SST.GetEntireConePathsAsLinks(ctx,"fwd",leftptrs[0],ldepth,limit)

		right_paths,Rnum = SST.GetEntireNCConePathsAsLinks(ctx,"bwd",rightptrs[0],rdepth,"",cntx,limit)
		xright_paths,Rnumx := SST.GetEntireConePathsAsLinks(ctx,"bwd",rightptrs[0],rdepth,limit)	

		if Lnum != Lnumx {
			fmt.Println("LEFT sizes differ at depth",ldepth,"=",Lnum,Lnumx)
			fmt.Println("EntireNCcone",left_paths)
			fmt.Println("Entirecone",xleft_paths)
			os.Exit(-1)
		}

		if Diff(left_paths,xleft_paths) {
			fmt.Println("LEFT SETS differ at depth",ldepth)
			fmt.Println("EntireNCcone",left_paths)
			fmt.Println("Entirecone",xleft_paths)
			os.Exit(-1)
		}

		if Rnum != Rnumx {
			fmt.Println("RIGHT sizes differ at depth",rdepth,"=",Rnum,Rnumx)
			os.Exit(-1)
		}

		if Diff(right_paths,xright_paths) {
			fmt.Println("RIGHT SETS differ at depth",rdepth)
			os.Exit(-1)
		}

		ldepth++
		rdepth++
	}
}

// **********************************************************

func Diff(left,right [][]SST.Link) bool {

	var L = make(map[SST.Link]bool)
	var R = make(map[SST.Link]bool)

	retval := false

	for path := 0; path < len(left); path++ {
		for l := 0; l < len(left[path]); l++ {
			L[left[path][l]] = true
		}
	}

	for path := 0; path < len(left); path++ {
		for l := 0; l < len(left[path]); l++ {
			R[right[path][l]] = true
		}
	}

	for r := range L {
		if !R[r] {
			fmt.Println("L not in R",r)
			retval = true
		}
	}

	for l := range R {
		if !L[l] {
			fmt.Println("R not in L",l)
			retval = true
		}
	}

	return retval
}







