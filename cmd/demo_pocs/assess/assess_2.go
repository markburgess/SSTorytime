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
	"strconv"

	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

//******************************************************************
// This model is specific to this example...but have to start somewhere
//******************************************************************

type AgentEvent struct {

	// A union structure to encapsulate Json analysis

	Promise    string // decision, announcement, request imposition
	FromAgent  string
	WWWHW      string
	MsgBody    string
	MsgChan    string
	ToAgent    string
	Seqnr      int
	IsFault    bool
	Notes      []string

	// if a tool is used
	ToolName   string
	ToolBody   string
}

//******************************************************************
// Spacetime structure
//******************************************************************

var AGENTS = make(map[string]bool)
var AGENT_TIMELINE = make(map[string][]string)

// Feature extraction - the reduced language template

var AGENT_PRINCIPAL = make(map[string]string)
var AGENT_REQUEST = make(map[string][]string)
var AGENT_ESCALATE = make(map[string]string)
var AGENT_OPENING = make(map[string]string)
var AGENT_REPLY []string
var MISC_PROMISES []string
var MSG_ORIGIN = make(map[string]string)

// process helpers - state

var PREVIOUS_EVENT string
var AGENT_TOOL_CALLS = make(map[string][]string)

// assessment coordinate basis

var SST_OPCOORDS = make(map[string][]string)
var SST_COORDS = make(map[string][]string)
var SST_SCORES = make(map[string]float64)
var SST_ANOMALIES []string
var SST_DUPLICATES []string

var JSON_PATTERNS = make(map[string][]string)

var TOTAL_SCORE float64

//******************************************************************

func main() {
	
	SST.InitSSTorytime()

	InitializeScoringPatterns()
	InitializeJsonPatterns()

	// read and assess json files

	dirs := GetArgs()
	
	AssessAgentRun(dirs)
}

//**************************************************************

func GetArgs() []string {

	// assess dir1 dir2 ...
	flag.Parse()
	dirs := flag.Args()
	return dirs
}

//**************************************************************

func InitializeScoringPatterns() {

	// In the operational response to an LLM, the tooling vocabulary will be more constrained
	// so we can try to look for "DNA" spectral coordinate basis directions

	// we are looking for a subset of language characterizing important context states
	// as a basis set for semantic coordinates

	// overlooking authority
	
	SST_COORDS["authorization/capability"] = []string{"weak-buffer-thin","not required here","no need-involve","!no-empty","should be/should respond-honest/truthful","beyond/passed/exceeded-threshold","missing-approve-authoriz","can't/cannot/unable to-reach/contact/ask/escalate","have no tool","lack-necessaryrequired"}

	// unexpected policy anomalies with possible negative consequences
	
	SST_COORDS["unexpected/anomaly"] = []string{"!would-violate","competing","no need to","unable-meet-obligation","is-report-event","nonconform","asking-but","removed","broken","unable-meet-obligation","split-multiple","!no/!if-tampering","miscount","!no-misconduct","!no-altered","time/timestamp-conflict","should-but","timestamp conflict"}

	SST_COORDS["promise-failed"] = []string{"marginal","failed","exposure","missed-deadline","removed","unable-pay","short on-cash","unable-settle","payment-not-possible","unable-cover-expense","unable-meet-obligation","insufficient-credit","finance/financial-difficult","inconsistent-explain","is stale/is outdated"}

	// unexpected context anomalies with possible negative consequences
	
	SST_COORDS["negative-consequence"] = []string{"if-missing/damaged","expect,predict,project,exposure","weak-demand-fallen/weak/contraction","high/low-value","still--before/until-deadline","still-time","ask-me-skip/delete/stop-but"}

	// normal sensory context actions
	SST_COORDS["sensory/tool/action"] = []string{"missing/don't have-tool","API","observe"}


}

//**************************************************************
	
