//**************************************************************
//
//  API.go
//
//**************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	_ "github.com/lib/pq"

)

//**************************************************************

func Vertex(sst *PoSST,name,chap string) Node {

        // Automatic NPtr numbering

	var n Node

	n.S = name
	n.Chap = chap

	return IdempDBAddNode(sst,n)
}

// **************************************************************************

func Edge(sst *PoSST,from Node,arrow string,to Node,context []string,weight float32) (ArrowPtr,int) {

	arrowptr,sttype := GetDBArrowsWithArrowName(sst,arrow)

	var link Link

	link.Arr = arrowptr
	link.Dst = to.NPtr
	link.Wgt = weight
	link.Ctx = TryContext(sst,context)

	IdempDBAddLink(sst,from,link,to)

	return arrowptr,sttype
}

// **************************************************************************

func HubJoin(sst *PoSST,name,chap string,nptrs []NodePtr,arrow string,context []string,weight []float32) Node {

	// Create a container node joining several other nodes in a list, like a hyperlink

	if nptrs == nil {
		fmt.Println("Call to HubJoin with a null list of pointers")
		os.Exit(-1)
	}

	if weight == nil {
		for n := 0; n < len(nptrs); n++ {
			weight = append(weight,1.0)
		}
	}

	if len(nptrs) != len(weight) {
		fmt.Println("Call to HubJoin with inconsistent node/weight pointer arrays: dimensions ",len(nptrs),"vs",len(weight))
		os.Exit(-1)
	}

	var chaps = make(map[string]int)

	if name == "" {
		name = "hub_"+arrow+"_"
		for n := range nptrs {
			name += fmt.Sprintf("(%d,%d)",nptrs[n].Class,nptrs[n].CPtr)
			node := GetDBNodeByNodePtr(sst,nptrs[n])
			chaps[node.Chap]++
		}
	}

	var to Node

	to.S = name

	if chap != "" {
		to.Chap = chap
	} else 	if chap == "" && len(chaps) == 1 {
		for ch := range chaps {
			to.Chap = ch
		}
	}

	container := IdempDBAddNode(sst,to)

	arrowptr,_ := GetDBArrowsWithArrowName(sst,arrow)

	for nptr := range nptrs {

		var link Link
		link.Arr = arrowptr
		link.Dst = container.NPtr
		link.Wgt = weight[nptr]
		link.Ctx = TryContext(sst,context)
		from := GetDBNodeByNodePtr(sst,nptrs[nptr])
		IdempDBAddLink(sst,from,link,container)
	}

	return GetDBNodeByNodePtr(sst,container.NPtr)
}





//
//  API.go
//

