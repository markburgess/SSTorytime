//******************************************************************
//
// Find all nodes that start stories with a named arrow type
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

	chapter := ""
        context := []string{"waves"}

	arrow := "freq"

	fmt.Println("\nLook for nodes starting some thread with a arrow ",arrow,"in the context",context)

        matches1 := SST.GetNCCNodesStartingStoriesForArrow(sst,arrow,"",chapter,context)

        for p := range matches1 {

                n := SST.GetDBNodeByNodePtr(sst,matches1[p])

                fmt.Println("-start with",n.S,"in",n.Chap)
        }

	fmt.Println("\nLook for node startinga thread with arrow",arrow,"regardlss of context")

        matches2,_ := SST.GetNodesStartingStoriesForArrow(sst,arrow)

        for p := range matches2 {

                n := SST.GetDBNodeByNodePtr(sst,matches2[p])

                fmt.Println("-start with",n.S,"in",n.Chap)
        }

	SST.Close(sst)
}