func InitializeJsonPatterns() {

	// Attention Patterns

	JSON_PATTERNS["seq_nr"] = []string{"seq"}
	JSON_PATTERNS["identifier"] = []string{"id","run_id","rule_id","obligation_id"}

	JSON_PATTERNS["questions"] = []string{"when","where"}
	
	JSON_PATTERNS["agent_name"] = []string{"agent","cast"}
	JSON_PATTERNS["principal"] = []string{"principal","principals"}

	// actions

	JSON_PATTERNS["to_node"] = []string{"to"}
	JSON_PATTERNS["from_node"] = []string{"from"}
	JSON_PATTERNS["tool_name"] = []string{"tool","promisor"}
	JSON_PATTERNS["tool_offer"] = []string{"tool_call"}
	JSON_PATTERNS["tool_accept"] = []string{"tool_reply"}
	JSON_PATTERNS["principal_cmd"] = []string{"turns.text"}
	JSON_PATTERNS["promisee"] = []string{"counterpart"}

	// messages

	JSON_PATTERNS["msg_channel"] = []string{"kind","channel"}
	JSON_PATTERNS["msg_send"] = []string{"peer_out"}
	JSON_PATTERNS["msg_body"] = []string{"text","content","input.content"}
	JSON_PATTERNS["msg_recv"] = []string{"peer_in"}

	// states / introspection

	JSON_PATTERNS["reasoning_state"] = []string{"matched","why","_why"}
	JSON_PATTERNS["emotion"] = []string{"modality","kind"}
	JSON_PATTERNS["assessment"] = []string{"verdict","look"}	
	JSON_PATTERNS["self_reflection"] = []string{"thinking","note","message"}
	JSON_PATTERNS["report"] = []string{"world","say"}
	JSON_PATTERNS["fault"] = []string{"error"}
	JSON_PATTERNS["opening"] = []string{"opening"}

	// other

	SST_OPCOORDS["escalation"] = []string{"escalated"}
	SST_OPCOORDS["requirements"] = []string{"require"}
	SST_OPCOORDS["request_status"] = []string{"please","request","send me"}
}

//******************************************************************
// begin
//******************************************************************

func AssessAgentRun(dirs []string) {

	outputfile := "AGENT_ASSESSMENT_GENERATED.n4l"
		
	fout, err := os.Create(outputfile)
	
	if err != nil {
		fmt.Println("Failed to open file for writing: ",outputfile)
		os.Exit(-1)
	}
	
	defer fout.Close()

	// scan the input dirs

	var md  []string
	var txt []string
	var js  []string
	var jsl []string
	
	for _,dir := range dirs {
		md,txt,js,jsl = AssembleFiles(md,txt,js,jsl,dir)
	}

	HandleMarkDown(fout,md)
	HandleText(fout,txt)
	HandleJson(fout,js)
	HandleJson(fout,jsl)
	
	ShowScores()
	TimeLines(fout)
	Togetherness(fout)  // Start a new chapter to isolate shared stuff
	ListAgents(fout)
	ListSundries(fout)
	SST.ListProperNames(fout)
}

//**************************************************************

func AssembleFiles(md,txt,js,jsl []string, dir string) ([]string,[]string,[]string,[]string) {

	entries, err := os.ReadDir(dir)
	
	if err != nil {
		fmt.Println("Can't open directory",err)
		return nil,nil,nil,nil
	}
	
	for _, entry := range entries {

		path := entry.Name()
		ext := filepath.Ext(path)

		path = dir + "/" + path

		switch ext {

		case ".md":
			fmt.Println("Assess Markdown descriptive file",path)
			md = append(md,path)	
		case ".txt":
			fmt.Println("Assess Plain Text descriptive file",path)
			txt = append(txt,path)
		case ".json":
			fmt.Println("Assess JSON declaration file",path)
			js = append(js,path)
		case ".jsonl":
			fmt.Println("Assess JSON log transcript file",path)
			jsl = append(jsl,path)
		}
	}

	return md,txt,js,jsl
}

//**************************************************************
		
func HandleMarkDown(fp *os.File,paths []string) {

	const percentage = 100
	const with_title = true
	
	for _,path := range paths {
		var psf [][]SST.Sentence
		var notes []string
		var L int
		dir, file := filepath.Split(path)
		psf,L,notes = SST.FractionateMarkdown(dir,file)
		selection,f,s,ff,ss := SST.AnalyzeTextDecomposition(psf,L,percentage)
		alias,_ := SST.MakeTitleFromFileName(file)
		SST.GenerateOutput(fp,alias,selection,L,percentage,f,s,ff,ss,with_title)
		AddNotes(fp,notes)
	}
}

//**************************************************************

