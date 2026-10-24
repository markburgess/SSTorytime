//*****************************************************************
//
// text_fractionation.go
//
//*****************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"sort"
	"regexp"
	_ "github.com/lib/pq"
)

//**************************************************************
// Text Fractionation (alphabetic language)
//**************************************************************

const N_GRAM_MAX = 6
const N_GRAM_MIN = 2  // fragments that are too small are exponentially large in number and meaningless

const DUNBAR_5 =5
const DUNBAR_15 = 15
const DUNBAR_30 = 45
const DUNBAR_150 = 150

// **************************************************************

var EXCLUSIONS []string

var STM_NGRAM_FREQ [N_GRAM_MAX]map[string]float64
var STM_NGRAM_LOCA [N_GRAM_MAX]map[string][]int
var STM_NGRAM_LAST [N_GRAM_MAX]map[string]int

var PROPER_NAME = make(map[string]string)
var TEXT_TABLE = make(map[string]string)

//**************************************************************

type TextRank struct {
	Significance float64
	Fragment     string
	Title        string
	Order        int
	Partition    int
}

//**************************************************************

type Sentence struct {
	S string
	Title string
	Frags []string
}

type AuxTxtMDSentence struct { // hangover from goldmark processing, tb-elim
	Level   int
	Title   string
	Number  string
	Content string
}

//**************************************************************
// Begin
//*****************************************************************

func ReadTextFile(filename string) string {

	// Read a string and strip out characters that can't be used in kenames
	// to yield a "pure" text for n-gram classification, with fewer special chars
	// The text marks end of sentence with a # for later splitting

	content,err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Couldn't find or open",filename)
		os.Exit(-1)
	}

	// Start by stripping HTML / XML tags before para-split
	// if they haven't been removed already

	m1 := regexp.MustCompile("<[^>]*>") 
	cleaned := m1.ReplaceAllString(string(content),";") 
	return cleaned
}

//******************************************************************

func FractionateTextFile(dir,name string) ([][]Sentence,int) {

	file := ReadTextFile(dir+name)
	proto_text := CleanText(file)
	name,_,_ = strings.Cut(name,".") // strip extension
	return FractionateText(proto_text,name)
}

//**************************************************************

func FractionateText(proto_text,filename string) ([][]Sentence,int) {

	pbsf,count := SplitIntoSectionsOfSentences(proto_text,filename)
	AnalyzeFractionIntent(pbsf)
	return pbsf,count
}

//**************************************************************

func SplitIntoSectionsOfSentences(file,filename string) ([][]Sentence,int) {

	var current AuxTxtMDSentence
	var title string = ""
	var level int

	intro := fmt.Sprintf("Introduction of %s",filename)

	current.Level = level
	current.Title = intro
	current.Number = ""
	current.Content = ""
	
	SECTIONS = append(SECTIONS, current)
	
	paras := strings.Split(file,"\n\n")

	for _,sec := range paras {

		// This is where we detect section headers

		issection,value,number := IsNewSection(sec,filename)
		
		if issection {
			// Set section.Title
			title = value
		}	

		current.Level = level
		current.Title = title
		current.Number = number
		current.Content = value

		SECTIONS = append(SECTIONS, current)
	}
	
	pbsf,count := AssemblePBSF(SECTIONS)
	return pbsf,count
}

// **************************************************************************

func IsNewSection(para string,filename string) (bool,string,string) {

	// See if we can classify a lump of text

	para = strings.ReplaceAll(para,"\n"," ")
	para = CheckAnnotations(para)
	para = CleanText(para)

	// Sections tend to start with a number or letter a),iv), 6.2.1, etc

	ischap, number, name := LooksLikeChapSection(para,filename)

	if ischap {
		return true,name,number
	}

	// This is not a header, just a normal paragraph

	return false,para,"unnumbered"
}

// **************************************************************************

