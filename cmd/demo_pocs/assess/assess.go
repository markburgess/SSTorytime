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
// This model is specific to this example...but have to start somewhere
//******************************************************************

type P2Pair struct {
	From        string // agent sender
	To          string // agent receiver
	ToolStart   string // the tool call that opens the channel
	OfferEvent  string // the sender event
	AcceptEvent string // the receiver event
}

// Spacetime structure

var AGENTS = make(map[string]bool)
var AGENT_TIMELINE = make(map[string][]string)

// Feature extraction

var AGENT_REQUEST = make(map[string][]string)
var AGENT_PRINCIPAL = make(map[string]string)
var AGENT_OPENING = make(map[string]string)
var AGENT_ESCALATE = make(map[string]string)
var AGENT_REPLY []string

var AGENT_TOOL_CALLS = make(map[string][]string)
var LAST_TOOL_CALL = make(map[P2Pair]string)
var LAST_ACCEPT  = make(map[P2Pair]string)
var P2P_MESSAGE = make(map[string]P2Pair)

var SST_COORDS = make(map[string][]string)
var SST_SCORES = make(map[string]int)
var SST_ANOMALIES []string
var SST_DUPLICATES []string

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
	
	SST_COORDS["risk"] = []string{"marginal","failed","exposure","weak-buffer-thin","not required here","no need-involve","empty","finance/financial-difficult","should-be-honest","beyond/passed/exceeded-threshold"}

	SST_COORDS["promise-failed"] = []string{"missed-deadline","removed","unable-pay","short on-cash","unable-settle","payment-not-possible","unable-cover-expense","unable-meet-obligation","insufficient-credit"}

	SST_COORDS["future"] = []string{"expect,predict,project,exposure","weak-demand-fallen/weak/contraction","high/low-value","resolve","still--before/until-deadline","still-time"}
	
	SST_COORDS["unexpected"] = []string{"removed","broken","unable-meet-obligation","split-multiple","tampering","miscount","misconduct"}

	SST_COORDS["other/anomaly"] = []string{"violate","competing","no need to","unable-meet-obligation","is-report-event","nonconform","asking-but"}

	SST_COORDS["conditional"] = []string{"if-missing/damaged"}

	SST_COORDS["imposition_request"] = []string{"approve-authoriz"}
	
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

		_, name, isft := IsFileType(entry.Name())
		
		if isft == ".md" {
			HandleAgentMarkDown(fp,dir,name)
		} 

		if isft == ".txt" {
			HandleAgentText(fp,dir,name)
		} 

	}

	// Specific run files, run last so we can pick up intel first
	
	namedfiles := []string{"manifest.json","run.json","transcript.jsonl"}
			
	for _,file := range namedfiles {
		cmpt := dir+file
		fmt.Println("Assessing data file: ",file)
		AssessJsonType(fp,cmpt)
	}

	// Mention the agents!

	ShowScores()
	TimeLines(fp)
	Togetherness(fp)  // Start a new chapter to isolate shared stuff
	ListAgents(fp)
	ListSundries(fp)
	SST.ListProperNames(fp)
}

//**************************************************************

func IsFileType(path string) (string,string,string) {

	dir, file := filepath.Split(path)
	ext := filepath.Ext(file)

	name := strings.ReplaceAll(file,ext,"")
	
	if ext == ".md" {
		fmt.Println("Assess Markdown descriptive file",dir+name)
		InferPathMeaning(name)
	}

	if ext == ".txt" {
		fmt.Println("Assess Plain Text descriptive file",dir+name)
		InferPathMeaning(name)
	}

	return dir,file,ext
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

	for i := range SST_ANOMALIES {
		fmt.Println("\n Highlighted event .. ",i+1,SST_ANOMALIES[i])
	}

	fmt.Println("\n")
	
	for category := range SST_SCORES {

		if SST_SCORES[category] > 0 {
			fmt.Printf(" - %s score = %d\n",category,SST_SCORES[category])
		}
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
		
func HandleAgentText(fp *os.File,dir,name string) {

	const percentage = 100

	// Don't reset the vocabulary reservoir after each document
	// SST.SECTIONS = nil
	// because we wan to find overlaps

	fmt.Println(" - fractionating",name)
	psf,L := SST.FractionateTextFile(dir,name)

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

		// Now analyze the log
		var previous_line []byte
		
		for n,this_line := range lines {
			AssessSingleJsonEvent(n,alias,this_line,previous_line)
			previous_line = this_line
		}

	// Junk?

	default:
		fmt.Println("File suffix",ext)
		os.Exit(0)
	}
}

//**************************************************************
// Declarative json
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

			if JunkMessage("",rhs) {
				return
			}

			HandleInferences(agent,action,item.Path,rhs)
			
			// Handle different types
			
			if json.Valid(rhsbyte) {
				if len(lhs) > 0 {
					fmt.Fprintf(fp,"%s (message body) 'json:\n%s' \n\n",lhs,CleanString(rhs))
				}
			} else {
				ismd,selections,notes,_ := IsMarkdownFragment(rhs,"run declaration")
				summary = notes
				
				if ismd {
					//markdown already scanned
					fmt.Println(" - writing embedded markdown fragment")
					//f,s,_,_ := SST.ExtractIntentionalTokens(L,selections,SST.N_GRAM_MIN,SST.N_GRAM_MAX)
					fmt.Fprintf(fp,"\n%s (message body) %s \n",agent,selections)
				} else {
					// plain text
					fmt.Fprintf(fp,"\n%s (message body) %s \n",agent,CleanString(rhs))
				}
			}
		}
	}
	
	AddNotes(fp,summary)
}