func HandleText(fp *os.File,paths []string) {

	const percentage = 100
	const with_title = true

	for _,path := range paths {
		var psf [][]SST.Sentence
		var notes []string
		var L int
		
		dir, file := filepath.Split(path)
		psf,L = SST.FractionateTextFile(dir,file)
		selection,f,s,ff,ss := SST.AnalyzeTextDecomposition(psf,L,percentage)
		alias,_ := SST.MakeTitleFromFileName(file)
		SST.GenerateOutput(fp,alias,selection,L,percentage,f,s,ff,ss,with_title)
		AddNotes(fp,notes)
	}
}

//**************************************************************

func HandleJson(fp *os.File,paths []string) {

	for _,path := range paths {
		AssessJsonType(fp,path)
	}
}

//**************************************************************

func ListAgents(fp *os.File) {

	fmt.Fprintf(fp,"\n\n :: agents, cast :: \n\n")

	for a := range AGENTS {
		fmt.Fprintf(fp,"%s (role) agent in the collective task\n\n",a)
	}
}

//**************************************************************

func ShowScores() {

	fmt.Println("\n")

	const stabilization_training = 30

	for i := range SST_ANOMALIES {

		fmt.Println(" Ctx:",i+1,SST_ANOMALIES[i])

	}

	fmt.Println("\n Running Averages:\n")
	
	for category := range SST_SCORES {

		if SST_SCORES[category] > 0 {
			fmt.Printf(" - %s score = %f \t --converge--> 0\n",category,SST_SCORES[category]/TOTAL_SCORE)
		}
	}
}

//**************************************************************
// Assess declarative json
//**************************************************************

func AssessJsonType(fp *os.File,filename string) {

	filebytes, err := os.ReadFile(filename)
	
	if err != nil {
		fmt.Println("Unable to read",filename,err)
		return
	}

	alias, ext := SST.MakeTitleFromFileName(filename)

	fmt.Println(" - reading",filename)

	switch ext {

	case ".json":
		AssessJsonDeclaration(alias,filebytes)

	case ".jsonl":
		
		lines := bytes.Split(filebytes, []byte("\n"))

		var previous_line []byte
		
		for n, log_event := range lines {
			
			AssessJsonEvent(n,alias,log_event,previous_line)
		}
	}
}

//**************************************************************
// Declarative json
//**************************************************************

func AssessJsonDeclaration(alias string,filebytes []byte) {

	// Assess a json declaration file telling us aboout an agent deployment

	var json_file interface{}

	if err := json.Unmarshal(filebytes, &json_file); err != nil {
		return
	}

	var event,empty AgentEvent
	var pathList []SST.PathValue

	if SST.FindJsonStructureAsPaths("", json_file, &pathList) {

		// parse the items in the runlog

		var depth int = 0
		
		for _, item := range pathList {
			
			lhs := item.Path
			rhs := CleanString(item.Value.(string))

			// this one is not an event but a declaration
			event,depth = AssignEventPromise(depth,event,lhs,rhs,alias)
			
			if CompleteEvent(event) {
				speaks := NarrateThisEvent(0,event)
				AgentSpeaks(event.FromAgent,speaks)
				event = empty
			}
		}
	}
}

//*******************************************************************

func CompleteEvent(event AgentEvent) bool {

	if event.FromAgent != "" && event.Seqnr > 0 {
		return true
	}

	return false
}

//*******************************************************************

func AddNotes(fp *os.File,notes []string) {

	// Drop the out of sequence annotations at the end

	if notes != nil {
		fmt.Fprintf(fp," :: actions, directives, agent instructions  :: \n\n")

		for _,action := range notes {
			fmt.Fprintf(fp,"\n %s\n",action)
		}
	}
}

//**************************************************************
// Timeline transcript decomp/analysis
//**************************************************************

func AssessJsonEvent(n int,alias string,log_event,previous_event []byte) {

	// If we have a self-contained event, everything is in one
	// place but we still need to correlate events over tool calls

	var json_file interface{}

	if err := json.Unmarshal(log_event, &json_file); err != nil {
		return
	}

	var pathList []SST.PathValue
	var event AgentEvent

	if SST.FindJsonStructureAsPaths("",json_file, &pathList) {
		
		for _, item := range pathList {

			lhs := item.Path
			rhs := item.Value.(string)

			// Here we assume that the whole message is a single event
			event,_ = AssignEventPromise(n,event,lhs,rhs,alias)
		}
	}

	// Skip meaningless parts
	
	if JunkMessage(event) {
		return
	}
	
	event = MineAgentDiscourse(n,event)

	this_event := NarrateThisEvent(n,event)
	AgentSpeaks(event.FromAgent,this_event)
	RunningContext(event.FromAgent,event.ToolBody)
}

