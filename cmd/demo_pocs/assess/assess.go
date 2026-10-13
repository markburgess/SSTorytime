//******************************************************************
//
// assess the definitions of an AI agent
//
//******************************************************************

package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"bytes"
	"encoding/json"

	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

//******************************************************************

var MARKDOWN_FRAGMENT = make(map[string][]SST.TextRank)
var MD_FRAGMENT_COUNTER int = 0
var AGENTS = make(map[string]bool)

//******************************************************************

func main() {

	SST.InitSSTorytime()
	
	// read and assess json files
	
	dir := "../../../../../../MIM-fuzzyrobots/AmData/runs/warehouse-kpi-sonnet55-s3-20261004-131451/"

	AssessAgentRun(dir)
}

//******************************************************************

func AssessAgentRun(dir string) {
	
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

	namedfiles := []string{"transcript.jsonl","run.json","manifest.json"}
			
	for _,file := range namedfiles {
		cmpt := dir+file
		fmt.Println("\nAssessing data file: ",cmpt)
		AssessJsonType(fp,cmpt)
	}

	// Markdown at end

		for _, entry := range files {

		ext := filepath.Ext(entry.Name())
		name := entry.Name()
		
		if strings.HasSuffix(ext,".md") {
			_, err := entry.Info()
			if err != nil {
				fmt.Printf(" ** Could not get info for %s: %v\n", name, err)
				continue
			}
			fmt.Println("Assess Markdown descriptive file",dir+name)
			HandleAgentMarkDown(fp,dir,name)
		} 
	}

	// Final markdown fragments, caches during sequence

	for alias := range MARKDOWN_FRAGMENT {
		
		const minN = SST.N_GRAM_MIN
		const maxN = SST.N_GRAM_MAX
		const percentage = 100
		var fragment,L int
		fmt.Sscanf(alias,"%d,%d",&fragment,&L)
		selection := MARKDOWN_FRAGMENT[alias]
		f,s,ff,ss := SST.ExtractIntentionalTokens(L,selection,minN,maxN)
		name := fmt.Sprintf("agent report %d",fragment)
		const no_title = false
		GenerateOutput(fp,name,selection,L,percentage,f,s,ff,ss,no_title)
	}
}

//**************************************************************
		
func HandleAgentMarkDown(fp *os.File,dir,name string) {

	const percentage = 100

	psf,L := SST.FractionateMarkdown(dir+name)
	SST.AnalyzeFractionIntent(psf)
	ranking1 := SST.SelectByRunningIntent(psf,L,percentage)
	ranking2 := SST.SelectByStaticIntent(psf,L,percentage)
	selection := SST.MergeSelections(ranking1,ranking2)
	
	const minN = SST.N_GRAM_MIN
	const maxN = SST.N_GRAM_MAX
	
	f,s,ff,ss := SST.ExtractIntentionalTokens(L,selection,minN,maxN)
	
	name = name[:len(name)-len(".md")]
	alias := MakeTitle(name)

	const with_title = true
	GenerateOutput(fp,alias,selection,L,percentage,f,s,ff,ss,with_title)
}

//**************************************************************

func AssessJsonType(fp *os.File,filename string) {

	filebytes, err := os.ReadFile(filename)
	
	if err != nil {
		fmt.Println("Unable to read",filename,err)
		return
	}
	
	ext := filepath.Ext(filename) 
	alias := MakeTitle(filename)

	switch ext {

	// These files are declarations of intent

	case ".json":
		AssessJsonDeclaration(fp,alias,filebytes)

	// These files are run-transcripts

	case ".jsonl":
		
		lines := bytes.Split(filebytes, []byte("\n"))

		fmt.Fprintf(fp,"\n :: agent transcript, dialogue, log, stream, jsonl, _sequence_ :: \n\n")
		
		for _,runpart := range lines {
			AssessSingleJsonEvent(fp,alias,runpart)
		}

		fmt.Fprintf(fp,"\n -:: _sequence_ :: \n\n")

	// Junk?

	default:
		fmt.Println("File suffix",ext)
		os.Exit(0)
	}
}

//**************************************************************

