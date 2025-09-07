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

	load_arrows := true
	ctx := SST.Open(load_arrows)

	// Contra colliding wavefronts as path integral solver

	const maxdepth = 6
	var ldepth,rdepth int = 1,1
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

		left_paths,Lnum = SST.GetEntireConePathsAsLinks(ctx,"any",leftptrs[0],ldepth)		
		right_paths,Rnum = SST.GetEntireConePathsAsLinks(ctx,"any",rightptrs[0],rdepth)		
		
		solutions,loop_corrections := WaveFrontsOverlap(ctx,left_paths,right_paths,Lnum,Rnum,ldepth,rdepth)

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

// **********************************************************

func WaveFrontsOverlap(ctx SST.PoSST,left_paths,right_paths [][]SST.Link,Lnum,Rnum,ldepth,rdepth int) ([][]SST.Link,[][]SST.Link) {

	// The wave front consists of Lnum and Rnum points left_paths[len()-1].
	// Any of the

	var solutions [][]SST.Link
	var loops [][]SST.Link

	// Start expanding the waves from left and right, one step at a time, alternately

	leftfront := WaveFront(left_paths,Lnum)
	rightfront := WaveFront(right_paths,Rnum)

	fmt.Println("\n  Left front radius",ldepth,":",ShowNode(ctx,leftfront))
	fmt.Println("  Right front radius",rdepth,":",ShowNode(ctx,rightfront))

	incidence := NodesOverlap(ctx,leftfront,rightfront)
	
	for lp := range incidence {

		rp := incidence[lp]

		var LRsplice []SST.Link		

		LRsplice = LeftJoin(LRsplice,left_paths[lp])
		adjoint := SST.AdjointLinkPath(right_paths[rp])
		LRsplice = RightComplementJoin(LRsplice,adjoint)

		fmt.Printf("...SPLICE PATHS L%d with R%d.....\n",lp,rp)
		fmt.Println("Left tendril",ShowNodePath(ctx,left_paths[lp]))
		fmt.Println("Right tendril",ShowNodePath(ctx,right_paths[rp]))
		fmt.Println("Right adjoint:",ShowNodePath(ctx,adjoint))
		fmt.Println(".....................\n")

		if IsDAG(LRsplice) {
			solutions = append(solutions,LRsplice)
		} else {
			loops = append(loops,LRsplice)
		}
	}

	fmt.Printf("  (found %d touching solutions)\n",len(incidence))
	return solutions,loops
}

// **********************************************************

func WaveFront(path [][]SST.Link,num int) []SST.NodePtr {

	// assemble the cross cutting nodeptrs of the wavefronts

	var front []SST.NodePtr

	for l := 0; l < num; l++ {
		front = append(front,path[l][len(path[l])-1].Dst)
	}

	return front
}

// **********************************************************

func NodesOverlap(ctx SST.PoSST,left,right []SST.NodePtr) map[int]int {

	var LRsplice = make(map[int]int)
	var list string

	// Return coordinate pairs of partial paths to splice

	for l := 0; l < len(left); l++ {
		for r := 0; r < len(right); r++ {
			if left[l] == right[r] {
				node := SST.GetDBNodeByNodePtr(ctx,left[l])
				list += node.S+", "
				LRsplice[l] = r
			}
		}
	}

	if len(list) > 0 {
		fmt.Println("  (i.e. waves impinge",len(LRsplice),"times at: ",list,")\n")
	}

	return LRsplice
}

// **********************************************************

func LeftJoin(LRsplice,seq []SST.Link) []SST.Link {

	for i := 0; i < len(seq); i++ {

		LRsplice = append(LRsplice,seq[i])
	}

	return LRsplice
}

// **********************************************************

func RightComplementJoin(LRsplice,adjoint []SST.Link) []SST.Link {

	// len(seq)-1 matches the last node of right join
	// when we invert, links and destinations are shifted

	for j := 1; j < len(adjoint); j++ {
		LRsplice = append(LRsplice,adjoint[j])
	}

	return LRsplice
}

// **********************************************************

func IsDAG(seq []SST.Link) bool {

	var freq = make(map[SST.NodePtr]int)

	for i := range seq {
		freq[seq[i].Dst]++
	}

	for n := range freq {
		if freq[n] > 1 {
			return false
		}
	}

	return true
}

// **********************************************************

func ShowNode(ctx SST.PoSST,nptr []SST.NodePtr) string {

	var ret string

	for n := range nptr {
		node := SST.GetDBNodeByNodePtr(ctx,nptr[n])
		ret += node.S + ","
	}

	return ret
}

// **********************************************************

func ShowNodePath(ctx SST.PoSST,lnk []SST.Link) string {

	var ret string

	for n := range lnk {
		node := SST.GetDBNodeByNodePtr(ctx,lnk[n].Dst)
		arrs := SST.GetDBArrowByPtr(ctx,lnk[n].Arr).Long
		ret += fmt.Sprintf("(%s) -> %s ",arrs,node.S)
	}

	return ret
}