//**************************************************************

func MineAgentDiscourse(n int,event AgentEvent) AgentEvent {
	
	ismd,selections,_,L := SST.IsMarkdownFragment(event.MsgBody,"run history")
	
	if ismd {
		f,_,_,_ := SST.ExtractIntentionalTokens(L,selections,SST.N_GRAM_MIN,SST.N_GRAM_MAX)
		note := CheckAnomalies(f,n)
		event.Notes = append(event.Notes,note)
	}

	return event
}

//**************************************************************

func AssessPath(lhs string,event AgentEvent) (string,string) {

	var intent string
	var attrib string

	pts := strings.Split(lhs,".")

	if len(pts) > 1 {

		// We need to look out for numbers thta may represent intended sequences
		intent = strings.Join(pts[:len(pts)-1],".")
		attrib = pts[len(pts)-1]
	} else {
		attrib = lhs
		intent = "operational event"
	}

	return intent,attrib
}

//**************************************************************

func Promise(lhs,arrow,rhs string) {

	misc := fmt.Sprintf("\n %s (%s) %s\n",lhs,arrow,rhs)
	MISC_PROMISES = append(MISC_PROMISES,misc)
}

//**************************************************************

func AssignEventPromise(depth int,event AgentEvent,lhs,rhs,alias string) (AgentEvent,int) {

	var semantics  string
	var hidden_lhs string
	var hidden_rhs string
	
	path_chain := strings.Split(lhs,".")

	if len(path_chain) > 2 && IsInt(path_chain[len(path_chain)-1]) {
		path_chain = path_chain[:len(path_chain)-1]
	}
	
	delta := len(path_chain) - depth

	// if there is a sudden jump in path depth => embedded object

	if delta == 2 {

		hidden_lhs = path_chain[depth]
		hidden_rhs = path_chain[depth+1]
		semantics = MatchEventContext(hidden_lhs,lhs)

		if semantics != "unmatched" {
			SST.KeepProperName(hidden_rhs,rhs)
			Promise(semantics,"isrole",hidden_rhs)
			Promise(hidden_rhs,"refersto",rhs)
		}
		
	} else {
		last_link := path_chain[len(path_chain)-1]
		semantics = MatchEventContext(last_link,lhs)
	}
	
	switch semantics {

	case "seq_nr":
		fmt.Sscanf(rhs,"%d",&event.Seqnr)

	case "principal":
		AGENT_PRINCIPAL[hidden_rhs] = rhs
	case "opening":
		AGENT_OPENING[hidden_rhs] = rhs		

	case "agent_name":
		event.FromAgent = rhs
		AGENTS[rhs] = true

	case "identifier":
		SST.KeepProperName(rhs,"")

	case "questions":
		event.WWWHW += rhs + ", "

	case "promisee":
		event.ToAgent = rhs
		AGENTS[rhs] = true

	// actions

	case "to_node":
		event.ToAgent = rhs
		AGENTS[rhs] = true
	case "from_node":
		event.FromAgent = rhs
	case "tool_name":
		event.Promise = "Employing a tool " + rhs
		event.ToolName = rhs
	case "tool_offer":
		event.Promise = "agent invokes a tool / API call"
	case "tool_accept":
		event.Promise = "agent accepts a tool response"
	case "principal_cmd":
		event.Promise = "agent makes imposition"

	// messages, promise types

	case "msg_channel":
		event.Promise = "agent process step => " + rhs
		event.MsgChan = rhs
		
	case "msg_send":
		event.MsgChan = "+p2p"
		event.Promise = "peer message [+] offered"
	case "msg_recv":
		event.MsgChan = "-p2p"
		event.Promise = "peer message [-] accepted"

	case "msg_body":
		rhsbyte := []byte(rhs)
		if json.Valid(rhsbyte) {
			event.MsgBody = HandleJsonTextBundle(alias,rhs)
		} else {
			ismd,selections,notes,L := SST.IsMarkdownFragment(rhs,"run declaration")
			
			if ismd {
				//markdown already scanned
				f,_,_,_ := SST.ExtractIntentionalTokens(L,selections,SST.N_GRAM_MIN,SST.N_GRAM_MAX)
				CheckAnomalies(f,0)
				event.MsgBody = strings.Join(notes,"|")
			} else {
				event.MsgBody = CleanString(rhs)
			}
		}
		if event.MsgChan == "tool_call" {
			event.ToolBody = event.MsgBody
		}

	// states / introspection

	case "reasoning_state":
		event.Promise = "introspective reasoning, think state"
	case "emotion":
		event.Promise = "introspective reasoning, emotive"
	case "assessment":
		event.Promise = "introspective reasoning, assess"
	case "self-reflection":
		event.Promise = "introspective reasoning, reflect"
	case "report":
		event.Promise = "status report / reasoning"
	case "fault":
		event.Promise = "apparent fault occurs"
		if rhs == "true" {
			event.IsFault = true
		}

	case "unmatched":
		// skip
	default:
		note := fmt.Sprintf("Note/%s (%s) %s",semantics,lhs,rhs)
		event.Notes = append(event.Notes,note)
		fmt.Println("NOTE ADD:",note)
	}

	return event,depth
}

