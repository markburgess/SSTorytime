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

	// Start with defs

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

	// Specific run files
	
	namedfiles := []string{"transcript.jsonl","run.json","manifest.json"}
			
	for _,file := range namedfiles {
		cmpt := dir+file
		fmt.Println("\nAssessing data file: ",cmpt)
		AssessJsonType(fp,cmpt)
	}

	// Mention the agents!

	ListAgents(fp)
	
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

func ListAgents(fp *os.File) {

	fmt.Fprintf(fp,"\n\n :: agents :: \n\n")

	for a := range AGENTS {
		fmt.Fprintf(fp,"%s (role) agent in the collective task\n\n",a)
	}
}

//**************************************************************
		
func HandleAgentMarkDown(fp *os.File,dir,name string) {

	const percentage = 100

	psf,L := SST.FractionateMarkdown(dir,name)
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

	if agent_speaking == "" {
		return
	}
	// When each step is collected

	if callAPI {
		fmt.Fprintf(fp," Agent ****`%s` calls tool/API function: %s\n",agent_speaking,callName)
		if callBody != "{}" {
			fmt.Fprintf(fp,"    \"     (message body) 'json body: %s' \n\n",callBody)
		}
	} else {
		// Json might contain Markdown as content!!

		ismd, notes, L := IsMarkdownFragment(thinking,"run history")
		
		if ismd {
			fmt.Fprintf(fp,"\n Agent ****`%s` thinks/announces: status assessment %d\n",agent_speaking,MD_FRAGMENT_COUNTER)
			fmt.Fprintf(fp,"    \"     (message body) agent report %d \n\n",MD_FRAGMENT_COUNTER)
			key := fmt.Sprintf("%d,%d",MD_FRAGMENT_COUNTER,L)
			MARKDOWN_FRAGMENT[key] = notes
			MD_FRAGMENT_COUNTER++
		} else {
			fmt.Fprintf(fp,"\n Agent ****`%s` thinks/announces: '%s' \n",agent_speaking,CleanString(thinking))
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
						ismd, notes,L := IsMarkdownFragment(rhs,"run declaration")
						
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
		fmt.Fprintf(fp,"\n-%s\n",filealias)
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

	last_context := ""
	
	
	for i := range selection {

		if len(selection[i].Fragment) < 1 {
			continue
		}

		context := SST.SpliceSet(ambi_by_part[selection[i].Partition])

		if context != last_context {
			fmt.Fprintf(fp,"\n\n :: %s :: \n\n",context)
		}
		
		var part string
		
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
		//fmt.Fprintf(fp,"              \" (%s) %s\n",SST.INV_CONT_FOUND_IN_S,part)
		
		AddIntentionalContext(collected_fragments,part,anom_by_part[selection[i].Partition],already)
		
		if !partcheck[part] {
			parts = append(parts,part)
			partcheck[part] = true
		}
	}

	fmt.Fprintf(fp,"\n# (end) ************\n")

}

//*******************************************************************

func IsMarkdownFragment(thinking string,name string) (bool,[]SST.TextRank,int) {

	// Cache a lump of analyzed markdown for later, outside of a graph sequence

	if strings.Contains(thinking,"\n") {
		const percentage = 100
		rhsbyte := []byte(thinking)
		psf,L := SST.FractionateSmallMarkdown(rhsbyte,name)
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
		main_name = fmt.Sprintf("%s",main_name)
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

	if !strings.Contains(after,"\n") {
		after = strings.ReplaceAll(after,",",",\n")
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

