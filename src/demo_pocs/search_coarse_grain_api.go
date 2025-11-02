//******************************************************************
//
// Find <end|start> transition matrix and calculate symmetries
//
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
	sst := SST.Open(load_arrows)

	// Contra colliding wavefronts as path integral solver

	const mindepth = 2
	const maxdepth = 15

	context := []string{""}
	chapter := ""

	start_bc := []string{"start"}
//	start_bc := []string{"f7"}
	end_bc := []string{"target_1","target_2","target_3"}
//	end_bc := []string{"h7"}

/*	context := []string{""}
	chapter := "slit"

	start_bc := []string{"A1"}
	end_bc := []string{"B6"}*/

	var start_set,end_set []SST.NodePtr

	for n := range start_bc {
		start_set = append(start_set,SST.GetDBNodePtrMatchingName(sst,start_bc[n],"")...)
	}

	for n := range end_bc {
		end_set = append(end_set,SST.GetDBNodePtrMatchingName(sst,end_bc[n],"")...)
	}

	var arrowptrs []SST.ArrowPtr
	var sttypes []int

	solutions := SST.GetPathsAndSymmetries(sst,start_set,end_set,chapter,context,arrowptrs,sttypes,mindepth,maxdepth)

	var count int

	// ***** paths ****

	fmt.Println("-- T R E E ----------------------------------")
	fmt.Println("Path solution",count,"from",start_bc,"to",end_bc)
	
	for s := 0; s < len(solutions); s++ {
		prefix := fmt.Sprintf(" - story %d: ",s)
		SST.PrintLinkPath(sst,solutions,s,prefix,"",nil)
	}
	count++
	fmt.Println("-------------------------------------------")

	// **** Process symmetries ***

	supernodes := SST.GetPathTransverseSuperNodes(sst,solutions,maxdepth)

	fmt.Println("Look for coarse grains, final matroid:",len(supernodes))

	for g := range supernodes {
		fmt.Print("\n    - Super node ",g," = {")
		for n := range supernodes[g] {
			node :=SST.GetDBNodeByNodePtr(sst,supernodes[g][n])
			fmt.Print(node.S,",")
		}
		fmt.Println("}")
	}

}

//******************************************************************

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









