//**************************************************************
//
// expt_etc_analysis.c
//
//**************************************************************

package SSTorytime

import (
	_ "github.com/lib/pq"
)

//**************************************************************

func CompleteETCTypes(sst PoSST,node Node) string {

	message := ""

	for st := 0; st < ST_TOP; st++ {

		if len(node.I[st]) > 0 {
			node.Psi,message = CollapsePsi(node,st)
		}
	}

	return message
}

//**************************************************************

func CollapsePsi(node Node,stindex int) (Etc, string) {

	// Follow the rules of SST Gamma(3,4) inference
	// convergent search to fixed point, ultimately event

	etc := node.Psi

	sttype := STIndexToSTType(stindex)

	message := ""

	arrow := ARROW_DIRECTORY[node.I[stindex][0].Arr].Long

	switch sttype {
		
	case NEAR:
		//fmt.Println("NEAR...")

	case -LEADSTO,LEADSTO:
		
		// skip bogus empty links
		for l := 0; l < len(node.I[stindex]); l++ {

			arrow = ARROW_DIRECTORY[node.I[stindex][l].Arr].Long

			if arrow == "empty" || arrow == "debug" {
				continue
			} else {
				break
			}
			
		}

		if arrow == "empty" || arrow == "debug" {
			return etc, message
		} else {
			etc.E = true
			etc.T = false
			etc.C = false
		}

	case CONTAINS:
		etc.T = true
		etc.T = false     // concept can't contain

	case -CONTAINS:
		etc.T = true

	case EXPRESS:
		if !etc.E {
			etc.C = true
			etc.T = true
		}

	case -EXPRESS:
		etc.T = true
		etc.C = false
	}
	
	message = "Node " + "\"" + node.S + "\"  (seems to be of type)  " + ShowPsi(etc)
	
	return etc,message
}


//
// expt_etc_analysis.c
//