func CheckAnnotations(s string) string {

	// Insert a "****" annotation where boldface **x y z...** has been used
	// Note, this assumes that **** is defined in the standard way

	const searchtext = "**"

	count := 0

	index := strings.Index(s,searchtext)

	if index < 0 {
		return s
	}
	
	for pos := index; pos < len(s); pos++ {

		if pos < len(s) - len(searchtext) {
			if s[pos:pos+2] == searchtext {
				count++
				pos += strings.Index(s[pos:],searchtext)
			}
		}
	}

	// Need matching **Start and end** to replace
	
	if count % 2 != 0 {
		return s
	}
	
	var openbrack bool = true
	var modified []rune
	runestring := []rune(s)
	rlen := len(runestring)

	for pos := 0; pos < rlen; pos++ {

		switch runestring[pos] {
		case '*':
			if pos+1 < rlen && runestring[pos+1] == '*' {

				if openbrack {
					modified = append(modified,'*')
					modified = append(modified,'*')
					modified = append(modified,'*')
					modified = append(modified,'*')
					modified = append(modified,'`')
					openbrack = false
				} else {
					modified = append(modified,'`')
				}
				
				pos++
				continue
			} else {
				modified = append(modified,'*')
			}
			
		default:
			modified = append(modified,runestring[pos])
		}
	}

	return string(modified)
}

//**************************************************************

func SplitSentences(para string) []string {

	// bounding sizes in characters
	const min_sentence = 20
	const min_paragraph = 100

	para = CleanText(para)

	var sentences []string
	var dlevel, slevel int

	var extract []rune

	if len(para) < min_paragraph {
		sentences = append(sentences,SanitizeSentence(extract))
		return sentences
	} 

	// First look for matching pairs of quotes
	
	for _,rn := range para {

		extract = append(extract,rn)

		switch rn {
			
		case '\'':
			if slevel == 0 {
				slevel++
			} else {
				slevel--
			}
				
		case '"':
			if dlevel == 0 {
				dlevel++
			} else {
				dlevel--
			}
 			
		case '!','.','。','?':
			if slevel == 0 && dlevel == 0 && len(extract) > min_sentence {
				sentences = append(sentences,SanitizeSentence(extract))
				extract = nil
				}
		}
	}

	// remainder
	
	if len(extract) > 0 {
		sentences = append(sentences,SanitizeSentence(extract))
	}

	return sentences
}

//**************************************************************

func SanitizeSentence(extract []rune) string {

	// In dialogue, some paragraphs may begin with a quote, which
	// does not end in the same paragraph, causing trouble
	
	if len(extract) > 1 && extract[0] == '"' && extract[1] == ' ' {
		for i := 1; i < len(extract); i++ {
			if extract[i] != ' ' && extract[i] != '\t' {
				extract = extract[i:]
				break
			}
		}
	}

	s := string(extract)
	
	if strings.Count(s,"\"") % 2 != 0 {
		s += "\""
	}
	
	return s
}

//**************************************************************

func SanitizeToken(s string) string {

	s = strings.ReplaceAll(s,"***","")
	s = strings.ReplaceAll(s,"`","")

	return s
}

//**************************************************************

func CleanText(s string) string {

	s = strings.TrimSpace(s)
	
	// Start by stripping HTML / XML tags before para-split
	// if they haven't been removed already

	if strings.Contains(s,"\"") {
		if strings.Contains(s,"'") {
			strings.ReplaceAll(s,"\"","_")
		} else {
			s = "'"+s+"'"
		}
	}
	
	m := regexp.MustCompile("<[^>]+>") 
	s = m.ReplaceAllString(s,"") 

	// Weird English abbrev
	s = strings.Replace(s,"[","",-1) 
	s = strings.Replace(s,"]","",-1)

	s = strings.Replace(s,"Mr.","Mr",-1)
	s = strings.Replace(s,"Mrs.","Mrs",-1) 
	s = strings.Replace(s,"Dr.","Dr",-1)
	s = strings.Replace(s,"Ms.","Ms",-1)
	s = strings.ReplaceAll(s,"$","CURRENCY-DOLLAR")
	s = strings.ReplaceAll(s,"\t"," ")
	lquote := fmt.Sprintf("%c",NON_ASCII_LQUOTE)
	s = strings.ReplaceAll(s,lquote,"\"")
	rquote := fmt.Sprintf("%c",NON_ASCII_RQUOTE)
	s = strings.ReplaceAll(s,rquote,"\"")
	s = strings.ReplaceAll(s,"`","'")
	s = strings.ReplaceAll(s,"’","'")
	s = strings.ReplaceAll(s,"(","[")
	s = strings.ReplaceAll(s,")","]")
	s = strings.TrimSpace(s)
	return s
}