//*******************************************************************

func AddNotes(fp *os.File,notes []string) {

	// Drop the out of sequence annotations at the end

	fmt.Println("Adding supplementary notes .... \n")

	if notes != nil {
		fmt.Fprintf(fp,"  :: actions, directives, agent instructions  :: \n\n")

		for _,action := range notes {
			fmt.Fprintf(fp,"\n %s\n",action)
		}
	}
}

//**************************************************************
// Timeline transcript decomp/analysis
//**************************************************************

func AssessSingleJsonEvent(n int,alias string,log_event,previous_event []byte) {

	// Revised version knowns a lot more about the json format than before

	var json_file interface{}

	if err := json.Unmarshal(log_event, &json_file); err != nil {
		fmt.Println("\n? Error unmarshalling",n,err)
		return
	}

	var pathList []PathValue
	var agent_speaking string = "unknown agent"
	var tool_name string
	var text_body string
	var msg_body string
	var channel string
	var is_fault bool
	var event_type string
	var message_type string
	var other_agent string
	var seqnr int
	var speaks string
	
	// CAUTION - this is inherently fragile due to the specific model detail.
	// An event can be a report, a tool invocation/response, or internal reasoning

	fmt.Print(".")

	if FindJsonStructureAsPaths("",json_file, &pathList) {

		for _, item := range pathList {

			lhs := item.Path
			rhs := CleanString(item.Value.(string))

			switch lhs {

			case "agent":
				agent_speaking = rhs
				AGENTS[rhs] = true	
			case "text":
				text_body = rhs
			case "tool":
				tool_name = rhs
			case "error":
				if rhs == "true" {
					is_fault = true
				}
			case "seq":
				fmt.Sscanf(rhs,"%d",&seqnr)
				
			case "input.content":
				// only for tool_call
				msg_body = rhs
			case "input.to":
				other_agent = rhs
			case "counterpart":
				other_agent = rhs

			case "channel":
				// can only be one of these at a time
				channel = rhs

				switch rhs {

				case "tool_call":
					message_type = "tool_call"
					event_type = "agent invokes a tool / API call"

				case "tool_result":
					message_type = "tool_result"
					event_type = "agent accepts a tool response"

				case "peer_out":
					message_type = "+p2p"
					event_type = "peer message [+] offered"
				case "peer_in" :
					message_type = "-p2p"
					event_type = "peer message [-] accepted"

				// intro/extrospection
					
				case "thinking":
					event_type = "introspective reasoning"
				case "note":
					event_type = "status report / reasoning"
				case "say":
					event_type = "agent speaks to the principal in rich markdown format"
				case "principal":
					event_type = "the user / principal / boss instructs"
				case "world":
					event_type = "report world state / observation"
					// actor =?
				}
			}
		}
	}

	this_event := ThisEvent(seqnr,n,event_type,agent_speaking,tool_name)

	if string(log_event) == string(previous_event) {
		fmt.Println("\n!! Retry of agent operation:",this_event)
		SST_DUPLICATES = append(SST_DUPLICATES,this_event)
	}
	
	if JunkMessage(message_type,text_body) {
		return
	}

	// Start with tool calls

	speaks += RecordEvent(agent_speaking,this_event)
	
	promise_direction := MakeP2P(agent_speaking,other_agent)

	if message_type == "tool_call" {
		speaks := RecordToolCall(this_event,seqnr,n,event_type,promise_direction,tool_name,msg_body,is_fault)
		AGENT_TOOL_CALLS[tool_name] = append(AGENT_TOOL_CALLS[tool_name],this_event)
		AGENT_TIMELINE[agent_speaking] = append(AGENT_TIMELINE[agent_speaking],speaks)
		return
	}

	// tool call may induce p2p communication channel

	ConnectP2PMessages(this_event,text_body,message_type,promise_direction)

	// Collect tool result

	if message_type == "tool_result" {
		speaks += RecordToolResult(this_event,text_body,promise_direction,is_fault)
		AGENT_TIMELINE[agent_speaking] = append(AGENT_TIMELINE[agent_speaking],speaks)
		return
	}
	
	// MCP markdown tooling?

	if channel == "say" {
		// If it's an MCP the json might contain Markdown as content!!
		
		ismd,selections,_,L := IsMarkdownFragment(text_body,"run history")

		if ismd {
			speaks += RecordEvent(agent_speaking,this_event)
			f,_,_,_ := SST.ExtractIntentionalTokens(L,selections,SST.N_GRAM_MIN,SST.N_GRAM_MAX)
			speaks += CheckAnomalies(agent_speaking,f,n)
			return
		}
	}

	// Default is introspective

	speaks += RecordEvent(agent_speaking,this_event)
	speaks += AppendMessageToEvent(agent_speaking,text_body)
	speaks += AppendImpositionRequests(agent_speaking,text_body)
	speaks += ScanForSemanticAnomaly(text_body,n)

	// If there is retrying in the timeline, repeated entries may occur
	
	AGENT_TIMELINE[agent_speaking] = append(AGENT_TIMELINE[agent_speaking],speaks)
}

