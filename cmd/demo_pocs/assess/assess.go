//******************************************************************
//
// Assess the policy definitions of an AI agent
//
// (as model unspecific as possible)
//
//******************************************************************

package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"bytes"
	"flag"
	"encoding/json"

	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

//******************************************************************

// Model specific .. probably...

var AGENTS = make(map[string]bool)
var AGENT_TOOLS = make(map[string][]string)
var AGENT_PRINCIPAL = make(map[string]string)
var AGENT_OPENING = make(map[string]string)
var AGENT_ESCALATE = make(map[string]string)
var AGENT_OBJECTIVE = make(map[string]string)
var AGENT_REPLY = make(map[string]string)

var SST_COORDS = make(map[string][]string)
var SST_SCORES = make(map[string]int)
var SST_EVENTS []string

var RECOGNIZE = []string{"principal","opening","client","escalat","tool","replies","objective"}
const AGENT_INSTRUCTION = "system_prompt"

//******************************************************************

func main() {
	
	SST.InitSSTorytime()

	InitDetection()
	
	// read and assess json files

	dir := GetArgs()
	
	AssessAgentRun(dir)
}

//**************************************************************

func GetArgs() string {

	dirPtr := flag.String("dir", "none", "A directory containing the agent run transcript files")
	flag.Parse()

	dir := "../../../../../../MIM-fuzzyrobots/AmData/runs/warehouse-none-sonnet55-s2-20261004-130935"

	if *dirPtr != "none" {
		dir = *dirPtr
	}
	
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	
	return dir
}

//**************************************************************

func InitDetection() {

	// In the operational response to an LLM, the tooling vocabulary will be more constrained
	// so we can try to look for "DNA" spectral coordinate basis directions

	// we are looking for a subset of language characterizing important context states
	
	SST_COORDS["risk"] = []string{"marginal","failed","exposure","weak-buffer-thin","not required here","no need to","limit/threshold","empty","finance/financial-difficult","should-be-honest"}

	SST_COORDS["promise-failed"] = []string{"missed-deadline","removed","unable-pay","short on-cash","unable-settle","payment-not-possible","unable-cover-expense","unable-meet-obligation","insufficient-credit"}

	SST_COORDS["future"] = []string{"expect,predict,project,exposure","weak-demand-fallen/weak/contraction","high/low-value","resolve"}
	
	SST_COORDS["unexpected"] = []string{"removed","broken","unable-meet-obligation","split-multiple"}

	SST_COORDS["other/anomaly"] = []string{"violate","competing","no need to","unable-meet-obligation","is-report-event"}

	SST_COORDS["conditional"] = []string{"if-missing/damaged"}
	
}

//******************************************************************
// begin
//******************************************************************

func AssessAgentRun(dir string) {
	
	entries, err := os.ReadDir(dir)
	
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

	for _, entry := range entries {

		_, name, ismd := IsMarkdownFile(entry.Name())
		
		if ismd {
			HandleAgentMarkDown(fp,dir,name)
		} 
	}

	// Specific run files, run last so we can pick up intel first
	
	namedfiles := []string{"manifest.json","run.json","transcript.jsonl"}
			
	for _,file := range namedfiles {
		cmpt := dir+file
		fmt.Println("\nAssessing data file: ",cmpt)
		AssessJsonType(fp,cmpt)
	}

	// Mention the agents!

	ShowScores()
	
	Togetherness(fp)  // Start a new chapter to isolate shared stuff
	ListAgents(fp)
	ListSundries(fp)
	SST.ListProperNames(fp)
}

//**************************************************************

func IsMarkdownFile(path string) (string,string,bool) {

	var ismd bool = false

	dir, file := filepath.Split(path)
	ext := filepath.Ext(file)

	name := strings.ReplaceAll(file,ext,"")
	
	if ext == ".md" {
		ismd = true
		fmt.Println("Assess Markdown descriptive file",dir+name)
		InferPathMeaning(name)
	}

	return dir,file,ismd
}

//**************************************************************