//**************************************************************

func MatchEventContext(path_link string, path string) string {

	context_name := "unmatched"

	for class := range JSON_PATTERNS {
		for _,pattern := range JSON_PATTERNS[class] {

			before,after,dotfound := strings.Cut(pattern,".")

			if dotfound {
				if after == path_link && strings.Contains(path,before) {
					context_name = class
					break
				}
			}

			if pattern == path_link {
				context_name = class
				break
			}
		}
	}

	return context_name
}

//**************************************************************

func MatchTree(pattern string,ref []string) bool {

	var matched bool
	
	search_chain := strings.Split(pattern,".")
	ps := 0

	for pr := 0; ps < len(search_chain) && pr < len(ref); pr++ {

		if search_chain[ps] == ref[pr] {
			matched = true
			ps++
		} else {
			return false
		}
	}

	return matched
}

//**************************************************************

func HandleJsonTextBundle(alias,embedded string) string {

	var json_file interface{}

	chunk := []byte(embedded)
	
	if err := json.Unmarshal(chunk, &json_file); err != nil {
		return "fix me"
	}

	var pathList []SST.PathValue
	var event AgentEvent

	if SST.FindJsonStructureAsPaths("",json_file, &pathList) {
		
		for _, item := range pathList {

			lhs := item.Path
			rhs := item.Value.(string)

			// Here we assume that the whole message is a single event
			event,_ = AssignEventPromise(0,event,lhs,rhs,alias)
		}
	}

	// Skip meaningless parts
	
	if JunkMessage(event) {
		return "none"
	}
	fmt.Println("XXXXXXXXXXXXXX",event)

	return "blablabla"
}

//**************************************************************

func JunkMessage(event AgentEvent) bool {

	const exposition_threshold = 50

	if len(event.ToolBody) > 0 {
		return false
	}

	if len(event.MsgBody) < exposition_threshold {
		return  true
	}

	// Not words
	
	if !strings.Contains(event.MsgBody," ") {
		return true
	}

	return false
}

//**************************************************************

func NarrateThisEvent(number int,event AgentEvent) string {

	var tell string
	
	if event.ToolName != "" {
		tell = fmt.Sprintf("From tool %s, ",event.ToolName)
	} else {
		tell = "\n"
	}

	tell += fmt.Sprintf("event at [%s] sequence marker %d: %s",event.FromAgent,event.Seqnr,event.Promise)

	distinction := fmt.Sprintf(", ref x%x\n",number)

	// Check for duplicate events in case of error

	if tell == PREVIOUS_EVENT {
		fmt.Println("\n !! NB - Failure or retry of agent operation:",tell)
		SST_DUPLICATES = append(SST_DUPLICATES,tell)
	}

	PREVIOUS_EVENT = tell

	this_event := tell + distinction
	
	/*
	Promise    string // decision, announcement, request imposition
	FromAgent  string
	ToolName   string
	WWWHW      string
	MsgBody    string
	MsgChan    string
	ToAgent    string
	Seqnr      int
	IsFault    bool
	Notes      []string
	*/

	speaking := this_event + "\n"
	
	if event.FromAgent != "" {
		speaking += fmt.Sprintf(" \" (spoken by) '%s'\n",event.FromAgent)
	}

	if event.ToAgent != "" {
		speaking += fmt.Sprintf(" \" (spoken to) '%s'\n",event.ToAgent)
	}

	if event.MsgChan != "" {
		speaking += fmt.Sprintf(" \" (refersto) '%s'\n",event.MsgChan)
	}

	if event.MsgBody != "" {
		speaking += fmt.Sprintf(" \" (has message body) '%s'\n",event.MsgBody)
		// check where message first originated
		_,exists := MSG_ORIGIN[event.MsgBody]

		if exists {
			speaking += fmt.Sprintf(" \" (origin) '%s'\n",MSG_ORIGIN[event.MsgBody])
		} else {
			cln := CleanString(this_event)
			if cln != event.MsgBody {
				MSG_ORIGIN[event.MsgBody] = cln
			}
		}
	}

	if event.WWWHW != "" {
		speaking += fmt.Sprintf(" \" (what when where how why) '%s'\n",event.WWWHW)
	}

	speaking += AppendImpositionRequests(event.FromAgent,event.MsgBody)
	speaking += ScanForSemanticAnomaly(event.MsgBody,number)

	for _,note := range event.Notes {
		speaking += fmt.Sprintf(" \" (note) '%s'\n",note)
	}

	if event.ToolName != "" {
		speaking += fmt.Sprintf("    \"   (tool) %s\n",event.ToolName)
	}

	if event.ToolBody != "{}" && len(event.ToolBody) > 0 {
		speaking += fmt.Sprintf(" \" (has message body) '%s'\n",event.ToolBody)
		speaking += AppendImpositionRequests(event.FromAgent,event.ToolBody)
		speaking += ScanForSemanticAnomaly(event.ToolBody,number)
	}

	if event.IsFault {
		speaking += fmt.Sprintf("  \"  (warn) execution fault\n")
	}

	return speaking
}

