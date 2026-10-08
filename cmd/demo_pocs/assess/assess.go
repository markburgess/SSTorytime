//******************************************************************
//
// assess
//
//******************************************************************

package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"path/filepath"
	"bytes"
	"encoding/json"

	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

//******************************************************************

func main() {

	// read and assess json files
	
	file := "../../../../../../MIM-fuzzyrobots/AmData/runs/warehouse-kpi-sonnet55-s3-20261004-131451/transcript.jsonl"

	//file = "../../../../../../MIM-fuzzyrobots/AmData/runs/warehouse-kpi-sonnet55-s3-20261004-131451/run.json"

	// file = "../../../../../../MIM-fuzzyrobots/AmData/runs/warehouse-kpi_consequence-sonnet55-s1-20261004-131046/manifest.json"
	
	filebytes, err := os.ReadFile(file)

	if err != nil {
		panic(err)
	}

	ext := filepath.Ext(file) 

	switch ext {
	case ".json":
		GetJSON(file,filebytes)
		
	case ".jsonl":
		lines := bytes.Split(filebytes, []byte("\n"))

		for _,l := range lines {
			GetJSON(file,l)
		}

	default:
		fmt.Println("File suffix",ext)
		os.Exit(0)
	}

	
	// ***********

	/*	
	sst := SST.Open(false)

	for i := 0; i < SST.N_GRAM_MAX; i++ {
		
		SST.STM_NGRAM_FREQ[i] = make(map[string]float64)
		SST.STM_NGRAM_LOCA[i] = make(map[string][]int)
		SST.STM_NGRAM_LAST[i] = make(map[string]int)
	}

	var lines = []string{zzz,one,two,three,four}
	
	for n,l := range lines {
		AFractionateText(n,l)
		SplitAndLook(sst,l)
	}*/	
}

//**************************************************************

