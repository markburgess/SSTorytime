//******************************************************************
//
// Experiment with old CFEngine context gathering approach
// Look at all the ways of grabbing context
//
//******************************************************************

package main

import (
	"fmt"
	"time"
	"strings"
	"flag"
	"os"
        SST "SSTorytime"
)

//******************************************************************

func main() {

	load_arrows := false
	ctx := SST.Open(load_arrows)


	// Start with the classic time classes

	now := time.Now()
	c,slot := SST.DoNowt(now)
	fmt.Println("TIME_CLASSES",c,"\nSLOT",slot)
	name,_ := os.Hostname()
	fmt.Println("HOST",name)

	search := Init()
	search = strings.TrimSpace(search)

	// SEARCH

	//search := "motor neurons"
	fmt.Println("SEARCHING FOR",search)

	ngrams := SST.NewNgramMap()
	nmin := 1
	SST.Fractionate(search,ngrams,nmin) // -> ngrams

	// Look at ad hoc text input (small language model)
/*	
	input_stream := "/home/mark/Laptop/Work/SST/data_samples/MobyDick.dat"
//	input_stream = "../../../org-42/roam/how-i-org.org"
	SST.FractionateTextFile(input_stream)
	fmt.Println("INPUT_STREAM",input_stream)
*/
	for nsearch := SST.N_GRAM_MAX-1; nsearch >= nmin; nsearch-- {

		for each := range ngrams[nsearch] {

			fmt.Println("\nMatching longest N=",nsearch,"GRAM",each,"--------------------")
		
			for ntext := SST.N_GRAM_MIN; ntext < SST.N_GRAM_MAX; ntext++ {
				for g := range SST.STM_NGRAM_RANK[ntext] {
					
					if strings.Contains(g,each) {
						fmt.Println("ng",ntext,g,"by",each,nsearch)
					}
				}
			}
			
			// When we search for something already in the db, we need to look at 
			// the EntireCone to see what concepts context joins together
			// If we find a match, then we know that there are nodes related that might not contain
			// the search string directly, offering a level of indirection

			fmt.Println("Look for direct matches in the proposed notes context fields.....\n")			
			ctx_set := SST.GetDBContextsMatchingName(ctx,each)

			if len(ctx_set) > 0 {

				fmt.Println("Context relevance items",nsearch,each,"->",ctx_set)
			}

			nan := SST.GetDBNodeArrowNodeByContexts(ctx,"",[]string{each})

			for n := range nan {
				fr := SST.GetDBNodeByNodePtr(ctx,nan[n].NFrom).S
				arr := SST.GetDBArrowByPtr(ctx,nan[n].Arr).Long
				to := SST.GetDBNodeByNodePtr(ctx,nan[n].NTo).S
				fmt.Println("  - ",fr,"(",arr,")",to)
			}

			// Now look for nodes that match by name, and their orbits
			
			fmt.Println("Look for graph node names ....")
			
			chap := ""
			nptrs := SST.GetDBNodePtrMatchingName(ctx,each,chap)
			
			confidence := 2
			
			var patches = make(map[string]int)
			
			for i := range nptrs {
				n := SST.GetDBNodeByNodePtr(ctx,nptrs[i])
				fmt.Printf("\n- %d Probe from \"",i)
				SST.ShowText(n.S,80)
				patches[n.S]++
				fmt.Println("\"")
				paths,_ := SST.GetEntireConePathsAsLinks(ctx,"any",nptrs[i],confidence)
				for p := range paths {
					for d := range paths[p] {
						if len(paths[p][d].Ctx) > 1 {
							
							for each := range paths[p][d].Ctx {
								fmt.Println("    - found related context: ",paths[p][d].Ctx[each])
								patches[paths[p][d].Ctx[each]]++
							}
						}
					}
				}
			}
			
			fmt.Println("Summary of related contexts: ")
			
			for c := range patches {
				fmt.Println("  ... ",c," ->",patches[c])
			}
		}
	}

	SST.Close(ctx)
}

//**************************************************************

func Init() string {

	flag.Parse()
	args := flag.Args()

	search := ""

	for i := range args {
		search += args[i] + " "
	}

	SST.MemoryInit()

	return search
}