func AssessSingleJsonEvent(fp *os.File,alias string,filebytes []byte) {

	// transcripts are files of events in sequence, taken
	// one by one and coagulated into a single stream
	
	var json_file interface{}

	if err := json.Unmarshal(filebytes, &json_file); err != nil {
		return
	}

	var pathList []PathValue
	var agent_speaking string
	var callAPI bool
	var json []byte
	var callName string
	var callBody string
	var thinking string
	
	if FindJsonStructureAsPaths("",json_file, &pathList) {

		for _, item := range pathList {

			rhs := item.Value.(string)
			
			switch item.Path {

			case "agent":
				agent_speaking = rhs
				AGENTS[rhs] = true

			case "text","thinking":

				callAPI, callName, json = CheckForAPICall(rhs)
				callBody = string(json)
				
				if !callAPI {
					thinking = rhs
				}
			}
		}
	}

	// When each step is collected

	if callAPI {
		fmt.Fprintf(fp," Agent ***`%s` calls tool/API function: %s\n",agent_speaking,callName)
		if callBody != "{}" {
			fmt.Fprintf(fp,"    \"     (message body) 'json body: %s' \n\n",callBody)
		}
	} else {
		// Json might contain Markdown as content!!

		ismd, notes, L := IsMarkdownFragment(thinking)
		
		if ismd {
			fmt.Fprintf(fp,"\n Agent ***`%s` thinks/announces: status assessment %d\n",agent_speaking,MD_FRAGMENT_COUNTER)
			fmt.Fprintf(fp,"    \"     (message body) agent report %d \n\n",MD_FRAGMENT_COUNTER)
			key := fmt.Sprintf("%d,%d",MD_FRAGMENT_COUNTER,L)
			MARKDOWN_FRAGMENT[key] = notes
			MD_FRAGMENT_COUNTER++
		} else {
			fmt.Fprintf(fp,"\n Agent ***`%s` thinks/announces: '%s' \n",agent_speaking,CleanString(thinking))
		}
	}
}

//**************************************************************

func AssessJsonDeclaration(fp *os.File,alias string,filebytes []byte) {

	var json_file interface{}

	if err := json.Unmarshal(filebytes, &json_file); err != nil {
		return
	}
	
	var pathList []PathValue
	var important = []string{"principal","opening","escalation","reason"}

	if FindJsonStructureAsPaths("", json_file, &pathList) {

		for _, item := range pathList {
			for _,word := range important {
				if strings.Contains(item.Path,word) {

					agent, action := InferPathMeaning(item.Path)
					lhs := fmt.Sprintf("event where %s declares ",agent,action)
					rhs := item.Value.(string)
					rhsbyte := []byte(rhs)
					
					// What to do with the RHS for linking etc
					
					if json.Valid(rhsbyte) {
						fmt.Fprintf(fp,"\n Declaration %d: %s\n",lhs,MD_FRAGMENT_COUNTER)
						fmt.Fprintf(fp,"    \"     (message body) '%s' \n\n",string(rhsbyte))	
					} else {
						ismd, notes,L := IsMarkdownFragment(rhs)
						
						if ismd {
							fmt.Fprintf(fp,"\n Declaration %d\n",MD_FRAGMENT_COUNTER)
							fmt.Fprintf(fp,"    \"     (message body) agent report %d \n\n",MD_FRAGMENT_COUNTER)
							key := fmt.Sprintf("%d,%d",MD_FRAGMENT_COUNTER,L)
							MARKDOWN_FRAGMENT[key] = notes
							MD_FRAGMENT_COUNTER++
						} else {
							fmt.Fprintf(fp,"\n DEclartion '%s' \n",CleanString(rhs))
						}
					}
				}
			}
		}
	}

}

//*******************************************************************

