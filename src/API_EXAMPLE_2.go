//******************************************************************
//
// API demo HubJoin function to join a list of nodeptrs to a hub
//
//******************************************************************

package main

import (
	"fmt"
        SST "SSTorytime"
)

//******************************************************************

func main () {


	load_arrows := false
	sst := SST.Open(load_arrows)

	names := []string{"test_node1","test_node2","test_node3"}
	weights := []float32{0.2, 0.4, 1.0}
	context := []string{"some","context","tags"}

	var nodes []SST.Node
	var nptrs []SST.NodePtr

	// Create a set of nodes tolink

	for n := range names {
		nodes = append(nodes,SST.Vertex(sst,names[n],"my chapter"))
		nptrs = append(nptrs,nodes[n].NPtr)
	}

	// Create a hyperlink between all the nodes to a common hub, with arrow "then"

	created1 := SST.HubJoin(sst,"","",nptrs,"then",context,weights)
	fmt.Println("Creates hub node",created1)

	// Then create a container for all

	created2 := SST.HubJoin(sst,"mummy_node","",nptrs,"belongs to",nil,nil)
	fmt.Println("Creates hub node",created2)

	SST.Close(sst)
}










