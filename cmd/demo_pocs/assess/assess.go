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

var AGENTS = make(map[string]bool)
var AGENT_OPENING = make(map[string]string)
var AGENT_TOOLS = make(map[string]bool)

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

	// Specific run files, run last so we can pick up intel first
	
	namedfiles := []string{"transcript.jsonl","manifest.json","run.json"}
			
	for _,file := range namedfiles {
		cmpt := dir+file
		fmt.Println("\nAssessing data file: ",cmpt)
		AssessJsonType(fp,cmpt)
	}

	// Mention the agents!

	ListAgents(fp)
	ListTools(fp,AGENT_TOOLS)
	SST.ListProperNames(fp)
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

	fmt.Println(" - fractionating",name)
	psf,L,notes := SST.FractionateMarkdown(dir,name)
	fmt.Println(" - scanning")
	SST.AnalyzeFractionIntent(psf)
	fmt.Println(" - ranking longitudinally")
	ranking1 := SST.SelectByRunningIntent(psf,L,percentage)
	fmt.Println(" - ranking transversely")
	ranking2 := SST.SelectByStaticIntent(psf,L,percentage)
	selection := SST.MergeSelections(ranking1,ranking2)
	
	const minN = SST.N_GRAM_MIN
	const maxN = SST.N_GRAM_MAX
	
	f,s,ff,ss := SST.ExtractIntentionalTokens(L,selection,minN,maxN)

	alias,_ := SST.MakeTitleFromFileName(name)

	const with_title = true
	fmt.Println(" - writing results for",alias)
	GenerateOutput(fp,alias,selection,L,percentage,f,s,ff,ss,with_title)
	AddNotes(fp,notes)
}

//**************************************************************

func AssessJsonType(fp *os.File,filename string) {

	filebytes, err := os.ReadFile(filename)
	
	if err != nil {
		fmt.Println("Unable to read",filename,err)
		return
	}

	alias, ext := SST.MakeTitleFromFileName(filename)

	switch ext {

	// These files are declarations of intent

	case ".json":
		AssessJsonDeclaration(fp,alias,filebytes)

	// These files are run-transcripts

	case ".jsonl":
		
		lines := bytes.Split(filebytes, []byte("\n"))

		fmt.Fprintf(fp,"\n :: _sequence_, agent transcript, agent trace :: \n\n")

		fmt.Fprintf(fp,"\nExecution starts\n")

		// Look for initial statements
		
		for agent := range AGENT_OPENING {
			if AGENT_OPENING[agent] != "" {
				fmt.Fprintf(fp,"\nAgent [%s] begins (message body) '%s'\n",AGENT_OPENING[agent])
			}
		}

		// Now unfold the log
		
		for n,runpart := range lines {
			AssessSingleJsonEvent(fp,n,alias,runpart)
		}

		fmt.Fprintf(fp,"\n -:: _sequence_ :: \n\n")

	// Junk?

	default:
		fmt.Println("File suffix",ext)
		os.Exit(0)
	}
}

//**************************************************************

func AssessSingleJsonEvent(fp *os.File,n int,alias string,filebytes []byte) {

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
	var call_name string
	var call_body string
	var thinking string
	
	if FindJsonStructureAsPaths("",json_file, &pathList) {

		for _, item := range pathList {

			rhs := item.Value.(string)
			
			switch item.Path {

			case "agent":
				agent_speaking = rhs
				AGENTS[rhs] = true

			case "text","thinking":

				callAPI, call_name, json = CheckForAPICall(rhs)
				call_body = string(json)
				AGENT_TOOLS[call_name] = true
			}
		}
	}

	if agent_speaking == "" {
		agent_speaking = "unknown agent"
	}
	
	if callAPI {
		fmt.Fprintf(fp,"agent event %d [%s] call to tool/API function: ****`%s`[..]\n",n,agent_speaking,call_name)
		if call_body != "{}" {
			fmt.Fprintf(fp,"  \"  (message body) 'json body: %s' \n",call_body)
		}
	} else {
		// Json might contain Markdown as content!!

		ismd,selections,_,L := IsMarkdownFragment(thinking,"run history")
		
		const exposition_threshold = 100
		
		if L > exposition_threshold {
			fmt.Printf(" ** agent %s makes a noteworthy action/assertion!\n",agent_speaking)
			fmt.Fprintf(fp,"   \" (note)  agent %s makes an extensive action/assertion!\n",agent_speaking)
		}
		
		if ismd {
			fmt.Fprintf(fp,"\n agent %s thinks/announces: status assessment\n",agent_speaking)
			const minN = SST.N_GRAM_MIN
			const maxN = SST.N_GRAM_MAX
			const percentage = 100
			f,_,_,_ := SST.ExtractIntentionalTokens(L,selections,minN,maxN)

			for _,sel := range f {
				for _,s := range sel {
					if s != "" {
						fmt.Fprintf(fp,"  \" (action) %s\n",s)
						ScanForDanger(s)
					}
				}
			}

		} else {
			cln := CleanString(thinking)
			if cln != "" {
				fmt.Fprintf(fp,"     \"  (action) '%s' \n",cln)
			}
		}
	}
}

//**************************************************************

