package postgrestestdb

import (
	SST "github.com/markburgess/SSTorytime/internal/sst"
)


func Main() {

	sst := SST.Open(false)

	SST.Close(sst)
}