func GenerateOutput(fp *os.File,filealias string,selection []SST.TextRank,L int, percentage float64,anom_by_part[][]string,ambi_by_part[][]string,all_anom[]string,all_ambi[]string,with_title bool) {

	// See AddMandatory() in N4L.go for reserved names (TBD, collect these one day as const)

	var collected_fragments = make(map[string][]string)

	if with_title {
		fmt.Fprintf(fp,"-Agent specification: %s\n",filealias)
	}
	
	fmt.Fprintf(fp,"\n# (begin) ************ %s\n",filealias)

	// Lookup tables of contents from Markdown

	SST.CompileTabular(fp,filealias,SST.TABLES)
	SST.CompileTOC(fp,filealias,SST.SECTIONS)
	SST.CompileProperNames(fp,filealias,selection,SST.TABLES)

	// 

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

		fmt.Fprintf(fp,"\n\n :: %s :: \n\n",context)
		
		if len(selection[i].Title) > 0 {
			part = selection[i].Title
		} else {
			part = "No section"
		}
		
		// Add context from n = 2,3 fractions
		
		if part != lastpart {
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
}

//*******************************************************************

func IsMarkdownFragment(thinking string) (bool,[]SST.TextRank,int) {

	// Cache a lump of analyzed markdown for later, outside of a graph sequence

	if strings.Contains(thinking,"\n") {
		const percentage = 100
		rhsbyte := []byte(thinking)
		psf,L := SST.FractionateSmallMarkdown(rhsbyte)
		SST.AnalyzeFractionIntent(psf)
		ranking1 := SST.SelectByRunningIntent(psf,L,percentage)
		ranking2 := SST.SelectByStaticIntent(psf,L,percentage)
		selection := SST.MergeSelections(ranking1,ranking2)
		return true,selection,L
	} else {
		var dummy []SST.TextRank
		return false,dummy,0
	}
}

//*******************************************************************

func MakeTitle(filename string) string {

	var main_name string
	
	if strings.Contains(filename,".") {
		subcat := strings.Split(filename,".")
		main_name = subcat[len(subcat)-1]
		main_name = strings.ReplaceAll(main_name,"_"," ")
		main_name = fmt.Sprintf("%s [%s]",main_name,subcat[0])
	} else {
		main_name = strings.ReplaceAll(filename,"_"," ")
	}

	return main_name
}

//*******************************************************************

func InferPathMeaning(path string) (string,string) {

	var agent, answer string
	parts := strings.Split(path,".")
	
	// Do we recognize the agent?
	
	for _,p := range parts {
		if AGENTS[p] {
			agent = p
		}

		if strings.Contains(p,"principal") || strings.Contains(p,"client") {
			answer = p
		}

		if strings.Contains(p,"reason") {
			answer = p
		}
	}

	return agent,answer
}

//*******************************************************************

func CleanString(s string) string {

	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s,"(","[")
	s = strings.ReplaceAll(s,")","]")
	s = strings.ReplaceAll(s,"- ","* ")
	return s
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

func CheckForAPICall(str string) (bool,string,[]byte) {

	// Look for "function_name({json...})"
	
	if str[len(str)-1] != ')' {
		return false,"",[]byte{}
	}

	before, after, found := strings.Cut(str, "(")
	
	if !found {
		return false,"",[]byte{}
	}

	const function_name_size = 60

	if len(before) > function_name_size {
		return false,"",[]byte{}
	}

	look_for_json := []byte(after[:len(after)-1])
	
	// Ok, looks like it has the correct shape

	if json.Valid(look_for_json) {
		return true,before,look_for_json
	}

	return true,before,look_for_json
}

//**************************************************************
// Json parsing
//**************************************************************

type PathValue struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

//**************************************************************

func FindJsonStructureAsPaths(prefix string, value interface{}, list *[]PathValue) bool {

	var result bool
	
	switch v := value.(type) {

	case map[string]interface{}:

		for key, child := range v {
			newPath := key
			if prefix != "" {
				newPath = prefix + "." + key
			}
			result = FindJsonStructureAsPaths(newPath, child, list)
		}

	case []interface{}:

		for i, child := range v {
			newPath := fmt.Sprintf("%s.%d", prefix, i)
			result = FindJsonStructureAsPaths(newPath, child, list)
		}

	default: // Reached a leaf node (string, float64, bool, nil)

		if result, ok := v.(string); ok {
			*list = append(*list, PathValue{
				Path:  prefix,
				Value: result,
			})

			return true
		}
	}

	return result
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

