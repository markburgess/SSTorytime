//******************************************************************
//
// Demo of accessing postgres with custom data structures and arrays
// converting to the package library format
//
//******************************************************************

package main

import (
	"fmt"
        SST "SSTorytime"
)

//******************************************************************

func main() {

        sst := SST.Open(false)

	fmt.Println("Reset..")

	var n1,n2,n3,n4,n5,n6 SST.Node
	var lnk12,lnk23,lnk34,lnk25,lnk56,lnk62 SST.Link
	
	n1.NPtr = SST.NodePtr{ CPtr : 1, Class: SST.LT128}
	n1.S = "Mary had a little lamb"
	n1.Chap = "home and away"

	n2.NPtr = SST.NodePtr{ CPtr : 2, Class: SST.LT128}
	n2.S = "Whose fleece was dull and grey"
	n2.Chap = "home and away"

	n3.NPtr = SST.NodePtr{ CPtr : 3, Class: SST.LT128}
	n3.S = "And everytime she washed it clean"
	n3.Chap = "home and away"

	n4.NPtr = SST.NodePtr{ CPtr : 4, Class: SST.LT128}
	n4.S = "It just went to roll in the hay"
	n4.Chap = "home and away"

	n5.NPtr = SST.NodePtr{ CPtr : 5, Class: SST.LT128}
	n5.S = "And every bar that Mary went"
	n5.Chap = "home and away"

	n6.NPtr = SST.NodePtr{ CPtr : 6, Class: SST.LT128}
	n6.S = "Was hot and loud and gay"
	n6.Chap = "home and away"

	lnk12.Arr = 77
	lnk12.Wgt = 0.34
	lnk12.Ctx = []string{"fairy castles","angel air"}
	lnk12.Dst = n2.NPtr

	lnk23.Arr = 77
	lnk23.Wgt = 0.34
	lnk23.Ctx = []string{"fairy castles","angel air"}
	lnk23.Dst = n3.NPtr

	lnk34.Arr = 77
	lnk34.Wgt = 0.34
	lnk34.Ctx = []string{"fairy castles","angel air"}
	lnk34.Dst = n4.NPtr

	lnk25.Arr = 77
	lnk25.Wgt = 0.34
	lnk25.Ctx = []string{"steamy hot tubs"}
	lnk25.Dst = n5.NPtr

	lnk56.Arr = 77
	lnk56.Wgt = 0.34
	lnk56.Ctx = []string{"steamy hot tubs","lady gaga"}
	lnk56.Dst = n6.NPtr

	// Add a loop

	lnk62.Arr = 77
	lnk62.Wgt = 0.99
	lnk62.Ctx = []string{"danger will robinson"}
	lnk62.Dst = n2.NPtr

	sttype := SST.LEADSTO

	n1 = SST.IdempDBAddNode(sst, n1)
	n2 = SST.IdempDBAddNode(sst, n2)
	SST.AppendDBLinkToNode(sst,n1.NPtr,lnk12,sttype)

	n2 = SST.IdempDBAddNode(sst, n2)
	n3 = SST.IdempDBAddNode(sst, n3)
	SST.AppendDBLinkToNode(sst,n2.NPtr,lnk23,sttype)

	n3 = SST.IdempDBAddNode(sst, n3)
	n4 = SST.IdempDBAddNode(sst, n4)
	SST.AppendDBLinkToNode(sst,n3.NPtr,lnk34,sttype)

	n2 = SST.IdempDBAddNode(sst, n2)
	n5 = SST.IdempDBAddNode(sst, n5)
	SST.AppendDBLinkToNode(sst,n2.NPtr,lnk25,sttype)

	n5 = SST.IdempDBAddNode(sst, n5)
	n6 = SST.IdempDBAddNode(sst, n6)
	SST.AppendDBLinkToNode(sst,n5.NPtr,lnk56,sttype)

	// Add loop
	SST.AppendDBLinkToNode(sst,n6.NPtr,lnk62,sttype)

	fmt.Println("----------------------------------")
	fmt.Println("Node section hypersurfaces:")

	const maxdepth = 8

	for depth := 0; depth < maxdepth; depth++ {
		val := SST.GetFwdConeAsNodes(sst,n1.NPtr,sttype,depth)
		fmt.Println("As NodePtr(s) fwd from",n1,"depth",depth)
		for l := range val {
			fmt.Println("   - Step",val[l])
		}

	}

	fmt.Println("----------------------------------")
	fmt.Println("Link section hypersurfaces:")

	for depth := 0; depth < maxdepth; depth++ {
		val := SST.GetFwdConeAsLinks(sst,n1.NPtr,sttype,depth)
		fmt.Println("Search as Links fwd from",n1,"depth",depth)
		for l := range val {
			fmt.Println("   - Step",val[l])
		}
	}

	fmt.Println("----------------------------------")
	fmt.Println("Link proper time normal paths:")

	for depth := 0; depth < maxdepth; depth++ {

		fmt.Println("Searching paths of length",depth,"/",maxdepth,"from",n1.NPtr)

		paths,_ := SST.GetFwdPathsAsLinks(sst,n1.NPtr,sttype,depth)

		for p := range paths {

			if len(paths[p]) > 1 {
			
				fmt.Println("    Path",p," len",len(paths[p]))

				for l := 0; l < len(paths[p]); l++ {
					fmt.Println("    ",l,"xx  --> ",paths[p][l].Dst,"weight",paths[p][l].Wgt)
				}
			}
		}
	}

	SST.Close(sst)
}