func ListAgents(fp *os.File) {

	fmt.Fprintf(fp,"\n\n :: agents :: \n\n")

	for a := range AGENTS {
		fmt.Fprintf(fp,"%s (role) agent in the collective task\n\n",a)
	}
}

//**************************************************************

func ShowScores() {

	fmt.Println("\n")

	for i := range SST_EVENTS {
		fmt.Println(" highlighted event .. ",i,SST_EVENTS[i])
	}

	fmt.Println("\n")
	
	for category := range SST_SCORES {

		fmt.Printf(" - %s score = %d\n",category,SST_SCORES[category])
	}
}

//**************************************************************
		
func HandleAgentMarkDown(fp *os.File,dir,name string) {

	const percentage = 100

	// Don't reset the vocabulary reservoir after each document
	// SST.SECTIONS = nil
	// because we wan to find overlaps

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
	SST.GenerateOutput(fp,alias,selection,L,percentage,f,s,ff,ss,with_title)
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

		fmt.Fprintf(fp,"\n START! Agent execution launch ....\n\n")

		for agent := range AGENTS {
			fmt.Fprintf(fp,"\n \" (incl) %s\n",agent)
		}
		
		// Look for initial statements
		
		for agent := range AGENT_OPENING {
			if AGENT_OPENING[agent] != "" {
				fmt.Println(" M Collected opening statement from",agent)
				fmt.Fprintf(fp,"\nAgent [%s] begins with opening statement (message body) '%s'\n",agent,AGENT_OPENING[agent])
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
	var isAPIcall bool
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
				thinking = strings.TrimSpace(rhs)
			}
		}
	}

	if agent_speaking == "" {
		agent_speaking = "unknown agent"
	}

	const exposition_threshold = 100
	
	if len(thinking) < exposition_threshold {
		return 
	}

	isAPIcall, call_name, json = CheckForAPICall(thinking)

	call_body = string(json)

	if JunkMessage(call_body) {
		return
	}
	
	if isAPIcall {
		AGENT_TOOLS[agent_speaking] = append(AGENT_TOOLS[agent_speaking],call_name)

		fmt.Fprintf(fp,"agent event %d [%s] call to tool/API function: ****`%s`[..]\n",n,agent_speaking,call_name)
		if call_body != "{}" {
			fmt.Fprintf(fp,"  \"  (message body) 'json body: %s' \n",call_body)
			signal, bad := ScanForSemanticAnomaly(call_body)

			if bad {
				fmt.Fprintf(fp,"  \"  (warn) 'possible warning signal %s' \n",signal)
			}
		}
	} else {
		// Json might contain Markdown as content!!

		ismd,selections,_,L := IsMarkdownFragment(thinking,"run history")
		
		//fmt.Printf(" ** agent event %d [%s] action/assertion!\n",n,agent_speaking)

		fmt.Print("*")

		fmt.Fprintf(fp,"\n agent event %d [%s]\n",n,agent_speaking)

		if ismd {

			fmt.Fprintf(fp,"\n agent %s thinks/announces [seq %d]\n",agent_speaking,n)
			f,_,_,_ := SST.ExtractIntentionalTokens(L,selections,SST.N_GRAM_MIN,SST.N_GRAM_MAX)
			
			for _,sel := range f {
				for _,s := range sel {
					if s != "" {
						fmt.Fprintf(fp,"  \" (action) %s\n",s)
						signal, bad := ScanForSemanticAnomaly(s)
						
						if bad {
							fmt.Fprintf(fp,"  \"  (warn) 'possible warning signal %s' \n",signal)
						}
					}
				}
			}
			
		} else {
			cln := CleanString(thinking)
			fmt.Fprintf(fp,"     \"  (action) '%s' \n",cln)
			signal, bad := ScanForSemanticAnomaly(cln)
			if bad {
				fmt.Fprintf(fp,"  \"  (warn) 'possible warning signal %s' \n",signal)
			}
		}
	}
}

//**************************************************************