//**************************************************************

func CleanNewLines(para string) string {

	para = strings.ReplaceAll(para,"\n"," ")
	para = strings.ReplaceAll(para," &","\\&")
	para = strings.ReplaceAll(para," |","\\|")
	para = strings.TrimSpace(para)
	return para
}

//**************************************************************

func AssemblePBSF(sections []AuxTxtMDSentence)  ([][]Sentence,int) {
	
	var pbsf [][]Sentence
	var count int

	for _,sec := range sections {
		
		sentences := SplitSentences(sec.Content)

		var cleaned []Sentence
		
		for s := range sentences {

			const min_size = 4
			
			if len(sentences[s]) < min_size {
				continue
			}

			var this Sentence
			this.S = sentences[s]
			this.Title = sec.Title

			count++
			
			frags := SplitPunctuationText(sentences[s])
			
			for f := range frags {
				if len(frags[f]) > min_size {
					_, name := DeNumberItem(frags[f])
					frags[f] = SanitizeToken(name)
				}

				content := strings.TrimSpace(frags[f])
				if len(content) > min_size {
					this.Frags = append(this.Frags,content)
				}
			}

			cleaned = append(cleaned,this)
		}

		if len(cleaned) > 0 {
			pbsf = append(pbsf,cleaned)
		}
	}
	
	return pbsf,count
}

//*****************************************************************

func DeNumberItem(text string) (string,string) {

	var number,name string

	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text,"\n"," ")
	name = text
	number = "unnumbered"
	
	if strings.Contains(text," ") {

		first := strings.Split(text," ")[0]

		if LooksLikeNumber(first) {
			name = CleanText(text[len(first):])
			number = strings.Trim(first,".")
		}
	}

	return number,name
}

//**************************************************************

func SplitPunctuationText(s string) []string {

	return SplitPunctuationTextWork(s,false)
}

//**************************************************************

func SplitPunctuationTextWork(s string,allow_small bool) []string {

	// first split sentence on intentional separators

	var subfrags []string

	frags := CountParens(s)

	for f := 0; f < len(frags); f++ {

		contents,hasparen := UnParen(frags[f])

		var sfrags []string

		if hasparen {
			// contiguous parenthesis
			subfrags = append(subfrags,frags[f])
			// and fractionated contents (recurse)
			sfrags = SplitPunctuationTextWork(contents,allow_small)
			sfrags = nil // count but don't repeat
		} else {
			re := regexp.MustCompile("([\"—“”!?,:;—]+[ \n])")
			sfrags = re.Split(contents, -1)
		}

		for sf := range sfrags {
			sfrags[sf] = strings.TrimSpace(sfrags[sf])
			
			if allow_small || len(sfrags[sf]) > 1 {
				subfrags = append(subfrags,sfrags[sf])
			}
		}
	}

	// handle parentheses first as a single fragment because this could mean un-accenting

	// now split on any punctuation that's not a hyphen
	
	return subfrags
}

//**************************************************************

func UnParen(s string) (string,bool) {

	var counter byte = ' '

	switch s[0] {
	case '(':
		counter = ')'
	case '[':
		counter = ']'
	case '{':
		counter = '}'
	}

	if counter != ' ' {
		if s[len(s)-1] == counter {
			trimmed := strings.TrimSpace(s[1:len(s)-1])
			return trimmed,true
		}
	}
	return strings.TrimSpace(s),false
}

