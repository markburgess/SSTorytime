package main

import (
        SST "SSTorytime"
	"fmt"
)


func main() {

	sst := SST.Open(false)

	l := SST.GetLastSawSection(sst)

	for r := range l {
		fmt.Println(l[r])
	}

	var nptr SST.NodePtr
	nptr.Class=2;
	nptr.CPtr=581

	x := SST.GetLastSawNPtr(sst,nptr)
	fmt.Println("X",x)

	SST.Close(sst)
}