func AssessJsonDeclaration(fp *os.File,alias string,filebytes []byte) {

	// Assess a json declaration file telling us aboout an agent deployment
	
	var json_file interface{}

	if err := json.Unmarshal(filebytes, &json_file); err != nil {
		return
	}

	var summary []string
	var pathList []PathValue

	if FindJsonStructureAsPaths("", json_file, &pathList) {

		// parse the items in the runlog
		
		for _, item := range pathList {

			agent, action := InferPathMeaning(item.Path)

			if agent == "" {
				continue
			}

			lhs := fmt.Sprintf("event where [%s] declares %s",agent,action)
			rhs := item.Value.(string)
			rhsbyte := []byte(rhs)
			
			// Handle different types
			
			if strings.Contains(action,"opening") {
				AGENT_OPENING[agent] = rhs
			}
	
			if json.Valid(rhsbyte) {
				if len(lhs) > 0 {
					fmt.Fprintf(fp,"%s (message body) '%s' \n\n",lhs,string(rhsbyte))
				}
			} else {
				ismd,selections,notes,L := IsMarkdownFragment(rhs,"run declaration")
				summary = notes
				
				if ismd {
					//markdown already scanned
					const minN = SST.N_GRAM_MIN
					const maxN = SST.N_GRAM_MAX
					fmt.Println(" - writing embedded markdown fragment")
					f,s,_,_ := SST.ExtractIntentionalTokens(L,selections,minN,maxN)
					//GenerateOutput(fp,alias,selection,L,percentage,f,s,ff,ss,with_title)
					fmt.Println("\n*******************\n")
					for _,st := range f {
						fmt.Println(" [[[ %s",st)
					}
					for _,st := range s {
						fmt.Println(" <<< %s",st)
					}
					fmt.Println("\n*********************\n")
					fmt.Fprintf(fp,"\n%s (message body) '%s' \n",agent,selections)
				} else {
					// plain text
					fmt.Fprintf(fp,"\n%s (message body) '%s' \n",agent,CleanString(rhs))
				}
			}
		}
	}
	
	AddNotes(fp,summary)
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
			fmt.Fprintf(fp,"\n\n :: %s [%s] :: \n\n",context,filealias)
		}
		
		var part string
		
		if part != lastpart {
			if len(context) > 0 {
				fmt.Fprintf(fp,"\n-:: _sequence_ ::")
				fmt.Fprintf(fp,"\n#######################################")
				fmt.Fprintf(fp,"\n :: %s, %s, _sequence_ ::\n",context,part)
				fmt.Fprintf(fp,"#######################################\n")
				fmt.Fprintf(fp,"\n %s\n\n",part)		
			}
			
			lastpart = part
			
		}

		fmt.Fprintf(fp,"\n %s [%s]\n\n",SST.SanitizeParen(selection[i].Fragment),filealias)
		
		AddIntentionalContext(collected_fragments,part,anom_by_part[selection[i].Partition],already)
		
		if !partcheck[part] {
			parts = append(parts,part)
			partcheck[part] = true
		}
	}

	fmt.Fprintf(fp,"\n# (end) ************\n")
}

//*******************************************************************

func AddNotes(fp *os.File,notes []string) {

	// Drop the out of sequence annotations at the end

	if notes != nil {
		fmt.Fprintf(fp,"  :: actions, unstructured notes, agent interactions  :: \n\n")

		for _,action := range notes {
			fmt.Fprintf(fp," %s\n",action)
		}
	}
}

//*******************************************************************

func ListTools(fp *os.File,tools map[string]bool) {

	fmt.Fprintf(fp,"\n :: tools :: \n\n")

	for tool := range tools {

		tool = strings.TrimSpace(tool)
		if tool != "" {
			fmt.Fprintf(fp,"\n %s (role) agent API/tool \n",tool)
		}
	}

}

//*******************************************************************

func IsMarkdownFragment(thinking string,name string) (bool,[]SST.TextRank,[]string,int) {

	// Cache a lump of analyzed markdown for later, outside of a graph sequence

	if strings.Contains(thinking,"\n") {
		const percentage = 100
		rhsbyte := []byte(thinking)
		psf,L,notes := SST.FractionateSmallMarkdown(rhsbyte,name)
		SST.AnalyzeFractionIntent(psf)
		ranking1 := SST.SelectByRunningIntent(psf,L,percentage)
		ranking2 := SST.SelectByStaticIntent(psf,L,percentage)
		selection := SST.MergeSelections(ranking1,ranking2)
		return true,selection,notes,L
	} else {
		var dummy []SST.TextRank
		return false,dummy,nil,0
	}
}

//*******************************************************************

func InferPathMeaning(path string) (string,string) {

	var agent, answer string
	parts := strings.Split(path,".")

	recognize := []string{"principal","opening","client","reason","escalation","reason"}	
	
	for _,p := range parts {

		// Do we recognize the agent?

		if AGENTS[p] {
			agent = p
		}
		
		for _,v := range recognize {
			
			if strings.Contains(p,v) {
				answer = p
				continue
			}
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
// Scanning for danger
//**************************************************************

func ScanForDanger(s string) {

	//fmt.Println(" . . . scan:",s)
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

