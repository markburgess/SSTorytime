
package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"

        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	for goes := 0; goes < 10; goes ++ {

		fmt.Println("\n\nEnter some text:")
		
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		
		SearchToJSON(sst,text)
	}

	SST.Close(sst)
}

//******************************************************************

func SearchToJSON(sst SST.PoSST, text string) {

	text = strings.TrimSpace(text)

	var start_set []SST.NodePtr

	search_items := strings.Split(text," ")

	for w := range search_items {
		start_set = append(start_set,SST.GetDBNodePtrMatchingName(sst,search_items[w],"")...)
	}

	for s := range start_set {
		r := SST.JSONNodeOrbit(sst, start_set[s]) 
		fmt.Println(s,r)
	}
	
}










