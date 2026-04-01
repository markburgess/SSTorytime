package main

import (
	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)


func main() {

	sst := SST.Open(false)

	SST.Close(sst)
}

