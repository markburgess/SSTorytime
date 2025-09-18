//******************************************************************
//
// Demo of node by node addition, assuming that the arrows are predefined
//
//******************************************************************

package main

import (
	"fmt"
        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	ctx := SST.Open(load_arrows)

	AddStory(ctx)
	LookupStory(ctx)

	SST.Close(ctx)
}

//******************************************************************

func AddStory(ctx SST.PoSST) {

	chap := "home and away"
	context := []string{""}
	var w float32 = 1.0

	n1 := SST.Vertex(ctx,"Mary had a little lamb",chap)
	n2 := SST.Vertex(ctx,"Whose fleece was dull and grey",chap)

	n3 := SST.Vertex(ctx,"And every time she washed it clean",chap)
	n4 := SST.Vertex(ctx,"It just went to roll in the hay",chap)

	n5 := SST.Vertex(ctx,"And when it reached a certain age ",chap)
	n6 := SST.Vertex(ctx,"She'd serve it on a tray",chap)

	SST.Edge(ctx,n1,"then",n2,context,w)

	// Vertex / Edge API users are not allowed to define new arrows

	SST.Edge(ctx,n2,"then",n3,context,w/2)
	SST.Edge(ctx,n2,"then",n5,context,w/2)

	// endings

	SST.Edge(ctx,n3,"then",n4,context,w)
	SST.Edge(ctx,n5,"then",n6,context,w)

}

//******************************************************************

func LookupStory(ctx SST.PoSST) {

	// Now reverse, print out the database paths

	start_set := SST.GetDBNodePtrMatchingName(ctx,"Mary had a","")
	_,sttype := SST.GetDBArrowsWithArrowName(ctx,"then")

	path_length := 4
	const maxlimit = SST.CAUSAL_CONE_MAXLIMIT

	for n := range start_set {

		paths,_ := SST.GetFwdPathsAsLinks(ctx,start_set[n],sttype,path_length,maxlimit)

		for p := range paths {

			if len(paths[p]) > 1 {
			
				fmt.Println("    Path",p," len",len(paths[p]))

				for l := 0; l < len(paths[p]); l++ {

					name := SST.GetDBNodeByNodePtr(ctx,paths[p][l].Dst).S
					fmt.Println("    ",l,"xx  --> ",
						paths[p][l].Dst,"=",name,"  , weight",
						paths[p][l].Wgt,"context",paths[p][l].Ctx)
				}
			}
		}
	}

}