//**************************************************************

func CountParens(s string) []string {

	var text = []rune(strings.TrimSpace(s))

	var match rune = ' '
	var count = make(map[rune]int)

	var subfrags []string
	var fragstart int = 0

	for i := 0; i < len(text); i++ {

		switch text[i] {
		case '(':
			count[')']++
			if match == ' ' {
				match = ')'
				frag := strings.TrimSpace(string(text[fragstart:i]))
				fragstart = i
				if len(frag) > 0 {
					subfrags = append(subfrags,frag)
				}
			}
		case '[':
			count[']']++
			if match == ' ' {
				match = ']'
				frag := strings.TrimSpace(string(text[fragstart:i]))
				fragstart = i
				if len(frag) > 0 {
					subfrags = append(subfrags,frag)
				}
			}
		case '{':
			count['}']++
			if match == ' ' {
				match = '}'
				frag := strings.TrimSpace(string(text[fragstart:i]))
				fragstart = i
				if len(frag) > 0 {
					subfrags = append(subfrags,frag)
				}
			}

			// end

		case ')',']','}':
			count[text[i]]--
			if count[match] == 0 {
				frag := text[fragstart:i+1]
				fragstart = i+1
				subfrags = append(subfrags,string(frag))
			}
		}

	}

	lastfrag := strings.TrimSpace(string(text[fragstart:len(text)]))

	if len(lastfrag) > 0 {
		subfrags = append(subfrags,string(lastfrag))
	}

	// Ignore unbalanced parentheses, because it's unclear why in natural language

	return subfrags
}

//**************************************************************

func Fractionate(frag string,L int,frequency [N_GRAM_MAX]map[string]float64,min int) [N_GRAM_MAX][]string {

	// A round robin cyclic buffer for taking fragments and extracting
	// n-ngrams of 1,2,3,4,5,6 words separateed by whitespace, passing

	var rrbuffer [N_GRAM_MAX][]string
	var change_set [N_GRAM_MAX][]string

	words := strings.Split(frag," ")

	for w := range words {
		rrbuffer,change_set = NextWord(words[w],rrbuffer)
	}

	return change_set
}

//**************************************************************

func AssessStaticIntent(frag string,L int,frequency [N_GRAM_MAX]map[string]float64,min int) float64 {

	// A round robin cyclic buffer for taking fragments and extracting
	// n-ngrams of 1,2,3,4,5,6 words separateed by whitespace, passing

	var change_set [N_GRAM_MAX][]string
	var rrbuffer [N_GRAM_MAX][]string
	var score float64

	words := strings.Split(frag," ")

	for w := range words {

		rrbuffer,change_set = NextWord(words[w],rrbuffer)

		for n := min; n < N_GRAM_MAX; n++ {
			for ng := range change_set[n] {
				ngram := change_set[n][ng]
				score += StaticIntentionality(L,ngram,STM_NGRAM_FREQ[n][ngram])
			}
		}
	}

	return score
}

//**************************************************************