//**************************************************************

func MakeP2P(start,end string) P2Pair {

	var vector P2Pair
	vector.From = start
	vector.To = end
	return vector
}

//**************************************************************

func ConnectP2PMessages(this_event,msg_body,mt string,p2p P2Pair) {

	if mt == "+p2p" || mt == "-p2p" {

		// store P2P between-interaction from independent events
		// for later, as we can't connect them yet...
		
		msg := P2P_MESSAGE[msg_body]
		
		switch mt {
			
		case "-p2p":
			msg.To = p2p.From
			msg.AcceptEvent = this_event
			// Now we are receiving, so need to invert the promise arrow
			var invp2p P2Pair
			invp2p.To = p2p.From
			invp2p.From = p2p.To
			LAST_ACCEPT[invp2p] = this_event

		case "+p2p":
			msg.From = p2p.From
			msg.OfferEvent = this_event
			msg.ToolStart = LAST_TOOL_CALL[p2p]
		}
		
		P2P_MESSAGE[msg_body] = msg
	}
}

//**************************************************************

func JunkMessage(thistype,s string) bool {

	if thistype == "tool_result" {
		return false
	}
	
	const exposition_threshold = 50

	if len(s) < exposition_threshold {
		return  true
	}

	// Not words
	
	if !strings.Contains(s," ") {
		return true
	}

	return false
}

//**************************************************************

func ThisEvent(seq,n int,event,agent_speaking,tool_name string) string {

	if tool_name != "" {
		event += fmt.Sprintf(" from %s ",tool_name)
	}
	
	return fmt.Sprintf("event at %d: %s [%s], ref x%x",seq,event,agent_speaking,n)
}

//**************************************************************

func RecordEvent(agent,event string) string {

	return fmt.Sprintf("\n%s\n",event)
}

//**************************************************************

func RecordToolCall(this_event string,seq,n int,event_type string,p2p P2Pair,tool,body string,fault bool) string {

	agent_speaking := p2p.From
	speaks := ""

	// A tool event is an access of a centrally shared tool API
	speaks += fmt.Sprintf("\n%s\n",this_event)

	// The tool acts as a hub for such interactions
	speaks += fmt.Sprintf("    \"   (tool) %s\n",tool)
	// The agent calling also acts as a hub for such interactions
	speaks += fmt.Sprintf("    \"   (%s) %s\n",SST.CONT_REF_S,agent_speaking)

	// cache the sender of the P2P message for future linkage / a reasonable gamble
	LAST_TOOL_CALL[p2p] = this_event

	// Now formulate the graph annotations
	if fault {
		speaks += fmt.Sprintf("  \"  (warn) execution fault\n")
	}

	if body != "{}" && len(body) > 0 {
		speaks += AppendMessageToEvent(agent_speaking,body)
		speaks += AppendImpositionRequests(agent_speaking,body)
		speaks += ScanForSemanticAnomaly(body,seq)
		
	}

	AGENT_TIMELINE[agent_speaking] = append(AGENT_TIMELINE[agent_speaking],speaks)

	return speaks
}

