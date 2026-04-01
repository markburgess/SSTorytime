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
	"math"
	_ "github.com/lib/pq"

)

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

type TextRank struct {
	Significance float64
	Fragment     string
	Order        int
	Partition    int
}

//**************************************************************

func NewNgramMap() [N_GRAM_MAX]map[string]float64 {

	var thismap [N_GRAM_MAX]map[string]float64

	for i := 1; i < N_GRAM_MAX; i++ {
		thismap[i] = make(map[string]float64)
	}

	return thismap
}

//**************************************************************

func CleanText(s string) string {

	// Start by stripping HTML / XML tags before para-split
	// if they haven't been removed already

	m := regexp.MustCompile("<[^>]*>") 
	s = m.ReplaceAllString(s,":\n") 

	// Weird English abbrev
	s = strings.Replace(s,"[","",-1) 
	s = strings.Replace(s,"]","",-1) 

	// Encode sentence space boundaries and end of sentence markers with a # for later splitting

	/* ellipsis
	m = regexp.MustCompile("([.][.][.])+")  // end of sentence punctuation
	s = m.ReplaceAllString(s,"---")

	m = regexp.MustCompile("[—]+")  // endash
	s = m.ReplaceAllString(s,", ") */

	return s
}

//******************************************************************

func FractionateTextFile(name string) ([][][]string,int) {

	file := ReadTextFile(name)
	proto_text := CleanText(file)
	pbsf := SplitIntoParaSentences(proto_text)

	count := 0

	for p := range pbsf {
		for s := range pbsf[p] {

			count++

			for f := range pbsf[p][s] {

				change_set := Fractionate(pbsf[p][s][f],count,STM_NGRAM_FREQ,N_GRAM_MIN)

				// Update global n-gram frequencies for fragment, and location histories

				for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {
					for ng := range change_set[n] {
						ngram := change_set[n][ng]
						STM_NGRAM_FREQ[n][ngram]++
						STM_NGRAM_LOCA[n][ngram] = append(STM_NGRAM_LOCA[n][ngram],count)
					}
				}
			}
		}
	}
	return pbsf,count
}

//**************************************************************

func SplitIntoParaSentences(file string) [][][]string {

	var pbsf [][][]string

	// first split by paragraph

	paras := strings.Split(file,"\n\n")

	for _,p := range paras {

		p = strings.TrimSpace(p)
		sentences := SplitSentences(p)

		var cleaned [][]string
		
		for s := range sentences {

			// NB, if parentheses contain multiple sentences, this complains, TBD

			frags := SplitPunctuationText(sentences[s])

			var codons []string

			for f := range frags {
				content := strings.TrimSpace(frags[f])
				if len(content) > 2 {			
					codons = append(codons,content)
				}
			}

			if len(codons) > 0 {
				cleaned = append(cleaned,codons)
			}
		}

		if len(cleaned) > 0 {
			pbsf = append(pbsf,cleaned)
		}
	}

	return pbsf
}

//**************************************************************

func SplitSentences(para string) []string {

	var sentences []string
	const small_string = 10

	re := regexp.MustCompile("[?!.。][ \n\t]")
	para = re.ReplaceAllString(para,"$0#")

	sents := strings.Split(para,"#")
	
	var str string

	for i := 0; i < len(sents); i++ {
		
		if i < len(sents)-1 && len(sents[i]) < small_string {
			str += sents[i]
			continue
		}

		str += sents[i]
		str = strings.ReplaceAll(str,"\n"," ")
		sentences = append(sentences,str)
		str = ""
	}
	
	return sentences
}



//**************************************************************

