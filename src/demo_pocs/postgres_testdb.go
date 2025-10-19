package main

import (
        SST "SSTorytime"
)


func main() {

	sst := SST.Open(false)

	SST.Close(sst)
}