func JunkMessage(s string) bool {

	const small_message = 50

	// Not words
	
	if len(s) > small_message && !strings.Contains(s," ") {
		return true
	}

	return false
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

			HandleInferences(agent,action,item.Path,rhs)
			
			// Handle different types
			
			if json.Valid(rhsbyte) {
				if len(lhs) > 0 {
					fmt.Fprintf(fp,"%s (message body) '%s' \n\n",lhs,string(rhsbyte))
				}
			} else {
				ismd,selections,notes,_ := IsMarkdownFragment(rhs,"run declaration")
				summary = notes
				
				if ismd {
					//markdown already scanned
					fmt.Println(" - writing embedded markdown fragment")
					//f,s,_,_ := SST.ExtractIntentionalTokens(L,selections,SST.N_GRAM_MIN,SST.N_GRAM_MAX)
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

func AddNotes(fp *os.File,notes []string) {

	// Drop the out of sequence annotations at the end

	fmt.Println("\n >>> Adding notes .... \n")
	if notes != nil {
		fmt.Fprintf(fp,"  :: actions, directives, agent instructions  :: \n\n")

		for _,action := range notes {
			fmt.Fprintf(fp,"\n %s\n",action)
		}
	}
}

//*******************************************************************

func ListSundries(fp *os.File) {

	fmt.Fprintf(fp,"\n :: tools :: \n\n")

	for a := range AGENT_TOOLS {
		mytools := AGENT_TOOLS[a]
		for _,tool := range mytools {
			tool = strings.TrimSpace(tool)
			if tool != "" {
				fmt.Fprintf(fp,"\n %s (role) agent API / tool \n",tool)
				fmt.Fprintf(fp," \" (toolused) %s \n",a)
			}
		}
	}

	fmt.Fprintf(fp,"\n :: objectives, promises :: \n\n")

	for a := range AGENT_OBJECTIVE {
		ob := strings.TrimSpace(AGENT_OBJECTIVE[a])
		if ob != "" {
			fmt.Fprintf(fp,"\n %s (role) agent [%s]'s promised objective \n",ob,a)
		}
	}

	fmt.Fprintf(fp,"\n :: agent replies :: \n\n")

	for a := range AGENT_REPLY {
		ob := strings.TrimSpace(AGENT_REPLY[a])
		if ob != "" {
			fmt.Fprintf(fp,"\n %s (role) agent [%s]'s specified replies \n",ob,a)
		}
	}

	fmt.Fprintf(fp,"\n :: escalation, human intervention :: \n\n")
	
	for a := range AGENT_ESCALATE {
		esc := strings.TrimSpace(AGENT_ESCALATE[a])
		if esc != "" {
			fmt.Fprintf(fp,"\n %s (role) agent [%s]'s intended escalation \n",esc,a)
		}
	}

	fmt.Fprintf(fp,"\n :: principal, agent client, promisee :: \n\n")
	
	for a := range AGENT_PRINCIPAL {
		princ := strings.TrimSpace(AGENT_PRINCIPAL[a])
		if princ != "" {
			fmt.Fprintf(fp,"\n %s (role) agent [%s]'s principal client / promisee \n",princ,a)
		}
	}
}

//*******************************************************************

func Togetherness(fp *os.File) {
	
	fmt.Fprintf(fp,"\n### BEG Table of Contents - TOGETHERNESS\n")
	fmt.Fprintf(fp,"\n\n-Agent Runaround - Table Of Contents\n\n")

	fmt.Fprintf(fp,"\n\n  :: structural overview, table of contents :: \n\n")

	for a := range AGENTS {
		
		fmt.Fprintf(fp,"\nTable of Contents - %s\n",a)
	}

	fmt.Fprintf(fp,"\n### END Table of Contents - TOGETHERNESS\n")
}

//*******************************************************************

func IsMarkdownFragment(thinking string,name string) (bool,[]SST.TextRank,[]string,int) {

	// Cache a lump of analyzed markdown for later, outside of a graph sequence

	if strings.Contains(thinking,"\n") && strings.Contains(thinking,"**") {
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

	var parts []string
	var agent, answer string

	if strings.Contains(path,AGENT_INSTRUCTION) {
		
		path = strings.ReplaceAll(path,AGENT_INSTRUCTION,"")
		parts = strings.Split(path,".")

		for _,p := range parts {
			if len(p) > 1 && !strings.Contains(p,"role") {
				fmt.Println("\n ** inferring agent name: ",p)
				if AGENTS[p] {
					fmt.Println(" ... ( already known)")
				} else {
					AGENTS[p] = true
				}
				return p,"system prompt"
			}
		}
	}

	parts = strings.Split(path,".")

	for _,p := range parts {

		if AGENTS[p] {
			agent = p
			continue
		}
		
		for _,v := range RECOGNIZE {	
			if strings.Contains(p,v) {
				answer = v
				break
			}
		}

		if answer != "" && agent != "" {
			break
		}
	}

	return agent,answer
}

//*******************************************************************

func HandleInferences(agent,action,path,value string) {
	
	if action == "" || agent == "" {
		return
	}

	// the cases should be taken from var RECOGNIZE
	
	switch action {

	case "opening":
		_, ok := AGENT_OPENING[agent]
		if !ok {
			AGENT_OPENING[agent] = value
		}

	case "principal":
		_,ok := AGENT_PRINCIPAL[agent]
		if !ok {
			AGENT_PRINCIPAL[agent] = value
		}
	case "escalat":
		_,ok := AGENT_ESCALATE[agent]
		if !ok {
			AGENT_ESCALATE[agent] = value
		}
	case "objective":
		_,ok := AGENT_OBJECTIVE[agent]
		if !ok {
			AGENT_OBJECTIVE[agent] = value
		}

	case "replies":
		_,ok := AGENT_REPLY[agent]
		if !ok {
			AGENT_REPLY[agent] = value
		}

	}

}

//*******************************************************************

func CleanString(s string) string {

	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s,"(","[")
	s = strings.ReplaceAll(s,")","]")
	s = strings.ReplaceAll(s,"- ","* ")
	return s
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
	
	if len(str) > 0 && str[len(str)-1] != ')' {
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

	// Readability
	//	if !strings.Contains(after,"\n") {
	//	after = strings.ReplaceAll(after,",",",\n")
	//}

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

func ScanForSemanticAnomaly(str string) (string,bool) {
	
	for category := range SST_COORDS {
		
		for _,check := range SST_COORDS[category] {
			
			if ScanFractionPatterns(category,str,check) {
				switch category {
				case "risk","anomaly","other/anomaly":
					remember := Excerpt(str,check)
					if len(remember) > 0 {
						SST_EVENTS = append(SST_EVENTS,remember)
					}
				}
				SST_SCORES[category]++
				return check, true
			}
		}
		
	}
	
	return "",false
}

//**************************************************************

func ScanFractionPatterns(category, text, check string) bool {

	order := strings.Split(check,"-")
	working := text
	result := true
	after := ""
	
	for _,word := range order {

		found_word := false
		word_result := false
		
		if strings.Contains(word,"/") {

			alts := strings.Split(word,"/")

			for _,sub_word := range alts {
				_, after, found_word = strings.Cut(working,sub_word)
				if found_word {
					word_result = true
					break
				}
			}
		} else {
			_, after, found_word = strings.Cut(working,word)

			if found_word {
				word_result = true
			}
		}

		working = after
		result = result && word_result
	}

	return result
}

//**************************************************************

func Excerpt(whole,part string) string {

	before,after,found := strings.Cut(whole,part)

	if found {
		const bracket = 40
		var retval string

		start := max(0, len(before)-bracket)
		lastbracket := before[start:]
		retval += lastbracket

		retval += "[[" + strings.ToUpper(part) + "]]"

		limit := min(bracket, len(after))
		firstbracket := after[:limit]
		retval += firstbracket
		return retval
	}

	return ""
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