func GetJSON(filename string,jsonData []byte) {

	var result map[string]any // 'any' is an alias for interface{}
	
	if err := json.Unmarshal(jsonData, &result); err != nil {
		
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	
	prefix := []string{"run","manifest","transcript"}
	last := strings.Split(filename,"/")
	name := last[len(last)-1]
	var main_keys = make(map[string]int)
	var order []string

	for k := range result {
		main_keys[k]++
		order=append(order,k)
	}	

	slices.Sort(order)
	
	for _,v := range order {
		fmt.Println("schema ",main_keys[v],v)
	}

	fmt.Println("------- START ------")
	
	for _,pr := range prefix {
		if strings.Contains(name,pr) {
			PrintTypedValue(result,main_keys,pr)
		}
	}
}

//**************************************************************

func PrintTypedValue(v any, keys map[string]int,prefix string) {

	switch val := v.(type) {

	// This is where the LHS key is handled at any level

	case map[string]any:
		fmt.Printf("%s [Struct]:\n", prefix)
		for k, subVal := range val {
			subprefix := fmt.Sprintf("%s > ",prefix)
			if keys[k] > 0 {
				fmt.Printf("\t ***>  K: %q -> ",k)
			} else {
				fmt.Printf("\t\t --->  K: %q -> ",k)
			}

			PrintTypedValue(subVal,keys, subprefix)
		}

	// Below are all the RHS
		
	case []any:
		fmt.Printf("%s [Array] (length %d):\n", prefix, len(val))
		for i, subVal := range val {
			fmt.Printf("\t\t\t  Idx [%d]: ", i)
			PrintTypedValue(subVal,keys,prefix)
		}
	case string:
		fmt.Printf("(string) %q\n", val)
	case float64:
		// Go unmarshals all JSON numbers to float64 by default
		fmt.Printf("(number) %v\n", val)
	case bool:
		fmt.Printf("(boolean) %v\n", val)
	case nil:
		fmt.Printf("(null) nil\n")
	default:
		fmt.Printf("(unknown) %v\n", val)
	}
}

//**************************************************************

func AFractionateText(n int,proto_text string) ([][]SST.Sentence,int) {

	pbsf,_ := SST.SplitIntoParaSentences(proto_text)

	count := 0

	for p := range pbsf {

		for s := range pbsf[p] {
			count++

			for f := range pbsf[p][s].Frags {

				change_set := Fractionate(pbsf[p][s].Frags[f],count,SST.STM_NGRAM_FREQ,SST.N_GRAM_MIN)

				// Update global n-gram frequencies for fragment, and location histories

				for n := 0; n < SST.N_GRAM_MAX; n++ {
					for ng := range change_set[n] {
						ngram := change_set[n][ng]
							
						SST.STM_NGRAM_FREQ[n][ngram]++
					}
				}
			}
		}
	}

	return pbsf,count
}


//**************************************************************

func SplitAndLook(sst SST.PoSST, line string) {
	
	fmt.Println("=============",line,"===========")

	doc_psf,_ := SST.SplitIntoParaSentences(line)

	var score  = make(map[string]int)
	var orbits = make(map[string]int)
	var arrows = make(map[string]int)

	var stm    = make(map[string]string)
	var stm_ngram_freq [SST.N_GRAM_MAX]map[string]float64

	for i := 0; i < SST.N_GRAM_MAX; i++ {
		stm_ngram_freq[i] = make(map[string]float64)
	}

	for p := 0; p < len(doc_psf); p++ {
		
		for s := 0; s < len(doc_psf[p]); s++ {

			for f := 0; f < len(doc_psf[p][s].Frags); f++ {
			
				text := strings.TrimSpace(doc_psf[p][s].Frags[f])
				
				stm["start"] = stm["prev"]
				stm["prev"] = stm["now"]
				stm["now"] = strings.ToLower(text)

				try := stm["now"]

				if len(try) < 3 {
					try = "x"
				}

				if len(stm["prev"]) > 2 {
					try += stm["prev"]
				}

				if len(stm["start"]) > 2 {
					try += stm["start"]
				}
				
				//fmt.Printf("-LOOK %s ==> \n",try)
				
				score = LookUp(sst,try,score,stm_ngram_freq)
				
			}
		}
	}
	
	fmt.Println("CUMULATIVE SCORES\n")
	for k, v := range score {
		fmt.Println(" +>",k,v)
	}
	
	fmt.Println("CUMULATIVE ORBITS\n")
	for k, v := range orbits {
		fmt.Println(" o>",k,v)
	}
	
	fmt.Println("ARROWS\n")
	for k, v := range arrows {
		fmt.Println(" ->",k,v)
	}
	
	
}

//**************************************************************

func LookUp(sst SST.PoSST, try string,score map[string]int,freq [SST.N_GRAM_MAX]map[string]float64) map[string]int {

	var count int
	chap := "Fuzzy Robot Semantics"
	cntx := []string{}
	seq := false
	arr := []SST.ArrowPtr{}
	limit := 10
	var orbits = make(map[string]int)
	var arrows = make(map[string]int)

	change_set := Fractionate(try,count,freq,SST.N_GRAM_MIN)
	
	for n := 0; n < SST.N_GRAM_MAX; n++ {
		for ng := range change_set[n] {
			ngram := change_set[n][ng]
			nptrs := SST.GetDBNodePtrMatchingNCCS(sst,ngram,chap,cntx,arr,seq,limit)
			
			if len(nptrs) > 0 {
				fmt.Printf("\n ... (%s)",ngram)
			}
			
			for _,nptr := range nptrs {
				orb := SST.GetNodeOrbit(&sst,nptr,"",limit)
				for _,np := range orb {
					for _,nx := range np {
						snode := SST.GetDBNodeByNodePtr(&sst,nx.Dst)
						fmt.Printf("   NEXT (%v,%v)",nx.Arrow,snode.S)
						orbits[strings.ToLower(snode.S)]++
						arrows[nx.Arrow]++
					}
				}
				
				node := SST.GetDBNodeByNodePtr(&sst,nptr)
				//score[w]++
				if node.I[SST.SELFPTR] != nil {
					ctx,_ := SST.GetDBContextByPtr(&sst,node.I[SST.SELFPTR][0].Ctx)
					fmt.Println("   ->",node.S,"\"\t ...classified as...",ctx)
					score[ctx]++
				}
			}
		}
	}

	return score
}

//**************************************************************

func Fractionate(frag string,L int,frequency [SST.N_GRAM_MAX]map[string]float64,min int) [SST.N_GRAM_MAX][]string {
	
	// A round robin cyclic buffer for taking fragments and extracting
	// n-ngrams of 1,2,3,4,5,6 words separateed by whitespace, passing
	
	var rrbuffer [SST.N_GRAM_MAX][]string
	var change_set [SST.N_GRAM_MAX][]string

	words := strings.Split(frag," ")

	for w := range words {
		rrbuffer,change_set = NextWord(words[w],rrbuffer)
	}

	return change_set
}

//**************************************************************

func NextWord(frag string,rrbuffer [SST.N_GRAM_MAX][]string) ([SST.N_GRAM_MAX][]string,[SST.N_GRAM_MAX][]string) {

	// Word by word, we form a superposition of scores from n-grams of different lengths
	// as a simple sum. This means lower lengths will dominate as there are more of them
	// so we define intentionality proportional to the length also as compensation

	var change_set [SST.N_GRAM_MAX][]string

	for n := 0; n < SST.N_GRAM_MAX; n++ {
		
		// Pop from round-robin

		if (len(rrbuffer[n]) > n-1) {
			rrbuffer[n] = rrbuffer[n][0:n]
		}
		
		// Push new to maintain length

		rrbuffer[n] = append(rrbuffer[n],frag)

		// Assemble the key, only if complete cluster
		
		if (len(rrbuffer[n]) > n-1) {
			
			var key string
			
			for j := 0; j < n; j++ {
				key = key + rrbuffer[n][j]
				if j < n-1 {
					key = key + " "
				}
			}

			key = SST.CleanNgram(key)

			/*if SST.ExcludedByBindings(SST.CleanNgram(rrbuffer[n][0]),key,SST.CleanNgram(rrbuffer[n][n-1])) {
				continue
			}*/

			change_set[n] = append(change_set[n],key)
		}
	}

	frag = SST.CleanNgram(frag)

	//if !SST.ExcludedByBindings(frag,frag,frag) {
		change_set[1] = append(change_set[1],frag)
	//}

	return rrbuffer,change_set
}

