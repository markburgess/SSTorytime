// Notes CLI logic (ported from upstream notes).
package main

import (
	"context"
	"fmt"

	SST "github.com/markburgess/SSTorytime/internal/sst"
)

func Page(ctx context.Context, sst SST.PoSST, chapter string, context []string, page int) {

	var last string
	var lastc string
	const maxlimit = 100

	notes := SST.GetDBPageMap(ctx, sst, chapter, context, page, maxlimit)

	for n := 0; n < len(notes); n++ {

		txtctx := sst.CONTEXT_DIRECTORY[notes[n].Context].Context

		if last != notes[n].Chapter || lastc != txtctx {
			fmt.Println("\n---------------------------------------------")
			fmt.Println("\nTitle:", notes[n].Chapter)
			fmt.Println("Context:", txtctx)
			fmt.Println("---------------------------------------------\n")
			last = notes[n].Chapter
			lastc = txtctx
		}

		for lnk := 0; lnk < len(notes[n].Path); lnk++ {

			text := SST.GetDBNodeByNodePtr(ctx, &sst, notes[n].Path[lnk].Dst)

			if lnk == 0 {
				fmt.Print("\n", text.S, " ")
			} else {
				arr := SST.GetDBArrowByPtr(ctx, &sst, notes[n].Path[lnk].Arr)
				fmt.Printf("(%s) %s ", arr.Long, text.S)
			}
		}
	}
}
