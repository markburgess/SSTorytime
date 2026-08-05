package main

import (
	SST "github.com/markburgess/SSTorytime/internal/sst"
)


func main() {

	sst := SST.Open(false)

	SST.Close(sst)
}