//**************************************************************

func AppendMessageToEvent(agent,node string) string {

	return fmt.Sprintf("    \"  (message body) '%s' \n",node)
}

//**************************************************************

func RecordToolResult(this_event,body_text string,p2p P2Pair,fault bool) string {

	other_agent := LookForAgentInString(body_text)
	p2p.To = other_agent

	// Connect current event to past p2p recv, if it exists
	previous := LAST_ACCEPT[p2p]

	if len(previous) > 0 {
		result := fmt.Sprintf("\n '%s' \n",this_event)
		result += fmt.Sprintf("    \"   (result-of) '%s'\n",previous)

		if fault {
			result += fmt.Sprintf("    \" (warn) Tool call failed!\n\n")
		}

		// Append for later sundry
		AGENT_REPLY = append(AGENT_REPLY,result)
	}

	return this_event
}

//**************************************************************

func LookForAgentInString(s string) string {

	for a := range AGENTS {
		if strings.Contains(s,a) {
			return a
		}
	}

	return ""
}

//**************************************************************

func AppendImpositionRequests(agent_speaking,message string) string {
	
	if strings.Contains(strings.ToLower(message),"request") {
		speak := fmt.Sprintf("\n  \"  (imposes) action requested\n")
		return speak
	}

	return ""
}

//**************************************************************

func CheckAnomalies(agent string, mdfractions [][]string,seq int) string {

	const small_text = 50
	annotation := ""
	
	for _,sel := range mdfractions {
		for _,s := range sel {
			if len(s) > small_text {
				annotation += fmt.Sprintf("\n  \" (action) %s\n",s)
				annotation += ScanForSemanticAnomaly(s,seq)
			}
		}
	}

	return annotation
}

//*******************************************************************

