
package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)
	nptr := SST.GetDBNodePtrMatchingName(sst,"lamb","")
	const maxdepth = 4

	multicone := "{\n"

	for n := 0; n < len(nptr); n++ {

		cone,_ := SST.GetFwdPathsAsLinks(sst,nptr[n],1,maxdepth)
		json := SST.JSONCone(sst,cone,"",nil)

		const empty = 5

		if len(json) > empty {
			multicone += fmt.Sprintf("\"%v\" : %s ",nptr[n],json)
			if n < len(nptr)-1 {
				multicone += ",\n"
			}
		}
	}

	multicone += "\n}\n"

	fmt.Println(multicone)
	SST.Close(sst)
}










