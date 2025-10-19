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
	"flag"
	"os"

        SST "SSTorytime"
)

var PAGENR int = 1

//******************************************************************

func main() {

	args := Init()

	load_arrows := true
	sst := SST.Open(load_arrows)

	chapter := ""

	for a := range args {
		chapter += args[a]
		if a < len(args)-1 {
			chapter += " "
		}
	}

	context := []string{""}

	Page(sst,chapter,context,PAGENR)
	fmt.Println()

	SST.Close(sst)
}

//**************************************************************

func Usage() {
	
	fmt.Printf("usage: Notes [chapter or section]\n")
	flag.PrintDefaults()

	os.Exit(2)
}

//**************************************************************

func Init() []string {

	pagePtr := flag.Int("page", 1, "page number for browsing")

	flag.Usage = Usage

	flag.Parse()
	args := flag.Args()

	PAGENR = *pagePtr

	if len(args) == 0 {
		fmt.Println("\nEnter a chapter to browse")
		os.Exit(-1)
	}

	SST.MemoryInit()

	return args
}

//******************************************************************

func Page(sst SST.PoSST,chapter string,context []string,page int) {

	var last string
	var lastc string

	notes := SST.GetDBPageMap(sst,chapter,context,page)

	for n := 0; n < len(notes); n++ {

		txtctx := SST.CONTEXT_DIRECTORY[notes[n].Context].Context
	
		if last != notes[n].Chapter || lastc != txtctx {
			fmt.Println("\n---------------------------------------------")
			fmt.Println("\nTitle:", notes[n].Chapter)
			fmt.Println("Context:", txtctx)
			fmt.Println("---------------------------------------------\n")
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