func AssessStaticTextAnomalies(L int,frequencies [N_GRAM_MAX]map[string]float64,locations [N_GRAM_MAX]map[string][]int) ([N_GRAM_MAX][]TextRank,[N_GRAM_MAX][]TextRank) {

	// Try to split a text into anomalous/ambient i.e. intentional + contextual  parts

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var anomalous [N_GRAM_MAX][]TextRank
	var ambient [N_GRAM_MAX][]TextRank

	for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {

		for ngram := range STM_NGRAM_LOCA[n] {

			var ns TextRank
			ns.Significance = AssessStaticIntent(ngram,L,STM_NGRAM_FREQ,N_GRAM_MIN)
			ns.Fragment = ngram

			if IntentionalNgram(n,ngram,L,coherence_length) {
				anomalous[n] = append(anomalous[n],ns)
			} else {
				ambient[n] = append(ambient[n],ns)
			}
		}
		
		sort.Slice(anomalous[n], func(i, j int) bool {
			return anomalous[n][i].Significance > anomalous[n][j].Significance
		})

		sort.Slice(ambient[n], func(i, j int) bool {
			return ambient[n][i].Significance > ambient[n][j].Significance
		})
	}

	var intent [N_GRAM_MAX][]TextRank
	var context [N_GRAM_MAX][]TextRank
	var max_intentional = [N_GRAM_MAX]int{0,0,DUNBAR_150,DUNBAR_150,DUNBAR_30,DUNBAR_15}

	for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {

		for i := 0; i < max_intentional[n] && i < len(anomalous[n]); i++ {
			intent[n] = append(intent[n],anomalous[n][i])
		}

		for i := 0; i < max_intentional[n] && i < len(ambient[n]); i++ {
			context[n] = append(context[n],ambient[n][i])
		}
	}

	return intent,context
}

//**************************************************************

func IntentionalNgram(n int,ngram string,L int,coherence_length int) bool {

	// If short file, everything is probably significant

	if n == 1 {
		return false 
	}

	if L < coherence_length {
		return true
	}

	occurrences,minr,maxr := IntervalRadius(n,ngram)

	// if too few occurrences, no difference between max and min delta

	if occurrences < 2 {
		return true
	}

	// the distribution of intraspacings is broad, so not just a regular pattern

	return maxr > minr + coherence_length
}

//**************************************************************

func IntervalRadius(n int, ngram string) (int,int,int) {

	// find minimax distances between n-grams (in sentences)

	occurrences := len(STM_NGRAM_LOCA[n][ngram])
	var dl int = 0
	var dlmin int = 99
	var dlmax int = 0

	// Find the width of the intraspacing distribution

	for occ := 0; occ < occurrences; occ++ {

		d := STM_NGRAM_LOCA[n][ngram][occ]
		delta := d - dl
		dl = d
		
		if dl == 0 {
			continue
		}
		
		if dl > dlmax {
			dlmax = delta
		}
		
		if dl < dlmin {
			dlmin = delta
		}
	}

	return occurrences,dlmin,dlmax
}

//**************************************************************

func AssessTextCoherentCoactivation(L int,ngram_loc [N_GRAM_MAX]map[string][]int) ([N_GRAM_MAX]map[string]int,[N_GRAM_MAX]map[string]int,int) {

	// In this global assessment of coherence intervals, we separate each into text that is unique (intentional)
	// and fragments that are repeated in any other interval, so this is an extreme view. Compare to fast/slow method
	// below

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var overlap [N_GRAM_MAX]map[string]int
	var condensate [N_GRAM_MAX]map[string]int

	C,partitions := CoherenceSet(ngram_loc,L,coherence_length)

	for n := 1; n < N_GRAM_MAX; n++ {

		overlap[n] = make(map[string]int)
		condensate[n] = make(map[string]int)

		// now run through linearly and split nearest neighbours

		// very short excerpts,there is nothing we can do in a single coherence set
		if partitions < 2 {
			for ngram := range C[n][0] {
				overlap[n][ngram]++
			}
		// multiple coherence zones
		} else {
			for pi := 0; pi < len(C[n]); pi++ {
				for pj := pi+1; pj < len(C[n]); pj++ {
					for ngram := range C[n][pi] {
						if C[n][pi][ngram] > 0 && C[n][pj][ngram] > 0 {
							// ambients
							delete(condensate[n],ngram)
							overlap[n][ngram]++
						} else {
							// unique things here
							_,ambient := overlap[n][ngram]
							if !ambient {
								condensate[n][ngram]++
							}
						}
					}
				}
			}
		}
	}
	return overlap,condensate,partitions
}

//**************************************************************

