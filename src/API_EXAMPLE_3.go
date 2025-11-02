//******************************************************************
//
// Maze solver, knowing start and end
//
//******************************************************************

package main

import (
	"fmt"

        SST "SSTorytime"
)

var path [9][]string

//******************************************************************

func main() {

	path[0] = []string{"maze_a7","maze_b7","maze_b6","maze_c6","maze_c5","maze_b5","maze_b4","maze_a4","maze_a3","maze_b3","maze_c3","maze_d3","maze_d2","maze_e2","maze_e3","maze_f3","maze_f4","maze_e4","maze_e5","maze_f5","maze_f6","maze_g6","maze_g5","maze_g4","maze_h4","maze_h5","maze_h6","maze_i6"}
	path[1] = []string{"maze_d1","maze_d2"}
	path[2] = []string{"maze_f1","maze_f2","maze_e2"}
	path[3] = []string{"maze_f2","maze_g2","maze_h2","maze_h3","maze_g3","maze_g2"}
	path[4] = []string{"maze_b1","maze_c1","maze_c2","maze_b2","maze_b1"}
	path[5] = []string{"maze_b7","maze_b8","maze_c8","maze_c7","maze_d7","maze_d6","maze_e6","maze_e7","maze_f7","maze_f8"}
	path[6] = []string{"maze_d7","maze_d8","maze_e8","maze_e7"}
	path[7] = []string{"maze_f7","maze_g7","maze_g8","maze_h8","maze_h7"}
	path[8] = []string{"maze_a2","maze_a1"}

	load_arrows := true
	sst := SST.Open(load_arrows)

	// Add the paths to a fresh database

	for p := range path {
		for leg := 1; leg < len(path[p]); leg++ {	

			chap := "solve maze"
			context := []string{""}
			var w float32 = 1.0

			nfrom := SST.Vertex(sst,path[p][leg-1],chap)
			nto := SST.Vertex(sst,path[p][leg],chap)

			SST.Edge(sst,nfrom,"fwd",nto,context,w)
		}
	}

	Solve(sst)

	SST.Close(sst)
}

//******************************************************************

func Solve(sst SST.PoSST) {

	// Contra colliding wavefronts as path integral solver

	const mindepth = 1;
	const maxdepth = 16
	var count int
	var arrowptrs []SST.ArrowPtr
	var sttype []int
	var context []string

	start_bc := "maze_a7"
	end_bc := "maze_i6"
	chapter := ""


	leftptrs := SST.GetDBNodePtrMatchingName(sst,start_bc,"")
	rightptrs := SST.GetDBNodePtrMatchingName(sst,end_bc,"")

	if leftptrs == nil || rightptrs == nil {
		fmt.Println("No paths available from end points")
		return
	}

	solutions := SST.GetPathsAndSymmetries(sst,leftptrs,rightptrs,chapter,context,arrowptrs,sttype,mindepth,maxdepth)

	if len(solutions) > 0 {
		for s := 0; s < len(solutions); s++ {
			prefix := fmt.Sprintf(" - story %d: ",s)
			SST.PrintLinkPath(sst,solutions,s,prefix,"",nil)
		}
		count++
	}
	
}

// **********************************************************

func ShowNode(sst SST.PoSST,nptr []SST.NodePtr) string {

	var ret string

	for n := range nptr {
		node := SST.GetDBNodeByNodePtr(sst,nptr[n])
		ret += node.S + ","
	}

	return ret
}

// **********************************************************

func ShowNodePath(sst SST.PoSST,lnk []SST.Link) string {

	var ret string

	for n := range lnk {
		node := SST.GetDBNodeByNodePtr(sst,lnk[n].Dst)
		arrs := SST.GetDBArrowByPtr(sst,lnk[n].Arr).Long
		ret += fmt.Sprintf("(%s) -> %s ",arrs,node.S)
	}

	return ret
}