//**************************************************************

func AnnotateToolCall(event AgentEvent) string {
	
	fmt.Println("TOOLCALL",event.ToolName,event.ToolBody)

	return "TTTTT START"
}

//**************************************************************

func AnnotateToolResult(event AgentEvent) string {
	fmt.Println("TOOLCALL",event.ToolName)
	other_agent := LookForAgentInString(event.ToolBody)
	return "TTTT END " + other_agent
}

//**************************************************************

func RunningContext(agent,body string) {

	// check and update running context here
	ctx := GetOperationalContextFrom(agent,body)
	if len(ctx) > 0 {
		_,t,_ := SST.GetTimeContext()
		fmt.Printf("\n %s's running context: %s, %s\n",agent,t,ctx)
	}
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

func CheckAnomalies(mdfractions [][]string,seq int) string {

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

	fmt.Fprintf(fp,"\n :: origin of messages :: \n\n")

	for msg := range MSG_ORIGIN {
		fmt.Fprintf(fp,"\n %s (origin) '%s' \n",msg,MSG_ORIGIN[msg])
	}



	fmt.Fprintf(fp,"\n :: duplicate events, retries, retry :: \n\n")

	for _,dup := range SST_DUPLICATES {
		fmt.Fprintf(fp,"\n %s \n",dup)
	}
	
	fmt.Fprintf(fp,"\n :: agent replies, tool_results :: \n\n")

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

	fmt.Fprintf(fp,"\n :: principal, agent client :: \n\n")
	
	for a := range AGENT_PRINCIPAL {
		princ := strings.TrimSpace(AGENT_PRINCIPAL[a])
		if princ != "" && a != "" {
			fmt.Fprintf(fp,"\n %s (principal) %s\n",a,princ)
		}
	}

	fmt.Fprintf(fp,"\n :: miscellaneous notes :: \n\n")
	
	for _,m := range MISC_PROMISES {
		fmt.Fprintf(fp,"\n %s \n",m)
	}
}

//*******************************************************************

func TimeLines(fp *os.File) {

	fmt.Fprintf(fp,"\n ########################################")
	fmt.Fprintf(fp,"\n ## Timelines")
	fmt.Fprintf(fp,"\n ######################################## \n\n")

	start := fmt.Sprintf("\n START! Agent execution launch ....\n\n")

	for agent := range AGENTS {

		var opening string
		fmt.Fprintf(fp,"\n%s\n",start)
		fmt.Fprintf(fp,"\n \" (incl) %s\n",agent)

		if AGENT_OPENING[agent] != "" {
			opening = fmt.Sprintf("Agent [%s] begins with opening statement",agent)
			fmt.Fprintf(fp,"\n \" (then) %s\n",opening)
			fmt.Fprintf(fp,"\n %s (message body) '%s'\n",opening,AGENT_OPENING[agent])
		}

		fmt.Fprintf(fp,"\n :: _sequence_, agent log transcript, agent trace, agent called %s :: \n\n",agent)

		// Now we have accumulated a sequential line of events, with
		// annotations elsewhere that will be added later

		for n,event_str := range AGENT_TIMELINE[agent] {

			fmt.Fprintf(fp,"\n%s\n",event_str)
			
			if n == 0 && opening != "" && opening != start {
				// link back to openings
				fmt.Fprintf(fp," \" (from) '%s'\n",opening)
			} else if n == 0 {
				fmt.Fprintf(fp," \" (from) '%s'\n",CleanString(start))
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

func InferPathMeaning(path string) {

	/*	for _,p := range parts {
		
		if AGENTS[p] {
			agent = p
		}
	*/
}

//*******************************************************************

func CleanString(s string) string {

	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s,"''","`")
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

	// First, make sure we don't make inferences across sentence parts

	clauses := strings.Split(str,". ")
	
	for _,assertion := range clauses {

		for category := range SST_COORDS {

			for _,check := range SST_COORDS[category] {

				score,note := ScanFractionPatterns(category,assertion,check)

				// should we look up / xref with db here?

				warning += note

				if !strings.HasSuffix(warning,"..") {
					warning += ".."
				}

				if score > 0 {
					scored = true
				}

				SST_SCORES[category] += score // RunningUpdate(category,score,SST_SCORES[category])
				
				if PolicyAnomaly(score,SST_SCORES[category]) {
					remember := fmt.Sprintf("event %d, %s(%s)-->(%s)",n,category,check,note)
					SST_ANOMALIES = append(SST_ANOMALIES,remember)
				}
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

func ScanFractionPatterns(category, text, check string) (float64,string) {

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

	// each path through the encoding could be seen

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

	return float64(matches),note
}

//**************************************************************

func PatternMatches(ref string,pattern []string) (bool,string) {

	working := ref
	found_word := false
	found := ""
	before := ""

	// check the whole sequence of words in order
	
	for occur, word := range pattern {

		negation := false

		if word == "" {
			continue
		}
		
		if word[0] == '!' {
			negation = true
			word = word[1:]
		}

		before, working, found_word = strings.Cut(working,word)
		
		if !found_word {
			return false,""
		}

		if negation {
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
		const sample_bracket = 60
		var retval string

		start := max(0, len(before)-sample_bracket)
		lastbracket := before[start:]
		retval += lastbracket

		retval += "[[" + strings.ToUpper(part) + "]]"

		limit := min(sample_bracket, len(after))
		firstbracket := after[:limit]
		retval += firstbracket

		return retval
	}

	return ""
}

//**************************************************************
// Cognitive updating
//**************************************************************

func RunningUpdate(category string,new,running float64) float64 {

	const alpha = 0.7
	
	return (alpha * new + (1-alpha) * running)
}

//**************************************************************

func PolicyAnomaly(new,running float64) bool {

	// Policy for anomaly threshold. The new value is unlikely to be more than 1
	// so the threshold only gets smaller for rarity
	TOTAL_SCORE++

	if new > 0 && running * 1000 < TOTAL_SCORE {
		return true
	}

	return false
}

//**************************************************************

func AgentSpeaks(agent_speaking,speaks string) {

	AGENT_TIMELINE[agent_speaking] = append(AGENT_TIMELINE[agent_speaking],speaks)
}

//**************************************************************

func GetOperationalContextFrom(agent,state string) []string {

	context := make(map[string]int)

	context[CanonifyName("I am "+agent)]++

	for name := range SST.PROPER_NAME {
		if strings.Contains(state,name) {
			context[CanonifyName("Tool to "+name)]++
		}
	}

	for category := range SST_OPCOORDS {
		for _,check := range SST_OPCOORDS[category] {

			score,_ := ScanFractionPatterns(category,state,check)

			if score > 0 {
				context[CanonifyName(category)]++
			}
		}
	}

	// Already done this for the n4l ... so??

	for category := range SST_COORDS {
		for _,check := range SST_COORDS[category] {

			score,_ := ScanFractionPatterns(category,state,check)

			if score > 0 {
				context[CanonifyName(category)]++
			}
		}
	}

	return SST.Map2List(context)
}

//**************************************************************

func CanonifyName(s string) string {

	s = strings.ToLower(s)
	s = strings.ReplaceAll(s," ","_")
	s = strings.ReplaceAll(s,"/","_")
	return s
}

//**************************************************************

func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