func SplitCommandText(s string) []string {

	return SplitPunctuationTextWork(s,true)
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
				C[n][p][ngram]++
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

//**************************************************************

func ExtractIntentionalTokens(L int,selected []TextRank,Nmin,Nmax int) ([][]string,[][]string,[]string,[]string) {

	// This function examines a fractionation of text for fractions, only for
	// sentences that are selected, and extracts some shared context

	const policy_skim = 15
	const reuse_threshold = 0
	const intent_threshold = 1

	slow,fast,doc_parts := AssessTextFastSlow(L,STM_NGRAM_LOCA)

	var grad_amb [N_GRAM_MAX]map[string]float64
	var grad_oth [N_GRAM_MAX]map[string]float64

	// returns

	var fastparts = make([][]string,doc_parts)
	var slowparts = make([][]string,doc_parts)
	var fastwhole []string
	var slowwhole []string

	for n := 1; n < Nmax; n++ {
		grad_amb[n] = make(map[string]float64)
		grad_oth[n] = make(map[string]float64)
	}

	for p := 0; p < doc_parts; p++ {

		for n := Nmin; n < Nmax; n++ {

			var amb []string
			var other []string

			for ngram := range fast[n][p] {
				if fast[n][p][ngram] > reuse_threshold {
					other = append(other,ngram)
				}
			}

			for ngram := range slow[n][p] {
				if slow[n][p][ngram] > reuse_threshold {
					amb = append(amb,ngram)
				}
			}
			
			// Sort by intentionality

			sort.Slice(amb, func(i, j int) bool {
				ambi :=	StaticIntentionality(L,amb[i],STM_NGRAM_FREQ[n][amb[i]])
				ambj := StaticIntentionality(L,amb[j],STM_NGRAM_FREQ[n][amb[j]])
				return ambi > ambj
			})

			sort.Slice(other, func(i, j int) bool {
				inti := StaticIntentionality(L,other[i],STM_NGRAM_FREQ[n][other[i]])
				intj := StaticIntentionality(L,other[j],STM_NGRAM_FREQ[n][other[j]])
				return inti > intj
			})
			
			for i := 0 ; i < policy_skim && i < len(amb); i++ {
				v := StaticIntentionality(L,amb[i],STM_NGRAM_FREQ[n][amb[i]])
				slowparts[p] = append(slowparts[p],amb[i])
				if v > intent_threshold {
					grad_amb[n][amb[i]] += v
				}
			}
			
			for i := 0 ; i < policy_skim && i < len(other); i++ {
				v := StaticIntentionality(L,other[i],STM_NGRAM_FREQ[n][other[i]])
				fastparts[p] = append(fastparts[p],other[i])
				if v > intent_threshold {
					grad_oth[n][other[i]] += v
				}
			}
		}
	}
	
	// Summary ranking of whole doc, but pick only if selected
	
	for n := Nmin; n < Nmax; n++ {
		
		var amb []string
		var other []string

		for s := range selected {
			for ngram := range grad_amb[n] {
				if !strings.Contains(selected[s].Fragment,ngram) {
					delete(grad_amb[n],ngram)
				}
			}

			for ngram := range grad_oth[n] {
				if !strings.Contains(selected[s].Fragment,ngram) {
					delete(grad_oth[n],ngram)
				}
			}
		}
				
		// there is possible overlap

		for ngram := range grad_oth[n] {
			_,dup := grad_amb[n][ngram]
			if dup {
				continue
			}
			other = append(other,ngram)
		}

		for ngram := range grad_amb[n] {
			amb = append(amb,ngram)
		}

		// Sort by intentionality
		
		sort.Slice(amb, func(i, j int) bool {
			ambi := StaticIntentionality(L,amb[i],STM_NGRAM_FREQ[n][amb[i]])
			ambj := StaticIntentionality(L,amb[j],STM_NGRAM_FREQ[n][amb[j]])
			return ambi > ambj
		})
		sort.Slice(other, func(i, j int) bool {
			inti := StaticIntentionality(L,other[i],STM_NGRAM_FREQ[n][other[i]])
			intj := StaticIntentionality(L,other[j],STM_NGRAM_FREQ[n][other[j]])
			return inti > intj
		})
		
		for i := 0 ; i < policy_skim && i < len(amb); i++ {
			slowwhole = append(slowwhole,amb[i])
		}

		for i := 0 ; i < policy_skim && i < len(other); i++ {
			fastwhole = append(fastwhole,other[i])
		}
	}	

	return fastparts,slowparts,fastwhole,slowwhole
}

//**************************************************************

func RunningIntentionality(t int, frag string) float64 {

	// A round robin cyclic buffer for taking fragments and extracting
	// n-ngrams of 1,2,3,4,5,6 words separateed by whitespace, passing

	var change_set [N_GRAM_MAX][]string
	var rrbuffer [N_GRAM_MAX][]string
	var score float64

	words := strings.Split(frag," ")
	decayrate := float64(DUNBAR_30)

	for w := range words {

		rrbuffer,change_set = NextWord(words[w],rrbuffer)

		for n := N_GRAM_MIN; n < N_GRAM_MAX; n++ {

			for ng := range change_set[n] {
				ngram := change_set[n][ng]
				work := float64(len(ngram))
				lastseen := STM_NGRAM_LAST[n][ngram]

				if lastseen == 0 {
					score = work
				} else {
					score += work * (1 - math.Exp(-float64(t-lastseen)/decayrate))
				}

				STM_NGRAM_LAST[n][ngram] = t
			}
		}
	}

	return score

}

//**************************************************************

func StaticIntentionality(L int, s string, freq float64) float64 {

	// Compute the effective significance of a string s
	// within a document of many sentences. The weighting due to
	// inband learning uses an exponential deprecation based on
	// SST scales (see "leg" meaning).

	work := float64(len(s)) 

	// if this doesn't occur at least 3 times, then why do we care?

	const ignore = 2

	if freq < ignore {
		return 0
	}

	// tempting to measure occurrences relative to total length L in sentences
	// but this is not the relevant scale. Coherence is on a shorter scale
	// set by cognitive limits, not author expansiveness / article scope ...

	phi := freq
	phi_0 := float64(DUNBAR_30) // not float64(L)

	// How often is too often for a concept?
	const rho = 1/30.0 

	crit := phi/phi_0 - rho

	meaning := phi * work / (1.0 + math.Exp(crit))

	return meaning
}



//
// text_fractionation.go
//

