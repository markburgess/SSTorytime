// **************************************************************************
//
// centrality_clustering.go
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	"sort"
	_ "github.com/lib/pq"

)

// **************************************************************************

func TallyPath(sst PoSST,path []Link,between map[string]int) map[string]int {

	// count how often each node appears in the different path solutions

	for leg := range path {
		n := GetDBNodeByNodePtr(&sst,path[leg].Dst)
		between[n.S]++
	}

	return between
}

// **************************************************************************

func BetweenNessCentrality(sst PoSST,solutions [][]Link) []string {

	var betweenness = make(map[string]int)

	for s := 0; s < len(solutions); s++ {
		betweenness = TallyPath(sst,solutions[s],betweenness)
	}

	var inv = make(map[int][]string)
 	var order []int

	for key := range betweenness {
		inv[betweenness[key]] = append(inv[betweenness[key]],key)
	}

	for key := range inv {
		order = append(order,key)
	}

	sort.Ints(order)

	var retval []string
	var betw string

	for key := len(order)-1; key >= 0; key-- {

		betw = fmt.Sprintf("%.2f : ",float32(order[key])/float32(len(solutions)))

		for el := 0; el < len(inv[order[key]]); el++ {

			betw += fmt.Sprintf("%s",inv[order[key]][el])
			if el < len(inv[order[key]])-1 {
				betw += ", "
			}
		}

		retval =  append(retval,betw)
	}
	return retval
}

// **************************************************************************

func SuperNodesByConicPath(solutions [][]Link, maxdepth int) [][]NodePtr {

	var supernodes [][]NodePtr
	
	for depth := 0; depth < maxdepth*2; depth++ {
		
		for p_i := 0; p_i < len(solutions); p_i++ {

			if depth == len(solutions[p_i])-1 {
				supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)
			}

			if depth > len(solutions[p_i])-1 {
				continue
			}

			supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_i][depth].Dst)

			for p_j := p_i+1; p_j < len(solutions); p_j++ {

				if depth < 1 || depth > len(solutions[p_j])-2 {
					break
				}

				if solutions[p_i][depth-1].Dst == solutions[p_j][depth-1].Dst && 
				   solutions[p_i][depth+1].Dst == solutions[p_j][depth+1].Dst {
					   supernodes = Together(supernodes,solutions[p_i][depth].Dst,solutions[p_j][depth].Dst)
				}
			}
		}		
	}

	return supernodes
}

// **************************************************************************

func SuperNodes(sst PoSST,solutions [][]Link, maxdepth int) []string {

	supernodes := SuperNodesByConicPath(solutions,maxdepth)

	var retval []string

	for g := range supernodes {

		super := ""

		for n := range supernodes[g] {
			node := GetDBNodeByNodePtr(&sst,supernodes[g][n])
			super += fmt.Sprintf("%s",node.S)
			if n < len(supernodes[g])-1 {
				super += ", "
			}
		}

		if g < len(supernodes)-1 {
			retval = append(retval,super)
		}
	}

	return retval
}


//
// centrality_clustering.go
//