func AssessTextFastSlow(L int,ngram_loc [N_GRAM_MAX]map[string][]int) ([N_GRAM_MAX][]map[string]int,[N_GRAM_MAX][]map[string]int,int) {

	// Use a running evaluation of context intervals to separate ngrams that are varying quickly (intentional)
	// from those changing slowly (context). For each region, what is different from the last is fast and what
	// remains the same as last is slow. This is remarkably effective and quick to calculate.

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var slow [N_GRAM_MAX][]map[string]int
	var fast [N_GRAM_MAX][]map[string]int

	C,partitions := CoherenceSet(ngram_loc,L,coherence_length)

	for n := 1; n < N_GRAM_MAX; n++ {

		slow[n] = make([]map[string]int,partitions)
		fast[n] = make([]map[string]int,partitions)

		// now run through linearly and split nearest neighbours

		// very short excerpts,there is nothing we can do in a single coherence set

		if partitions < 2 {

			slow[n][0] = make(map[string]int)
			fast[n][0] = make(map[string]int)

			for ngram := range C[n][0] {
				fast[n][0][ngram]++
			}

		// multiple coherence zones

		} else {
			for p := 1; p < partitions; p++ {

				slow[n][p-1] = make(map[string]int)
				fast[n][p-1] = make(map[string]int)

				for ngram := range C[n][p-1] {

					if C[n][p][ngram] > 0 && C[n][p-1][ngram] > 0 {
						// ambients
						slow[n][p-1][ngram]++
					} else {
						// unique things here
						fast[n][p-1][ngram]++
					}
				}
			}
		}
	}

	return slow,fast,partitions
}

//**************************************************************

func CoherenceSet(ngram_loc [N_GRAM_MAX]map[string][]int, L,coherence_length int) ([N_GRAM_MAX][]map[string]int,int) {

	var C [N_GRAM_MAX][]map[string]int

	partitions := L/coherence_length + 1

	for n := 1; n < N_GRAM_MAX; n++ {
		
		C[n] = make([]map[string]int,partitions)

		for p := 0; p < partitions; p++ {
			C[n][p] = make(map[string]int)
		}

		for ngram := range ngram_loc[n] {
			
			// commute indices and expand to a sparse representation for simplicity

			for s := range ngram_loc[n][ngram] {
				p := ngram_loc[n][ngram][s] / coherence_length
				if p < partitions {
					C[n][p][ngram]++
				}
			}
		}
	}

	return C,partitions
}

//**************************************************************

func NextWord(frag string,rrbuffer [N_GRAM_MAX][]string) ([N_GRAM_MAX][]string,[N_GRAM_MAX][]string) {

	// Word by word, we form a superposition of scores from n-grams of different lengths
	// as a simple sum. This means lower lengths will dominate as there are more of them
	// so we define intentionality proportional to the length also as compensation

	var change_set [N_GRAM_MAX][]string

	for n := 1; n < N_GRAM_MAX; n++ {
		
		// Pop from round-robin

		if (len(rrbuffer[n]) > n-1) {
			rrbuffer[n] = rrbuffer[n][1:n]
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

			key = CleanNgram(key)

			if ExcludedByBindings(CleanNgram(rrbuffer[n][0]),key,CleanNgram(rrbuffer[n][n-1])) {
				continue
			}

			change_set[n] = append(change_set[n],key)
		}
	}

	frag = CleanNgram(frag)

	if N_GRAM_MIN <= 1 && !ExcludedByBindings(frag,frag,frag) {
		change_set[1] = append(change_set[1],frag)
	}

	return rrbuffer,change_set
}

//**************************************************************

func CleanNgram(s string) string {

	re := regexp.MustCompile("[-][-][-].*")
	s = re.ReplaceAllString(s,"")
	re = regexp.MustCompile("[\"—“”!?`,.:;—()_]+")
	s = re.ReplaceAllString(s,"")
	s = strings.Replace(s,"  "," ",-1)
	s = strings.Trim(s,"-")
	s = strings.Trim(s,"'")

	return strings.ToLower(s)
}

