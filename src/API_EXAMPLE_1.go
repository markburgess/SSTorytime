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
	sst := SST.Open(load_arrows)

	AddStory(sst)
	LookupStory(sst)

	SST.Close(sst)
}

//******************************************************************

func AddStory(sst SST.PoSST) {

	chap := "home and away"
	context := []string{""}
	var w float32 = 1.0

	n1 := SST.Vertex(sst,"Mary had a little lamb",chap)
	n2 := SST.Vertex(sst,"Whose fleece was dull and grey",chap)

	n3 := SST.Vertex(sst,"And every time she washed it clean",chap)
	n4 := SST.Vertex(sst,"It just went to roll in the hay",chap)

	n5 := SST.Vertex(sst,"And when it reached a certain age ",chap)
	n6 := SST.Vertex(sst,"She'd serve it on a tray",chap)

	SST.Edge(sst,n1,"then",n2,context,w)

	// Vertex / Edge API users are not allowed to define new arrows

	SST.Edge(sst,n2,"then",n3,context,w/2)
	SST.Edge(sst,n2,"then",n5,context,w/2)

	// endings

	SST.Edge(sst,n3,"then",n4,context,w)
	SST.Edge(sst,n5,"then",n6,context,w)

}

//******************************************************************

func LookupStory(sst SST.PoSST) {

	// Now reverse, print out the database paths

	start_set := SST.GetDBNodePtrMatchingName(sst,"Mary had a","")
	_,sttype := SST.GetDBArrowsWithArrowName(sst,"then")

	path_length := 4
	const maxlimit = SST.CAUSAL_CONE_MAXLIMIT

	for n := range start_set {

		paths,_ := SST.GetFwdPathsAsLinks(sst,start_set[n],sttype,path_length,maxlimit)

		for p := range paths {

			if len(paths[p]) > 1 {
			
				fmt.Println("    Path",p," len",len(paths[p]))

				for l := 0; l < len(paths[p]); l++ {

					name := SST.GetDBNodeByNodePtr(sst,paths[p][l].Dst).S
					fmt.Println("    ",l,"xx  --> ",
						paths[p][l].Dst,"=",name,"  , weight",
						paths[p][l].Wgt,"context",paths[p][l].Ctx)
				}
			}
		}
	}

}