func ListSundries(fp *os.File) {

	fmt.Fprintf(fp,"\n :: tools :: \n\n")

	for tool := range AGENT_TOOL_CALLS {

		myevents := AGENT_TOOL_CALLS[tool]

		for _,event := range myevents {
			fmt.Fprintf(fp," \" (toolused) %s \n",event)
		}

		fmt.Fprintf(fp,"\n %s (role) agent API / tool \n",tool)
	}


	fmt.Fprintf(fp,"\n :: messages, peer communication :: \n\n")

	for a := range P2P_MESSAGE {

		msg := P2P_MESSAGE[a]

		if len(msg.From) > 0 && len(msg.To) > 0 && len(msg.OfferEvent) > 0 && len(msg.AcceptEvent) > 0 {

			// Connect the p2p sub-process around a shared message

			// Message, contents, and start event
			fmt.Fprintf(fp,"\n '%s' (p2p-init) '%s' \n",msg.ToolStart,msg.OfferEvent)
			fmt.Fprintf(fp,"\n  \"  (message body) '%s' \n",a)
			fmt.Fprintf(fp,"\n '%s' (sender) '%s'\n",a,msg.From)
			fmt.Fprintf(fp,"\n   \" (receiver) '%s'\n",msg.To)

			// p2p channel forward direction
			fmt.Fprintf(fp,"\n '%s' (p2p-plus) '%s' \n",msg.OfferEvent,msg.AcceptEvent)
			//fmt.Fprintf(fp,"\n  \" (message body) '%s' \n",a)
			fmt.Fprintf(fp,"\n '%s' (p2p-minus) '%s' \n",msg.OfferEvent,msg.AcceptEvent)
			//fmt.Fprintf(fp,"\n  \" (message body) '%s' \n",a)
		}
	}

	fmt.Fprintf(fp,"\n :: agent replies :: \n\n")

	for _,reply := range AGENT_REPLY {
		fmt.Fprintf(fp,"\n %s \n",reply)
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

func TimeLines(fp *os.File) {

	fmt.Fprintf(fp,"\n :: ######################################## \n")
	fmt.Fprintf(fp,"\n :: ## Timelines \n")
	fmt.Fprintf(fp,"\n :: ######################################## \n\n")

	start := fmt.Sprintf("\n START! Agent execution launch ....\n\n")

	for agent := range AGENTS {

		fmt.Fprintf(fp,"\n :: agent transcript, agent trace, agent %s :: \n\n",agent)
		
		var opening string
		fmt.Fprintf(fp,"%s\n",start)
		fmt.Fprintf(fp,"\n \" (incl) %s\n",agent)

		if AGENT_OPENING[agent] != "" {
			opening = fmt.Sprintf("Agent [%s] begins with opening statement",agent)
			fmt.Fprintf(fp,"\n \" (then) %s\n",opening)
			fmt.Fprintf(fp,"\n %s (message body) '%s'\n",opening,AGENT_OPENING[agent])
		}

		fmt.Fprintf(fp,"\n\n +:: _sequence_ :: \n")

		// Now we have accumulated a sequential line of events, with
		// annotations elsewhere that will be added later

		for n,event := range AGENT_TIMELINE[agent] {

			fmt.Fprintln(fp,event)
			
			if n == 0 && opening != "" {
				// link back to openings
				fmt.Fprintf(fp," \" (from) '%s'\n",opening)
			}
		}

		fmt.Fprintf(fp,"\n -:: _sequence_ :: \n")
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
				//fmt.Println("\n ** inferring agent name: ",p)
				if AGENTS[p] {
					//fmt.Println(" ... ( already known)")
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
	}

}

//*******************************************************************

func CleanString(s string) string {

	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s,"''","DBL-QU")
	s = strings.ReplaceAll(s,"'","`")
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

func ScanForSemanticAnomaly(str string,n int) string {

	var scored bool
	var warning string
	
	for category := range SST_COORDS {
		
		for _,check := range SST_COORDS[category] {

			score,note := ScanFractionPatterns(category,str,check)

			warning += note
			
			if !strings.HasSuffix(warning,"..") {
				warning += ".."
			}

			if score > 0 {
				scored = true
				SST_SCORES[category] += score
				remember := fmt.Sprintf("event %d, %s(%s)-->(%s)",n,category,check,note)
				SST_ANOMALIES = append(SST_ANOMALIES,remember)
			}
		}
		
	}

	if scored && len(warning) > 0 {
		speaks := fmt.Sprintf("\n  \"  (warn) 'possible warning signal containing: %s' \n",warning)
		return speaks
	}
	
	return ""
}

//**************************************************************

func ScanFractionPatterns(category, text, check string) (int,string) {

	var paths [][]string
	var matches int
	var all_paths int = 1
	var note string
	
	// expand the compact encoding
		
	for _,order := range strings.Split(check,"-") {

		var seq []string
		
		for _,alt := range strings.Split(order,"/") {
			seq = append(seq,alt)
		}

		paths = append(paths,seq)
		all_paths *= len(seq)
	}

	var clock = make([]int,len(paths))
	
	for turn := 0; turn < all_paths; turn++ {

		var pattern []string
		var found bool

		for pos := 0; pos < len(paths); pos++ {
			pattern = append(pattern,paths[pos][clock[pos]])
		}
		
		found,note = PatternMatches(text,pattern)

		if found {
			matches++
		}

		// advance clock, fast hand first
		
		if clock[0] < len(paths[0]) {
			clock[0]++
		}
		
		// clock wraps around

		for pos := 0; pos < len(paths); pos++ {
			
			if clock[pos] == len(paths[pos]) {
				clock[pos] = 0

				if pos < len(paths)-1 {
					clock[pos+1]++
				} else {
					continue
				}
			}
		}

		if clock[len(paths)-1] == len(paths) {
			break
		}
	}

	
	return matches,note
}

//**************************************************************

func PatternMatches(ref string,pattern []string) (bool,string) {

	working := ref
	found_word := false
	found := ""
	before := ""
	
	// check the whole sequence of words in order
	
	for occur, word := range pattern {

		before, working, found_word = strings.Cut(working,word)
		
		if !found_word {
			return false,""
		}

		const together_range_limit = 20
		
		if occur > 1 && len(before) > together_range_limit {
			return false,""
		}
		
		// extract an source reference / annotation
		
		const margin = 30
		start := max(0, len(before)-margin)
		found += before[start:] + "[" + strings.ToUpper(word) + "] "
		end := min(margin, len(working))
		found += working[:end] + " / "

	}
	
	return true,found
}

//**************************************************************

func Excerpt(whole,part string) string {

	before,after,found := strings.Cut(whole,part)

	if found {
		const bracket = 50
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
	Path  string
	Value interface{}
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

		if result, ok := v.(float64); ok {
			*list = append(*list, PathValue{
				Path:  prefix,
				Value: fmt.Sprintf("%d",int(result)),
			})

			return true
		}
		
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

