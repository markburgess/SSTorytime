//******************************************************************
//
// Exploring how to present knowledge systematically, e.g.
// e.g. review/review for an exam!
//  version 3 with axial backbone as a reference to simplify
//
//******************************************************************

package main

import (
	"fmt"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := true
	sst := SST.Open(load_arrows)

	context := []string{""}

	Page(sst,"notes on chinese",context,1)
	fmt.Println("\n........")
	Page(sst,"notes on chinese",context,2)
	fmt.Println("\n........")
	Page(sst,"notes on chinese",context,3)

	SST.Close(sst)
}

//******************************************************************

func Page(sst SST.PoSST,chapter string,context []string,page int) {

	var last string
	var lastc string

	notes := SST.GetDBPageMap(sst,chapter,context,page)

	for n := 0; n < len(notes); n++ {

		txtctx := SST.CONTEXT_DIRECTORY[notes[n].Context].Context
		
		if last != notes[n].Chapter || lastc != txtctx {
			fmt.Println("\n\nTitle:", notes[n].Chapter)
			fmt.Println("Context:", txtctx)
			last = notes[n].Chapter
			lastc = txtctx
		}

		for lnk := 0; lnk < len(notes[n].Path); lnk++ {
			
			text := SST.GetDBNodeByNodePtr(sst,notes[n].Path[lnk].Dst)
			
			if lnk == 0 {
				fmt.Print("\n",text.S," ")
			} else {
				arr := SST.GetDBArrowByPtr(sst,notes[n].Path[lnk].Arr)
				fmt.Printf("(%s) %s ",arr.Long,text.S)
			}
		}
	}
}








