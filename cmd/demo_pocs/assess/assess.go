//******************************************************************
//
// assess the definitions of an AI agent
//
//******************************************************************

package main

import (
	"fmt"
	"os"
	//	"slices"
	"strings"
	"path/filepath"
	"bytes"
	"encoding/json"

	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

//******************************************************************

func main() {

	SST.InitSSTorytime()
	
	// read and assess json files
	
	dir := "../../../../../../MIM-fuzzyrobots/AmData/runs/warehouse-kpi-sonnet55-s3-20261004-131451/"

	HandleAgentMarkDown(dir)
	
	files := []string{"transcript.jsonl","run.json","manifest.json"}
	
	for _,file := range files {
		cmpt := dir+file
		fmt.Println("\nAssessing run : ",cmpt)
		AssessJsonType(cmpt)
	}

}

//**************************************************************

func HandleAgentMarkDown(dir string) {

	const percentage = 100

	files, err := os.ReadDir(dir)
	
	if err != nil {
		fmt.Println("Error",err)
		return
	}

	outputfile := "AGENT_ASSESSMENT_GENERATED.n4l"

	fp, err := os.Create(outputfile)

	if err != nil {
		fmt.Println("Failed to open file for writing: ",outputfile)
		os.Exit(-1)
	}

	defer fp.Close()
	
	for _, entry := range files {
		
		ext := filepath.Ext(entry.Name())
		name := entry.Name()

		if strings.HasSuffix(ext,".md") {
			_, err := entry.Info()
			if err != nil {
				fmt.Printf(" ** Could not get info for %s: %v\n", name, err)
				continue
			}

			fmt.Println("Assess MD file",dir+name)

			psf,L := SST.FractionateMarkdown(dir+name)

			SST.AnnotateFractionIntent(psf)
		
			ranking1 := SST.SelectByRunningIntent(psf,L,percentage)
			ranking2 := SST.SelectByStaticIntent(psf,L,percentage)
			selection := SST.MergeSelections(ranking1,ranking2)
			
			const minN = 1 // >= N_GRAM_MIN
			const maxN = 4 // <= N_GRAM_MAX

			f,s,ff,ss := SST.ExtractIntentionalTokens(L,selection,minN,maxN)

			GenerateOutput(fp,selection,L,percentage,f,s,ff,ss)
		}		
	}

}

//*******************************************************************

func GenerateOutput(fp *os.File,selection []SST.TextRank,L int, percentage float64,anom_by_part[][]string,ambi_by_part[][]string,all_anom[]string,all_ambi[]string) {

	// See AddMandatory() in N4L.go for reserved names (TBD, collect these one day as const)

	var collected_fragments = make(map[string][]string)

	filealias := "agent_assessment"
	
	fmt.Fprintf(fp," - System Prompt Declaration %s\n",filealias)

	fmt.Fprintf(fp,"\n# (begin) ************\n")

	// Lookup tables of contents from Markdown

	SST.CompileTabular(fp,filealias,SST.TABLES)
	SST.CompileTOC(fp,filealias,SST.SECTIONS)

	// 

	fmt.Fprintf(fp,"\n#######################################")
	fmt.Fprintf(fp,"\n :: _sequence_ , %s::\n", filealias)
	fmt.Fprintf(fp,"#######################################\n")
	
	var partcheck = make(map[string]bool)
	var parts []string
	var lastpart string
	var already = make(map[string]bool)

	for i := range selection {

		if len(selection[i].Fragment) < 1 {
			continue
		}
		
		var part string
		context := SST.SpliceSet(ambi_by_part[selection[i].Partition])
		
		if len(selection[i].Title) > 0 {
			part = selection[i].Title
		} else {
			part = "No section"
		}
		
		// Add context from n = 2,3 fractions
		
		if part != lastpart {
			if len(context) > 0 {
				fmt.Fprintf(fp,"\n#######################################")
				fmt.Fprintf(fp,"\n :: %s ::\n",context)
				fmt.Fprintf(fp,"#######################################\n")
			}
			
			lastpart = part
			
		}

		fmt.Fprintf(fp,"\n@sen%d   %s\n\n",selection[i].Order,SST.SanitizeParen(selection[i].Fragment))		
		fmt.Fprintf(fp,"              \" (%s) %s\n",SST.INV_CONT_FOUND_IN_S,part)
		
		AddIntentionalContext(collected_fragments,part,anom_by_part[selection[i].Partition],already)
		
		if !partcheck[part] {
			parts = append(parts,part)
			partcheck[part] = true
		}
	}
	
	fmt.Fprintf(fp,"\n -:: _sequence_ , %s::\n", filealias)
	fmt.Fprintf(fp,"\n# (end) ************\n")

	// some stats

	fmt.Fprintf(fp,"\n# Final fraction %.2f of requested %.2f\n",float64(len(selection)*100)/float64(L),percentage)

	//WriteSampleSelections(fp,selection,L)

	// add the parts' fragments

	fmt.Fprintf(fp,"\n #\n # Concepts by part / region \n #\n")

	for key := range collected_fragments {

		fmt.Fprintf(fp,"\n\n %s\n",key)
		for _,s := range collected_fragments[key] {
			fmt.Fprintf(fp,"              \" (%s) %s\n",SST.CONT_FRAG_S,s)
		}
	}

	// document the parts

	fmt.Fprintf(fp,"\n #\n # SUMMARY OF CONTEXT AND SIGNIFICANT FRAGMENTS\n #\n")

	fmt.Fprintf(fp,"\n :: themes and topics you might want to annotate/replace ::\n")

	fmt.Fprintf(fp,"\n :: parts, sections ::\n")

	for p := range parts {
		
		container,exists := SST.DOC_DIRECTORY[parts[p]]

		if exists {
			fmt.Fprintf(fp,"\n %s (%s) %s \n",parts[p],SST.INV_CONT_FRAG_IN_S,container)
		} else {
			fmt.Fprintf(fp,"\n %s \n",parts[p])
		}

		if p < len(ambi_by_part) {
			for w := range ambi_by_part[p] {
				fmt.Fprintf(fp,"  #AMBI %s\n",ambi_by_part[p][w])
			}
		}

		if p < len(anom_by_part) {
			for w := range anom_by_part[p] {
				fmt.Fprintf(fp,"   #INTENT %s\n",anom_by_part[p][w])
			}
		}
	}


	/* whole document summary

	for w := range all_ambi {
		fmt.Fprintf(fp," # %s\n",all_ambi[w])
	}

	for w := range all_anom {
		fmt.Fprintf(fp,"  # %s\n",all_anom[w])
	} */		
}

//*******************************************************************

func AddIntentionalContext(collected map[string][]string,key string,ctx []string,already map[string]bool) {

	for w := 0; w < len(ctx); w++ {

		if !already[ctx[w]] {
			collected[key] = append(collected[key],ctx[w])
			already[ctx[w]] = true
		}
	}
}

//**************************************************************

func AssessJsonType(filename string) {

	filebytes, err := os.ReadFile(filename)
	
	if err != nil {
		fmt.Println("Unable to read",filename,err)
		return
	}
	
	ext := filepath.Ext(filename) 

	switch ext {
	case ".json":
		AssessDeclaration(filename,filebytes)
		
	case ".jsonl":
		
		lines := bytes.Split(filebytes, []byte("\n"))

		for _,runpart := range lines {
			AssessTranscript(filename,runpart)
			fmt.Println("-------------------------------------")
		}

	default:
		fmt.Println("File suffix",ext)
		os.Exit(0)
	}
}

//**************************************************************

func AssessTranscript(filename string,filebytes []byte) {

	var json_file interface{}

	if err := json.Unmarshal(filebytes, &json_file); err != nil {
		return
	}

	var pathList []PathValue

	ExtractPaths("", json_file, &pathList)

	for _, item := range pathList {
		switch item.Path {
		case "text","agent","thinking":
			fmt.Printf("%-10s : \"%v\" (%T)\n",item.Path, item.Value, item.Value)
		}
	}

}

//**************************************************************

func AssessDeclaration(filename string,filebytes []byte) {

	var json_file interface{}

	if err := json.Unmarshal(filebytes, &json_file); err != nil {
		return
	}

	var pathList []PathValue
	var important = []string{"principal","opening","world_event","escalation","think"}

	ExtractPaths("", json_file, &pathList)

	for _, item := range pathList {
		for _,word := range important {
			if strings.Contains(item.Path,word) {
				fmt.Printf("%-10s : \"%v\" (%T)\n", item.Path, item.Value, item.Value)
			}
		}
	}

}

//**************************************************************

func GetFileType(filename string) string {

	prefix := []string{"run","manifest","transcript"}

	last := strings.Split(filename,"/")
	name := last[len(last)-1]
	
	for _,pr := range prefix {
		if strings.Contains(name,pr) {
			return pr
		}
	}
	
	return "none"
}

//**************************************************************


// PathValue represents a single extracted leaf entry.

type PathValue struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

//**************************************************************

func ExtractPaths(prefix string, value interface{}, list *[]PathValue) {

	switch v := value.(type) {
	case map[string]interface{}:
		for key, child := range v {
			newPath := key
			if prefix != "" {
				newPath = prefix + "." + key
			}
			ExtractPaths(newPath, child, list)
		}

	case []interface{}:
		for i, child := range v {
			newPath := fmt.Sprintf("%s.%d", prefix, i)
			ExtractPaths(newPath, child, list)
		}

	default:
		// Reached a leaf node (string, float64, bool, nil)
		*list = append(*list, PathValue{
			Path:  prefix,
			Value: v,
		})
	}
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

