//******************************************************************
//
// Exploring how to present node text
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

	searchtext := "hypo"
	chaptext := "brain"
	context := []string{""}

	fmt.Println("Look for",searchtext,"\n")
	Search(sst,chaptext,context,searchtext)

	searchtext = "S1"
	chaptext = ""
	context = []string{"physics"}

	fmt.Println("Look for",searchtext,"\n")
	Search(sst,chaptext,context,searchtext)

	SST.Close(sst)
}

//******************************************************************

func Search(sst SST.PoSST, chaptext string,context []string,searchtext string) {
	
	nptrs := SST.GetDBNodePtrMatchingName(sst,searchtext,chaptext)

	for nptr := range nptrs {
		fmt.Print(nptr,": ")
		SST.PrintNodeOrbit(sst,nptrs[nptr],100)


	}

}









