//******************************************************************
//
// Demo of accessing SST in postgres
// Test of future (causal) cone and independent path retrieval
// (By NPtr and Link reference only)
//
// Prepare:
// cd examples
// ../src/N4L-db -u doors.n4l
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

	fmt.Println("Node causal forward cone doors.N4l")

	const maxdepth = 8
	sttype := SST.LEADSTO

	levels := make([][]SST.NodePtr,maxdepth)

	// Get the start node

	start_set := SST.GetDBNodePtrMatchingName(sst,"start","")

	for start := range start_set {

		fmt.Println(" ---------------------------------")
		fmt.Println(" - Total forward cone from: ",start_set[start])
		fmt.Println(" ---------------------------------")

		allnodes := SST.GetFwdConeAsNodes(sst,start_set[start],sttype,maxdepth)
		
		for l := range allnodes {
			fmt.Println("   - node",allnodes[l])
		}
		
		for depth := 0; depth < maxdepth; depth++ {
			
			fmt.Println(" ---------------------------------")
			fmt.Println(" - Cone layers ",depth," from: ",start_set[start])
			fmt.Println(" ---------------------------------")
			
			levels[depth] = make([]SST.NodePtr,0)
			
			allnodes := SST.GetFwdConeAsNodes(sst,start_set[start],sttype,depth)
			
			for l := range allnodes {
				if IsNew(allnodes[l],levels) {
					levels[depth] = append(levels[depth],allnodes[l])
				}
			}			
			fmt.Println("level",depth,levels[depth])
		}
		
		fmt.Println("Link proper time normal paths:")
			
		for start := range start_set {
			
			for depth := 0; depth < maxdepth; depth++ {
				
				fmt.Println("Searching paths of length",depth,"/",maxdepth,"from",start_set[start])
				
				paths,_ := SST.GetFwdPathsAsLinks(sst,start_set[start],sttype,depth)
				
				for p := range paths {
					
					if len(paths[p]) > 1 {
						
						fmt.Println("    Path",p," len",len(paths[p]))
						
						for l := 0; l < len(paths[p]); l++ {
							fmt.Print(" --> ",paths[p][l].Dst," weight",paths[p][l].Wgt)
						}
						
						fmt.Println()
					}
				}
			}
		}
	}		
	SST.Close(sst)
}

//******************************************************************

func IsNew(nptr SST.NodePtr,levels [][]SST.NodePtr) bool {

	for l := range levels {
		for e := range levels[l] {
			if levels[l][e] == nptr {
				return false
			}
		}
	}
	return true
}








