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
	"strings"
	"encoding/json"


        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	ctx := SST.Open(load_arrows)

	context := []string{""}
	arrow := "then"

	Story(ctx,"chinese",context,"fox",arrow)

	SST.Close(ctx)
}

//******************************************************************

func Story(ctx SST.PoSST, chapter string,context []string,searchtext string,arrname string) {

	searchtext = strings.TrimSpace(searchtext)

	const limit = 10
	stories := SST.GetSequenceContainers(ctx,arrname,searchtext,chapter,context,limit)

	//for s := range stories {

	if stories == nil {
		fmt.Println("\nReturned table of contents, no unique story...\n")
	} else {
		story,_ := json.Marshal(stories)
		fmt.Println(string(story))
	}

	//}
}