//*******************************************************************

func GenerateOutput(fp *os.File,filealias string,selection []TextRank,L int, percentage float64,anom_by_part[][]string,ambi_by_part[][]string,all_anom[]string,all_ambi[]string,with_title bool) {

	// See AddMandatory() in N4L.go for reserved names (TBD, collect these one day as const)

	fmt.Fprintf(fp,"\n### BEGIN main file: %s \n",filealias)
	
	if with_title {
		fmt.Fprintf(fp,"\n-%s\n",filealias)
	}
	
	// Lookup tables of contents from Markdown

	CompileTabular(fp,filealias,TABLES)
	CompileProperNames(fp,filealias,selection,TABLES)
	CompileTOC(fp,filealias,SECTIONS)

	
	fmt.Fprintf(fp,"\n### BEGIN content from: %s \n",filealias)

	// 

	var collected_fragments = make(map[string][]string)
	var partcheck = make(map[string]bool)
	var parts []string
	var context, last_context string
	var last_part string
	var already = make(map[string]bool)
	
	const minimum_length = 15
	
	for i := range selection {

		// Skip spurious
		
		if len(selection[i].Fragment) < minimum_length {
			continue
		}

		context,last_context = CheckNewContext(fp,last_context,ambi_by_part[selection[i].Partition],filealias)

		part := GetSectionTitle(selection[i],context,filealias)

		// Switch section, new sequence
		
		if part != last_part {
			fmt.Fprintf(fp,"\n-:: _sequence_ ::")
			
			fmt.Fprintf(fp,"\n######## Section name and subsequence ########")
			fmt.Fprintf(fp,"\n %s\n\n",part)
			fmt.Fprintf(fp,"\n :: %s, _sequence_ ::\n",context)
			fmt.Fprintf(fp,"#######################################\n")
			
			last_part = part
		
			if !partcheck[part] {
				parts = append(parts,part)
				partcheck[part] = true
			}
		}

		// !!!!!  Output an event - subordinate to a part/section  !!!!!

		OutputDocumentEvent(fp,selection[i],filealias,part)
		
		AddIntentionalContext(collected_fragments,part,anom_by_part[selection[i].Partition],already)
	}

	fmt.Fprintf(fp,"\n-:: _sequence_ ::")
	fmt.Fprintf(fp,"\n### END content from: %s\n",filealias)

	// ********************************************************
	// Remainder is purely for debugging and analysis .....
	// ********************************************************

	fmt.Fprintf(fp,"\n# Final fraction %.2f of requested %.2f\n",float64(len(selection)*100)/float64(L),percentage)

	WriteSampleSelections(fp,selection,L)

	// Add in the section/part fragments to the context/toc/index

	WriteCollectedContext(fp,collected_fragments)
	
	// Document the parts

	fmt.Fprintf(fp,"\n #\n # SUMMARY OF CONTEXT AND SIGNIFICANT FRAGMENTS\n #\n")

	fmt.Fprintf(fp,"\n :: themes and topics you might want to annotate/replace ::\n")

	for p := range parts {
		
		container,exists := DOC_DIRECTORY[parts[p]]

		if exists {
			fmt.Fprintf(fp,"\n %s (%s) %s \n",parts[p],INV_CONT_FRAG_IN_S,container)
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

	fmt.Printf("Final fraction %.2f of requested %.2f sampled\n",float64(len(selection)*100)/float64(L),percentage)

}

//*******************************************************************

func OutputDocumentEvent(fp *os.File, selection TextRank,alias string,part_title string) {

	event := fmt.Sprintf("\n %s [%s]",SanitizeParen(selection.Fragment),alias)
	
	fmt.Fprintf(fp,"%s (%s) %s \n",event,INV_CONT_NAME_S,part_title)

	// Now add cross referencing anchors to proper names
	
	names := SearchStringForProperNames(event)
	
	for _,n := range names {
		fmt.Fprintf(fp,"    \"  (refersto) %s\n",n)
	}
}

//*******************************************************************

func WriteSampleSelections(fp *os.File, selection []TextRank, L int) {

	fmt.Fprintf(fp,"\n# Selected %d samples of %d: ",len(selection),L)

	for i := range selection {
		fmt.Fprintf(fp,"%d ",selection[i].Order)
		}

	fmt.Fprintf(fp,"\n#\n")
}

//*******************************************************************

func WriteCollectedContext(fp *os.File, collected_fragments map[string][]string) {

	fmt.Fprintf(fp,"\n #\n # Concepts by part / region \n #\n")

	for key := range collected_fragments {

		fmt.Fprintf(fp,"\n\n %s\n",key)
		for _,s := range collected_fragments[key] {
			fmt.Fprintf(fp,"              \" (%s) %s\n",CONT_FRAG_S,s)
		}
	}

}

//*******************************************************************

func AddIntentionalContext(collected map[string][]string,key string,ctx []string,already map[string]bool) {

	for w := 0; w < len(ctx); w++ {

		if !LooksLikeContext(ctx[w]) {
			continue
		}

		if !already[ctx[w]] {
			collected[key] = append(collected[key],ctx[w])
			already[ctx[w]] = true
		}
	}

	collected[key] = append(collected[key],"Table of Contents")
}

// **************************************************************************

func GetSectionTitle(selection TextRank,context string,filealias string) string {

	// The title is originally set in SplitIntoSectionsOfSentences()
	
	if selection.Title != "" {
		return selection.Title
	}

	// Make up a stand-in title

	return MissingTitle(selection.Partition,filealias,context)
}

//*******************************************************************

func MissingTitle(p int,filealias string,context string) string {

	var basename string
	
	if p == 0 {
		basename = fmt.Sprintf("Introduction to %s",filealias)
	} else {
		basename = fmt.Sprintf("part %d of %s",p,filealias)
	}
	
	// include ambient context in the section name

	if len(context) > 0 {
		return fmt.Sprintf("%s: %s",basename,context)
	} else {
		return basename
	}
}

//*****************************************************************
// Build ancillary structures
//*****************************************************************

func CompileTOC(fp *os.File, filealias string, sections []AuxTxtMDSentence) {

	// map to n4l
	
	var done = make(map[string]bool)

	fmt.Fprintf(fp,"\n # BEG table of contents ############################\n")
	fmt.Fprintf(fp,"\n :: _sequence_, Table of Contents - %s :: \n",filealias)

	fmt.Fprintf(fp,"\n Table of Contents - %s \n",filealias)
	
	for _,sec := range sections {

		title := strings.ToLower(sec.Title)
		
		if !done[title] {
			
			if LooksLikeProperName(sec.Title) || strings.Contains(sec.Title," ") {

				done[title] = true

				// Avoid double annotation
				
				if strings.Contains(sec.Title,filealias) {
					fmt.Fprintf(fp,"\n   \"  (%s) %s\n",CONT_NAME_L,sec.Title)
				} else {
					fmt.Fprintf(fp,"\n   \"  (%s) %s [%s]\n",CONT_NAME_L,sec.Title,filealias)
				}
			}
		}
	}

	fmt.Fprintf(fp,"\n -:: _sequence_ :: \n")
	fmt.Fprintf(fp,"\n # END table of contents for %s ############################\n",filealias)
}

//*****************************************************************

func ListProperNames(fp *os.File) {

	fmt.Println("\nCompiling proper names ....\n\n")

	fmt.Fprintf(fp,"\n\n :: proper names :: \n\n")
	
	for name := range PROPER_NAME {

		if PROPER_NAME[name] != "" {
			fmt.Fprintf(fp,"\n %s (is explained by) %s \n",name,PROPER_NAME[name])
		} else {
			fmt.Fprintf(fp,"\n %s\n",name)
		}
		
	}
}


//
// text_fractionation.go
//

