//******************************************************************
//
// Exploring how to present a search text, with API
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

const (
	host     = "localhost"
	port     = 5432
	user     = "sstoryline"
	password = "sst_1234"
	dbname   = "sstoryline"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	cntx := []string{ "yes", "thank you", "(food)"}
	chapter := "chinese"
	name := "lamb"
	const limit = 10
	nptrs := SST.GetDBNodePtrMatchingNCC(sst,name,chapter,cntx,nil,limit)

	fmt.Println("RETURNED",nptrs)

	fmt.Println("\nExpanding..")

	for n := range nptrs {
		node := SST.GetDBNodeByNodePtr(sst,nptrs[n])
		fmt.Println("Found:",node.S)
	}

	SST.Close(sst)
}

