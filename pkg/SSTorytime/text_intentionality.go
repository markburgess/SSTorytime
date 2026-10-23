// *********************************************************************
//
// text_intentionality.go
//
// *********************************************************************

package SSTorytime

import (
	"fmt"
	"strings"
	"sort"
	"math"
	"time"
	_ "github.com/lib/pq"
)

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

//**************************************************************

func AnalyzeFractionIntent(pbsf [][]Sentence) {

	var count int
	
	for p := range pbsf {
		for s := range pbsf[p] {
			count++
			for f := range pbsf[p][s].Frags {

				change_set := Fractionate(pbsf[p][s].Frags[f],count,STM_NGRAM_FREQ,N_GRAM_MIN)

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
}

//******************************************************************

func ContextIntentAnalysis(spectrum map[string]int) ([]string,[]string) {

        // Used in table of contents.
	// Split a list of contexts into mutex intentional and ambient

	var intentional []string
	var ambient []string
	
	const intent_limit = 2  // policy from research

	// The spectrum is the number of times a DNA fragment appears in N4L classifications
	
	for f := range spectrum {

		if spectrum[f] < intent_limit {
			intentional = append(intentional,f)
		} else {
			ambient = append(ambient,f)
		}
	}

	return intentional,ambient
}


//*******************************************************************

func SelectByRunningIntent(psf [][]Sentence,L int,percentage float64) []TextRank {

	// Rank sentences

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var sentences []TextRank
	var sentence_counter int

	for p := range psf {

		for s := range psf[p] {

			score := 0.0

			for f := 0; f < len(psf[p][s].Frags); f++ {

				score += RunningIntentionality(sentence_counter,psf[p][s].Frags[f])
			}

			var this TextRank
			this.Fragment = psf[p][s].S
			this.Title = psf[p][s].Title
			this.Significance = score
			this.Order = sentence_counter
			this.Partition = sentence_counter / coherence_length
			sentences = append(sentences,this)
			sentence_counter++
		}
	}

	skimmed := OrderAndRank(sentences,percentage)

	return skimmed
}

// ***************************************************

func SelectByStaticIntent(psf [][]Sentence,L int,percentage float64) []TextRank {

	// Rank sentences

	const coherence_length = DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var sentences []TextRank
	var sentence_counter int

	for p := range psf {

		for s := range psf[p] {

			score := 0.0

			for f := 0; f < len(psf[p][s].Frags); f++ {

				score += AssessStaticIntent(psf[p][s].Frags[f],L,STM_NGRAM_FREQ,1)
			}

			var this TextRank
			this.Fragment = psf[p][s].S
			this.Title = psf[p][s].Title
			this.Significance = score
			this.Order = sentence_counter
			this.Partition = sentence_counter / coherence_length
			sentences = append(sentences,this)
			sentence_counter++
		}
	}

	skimmed := OrderAndRank(sentences,percentage)

	return skimmed
}

//*********************************************************************************

func OrderAndRank(sentences []TextRank,percentage float64) []TextRank {

	var selections []TextRank
	
	// Order by intentionality first to skim cream

	sort.Slice(sentences, func(i, j int) bool {
		return sentences[i].Significance > sentences[j].Significance
	})

	// Measure relative threshold for percentage of document
	// the lower the threshold, the lower the significance of the document

	threshold := percentage / 100.0

	limit := int(threshold * float64(len(sentences)))

	for i := 0; i < limit; i++ {
		selections = append(selections,sentences[i])
	}

	// Order by line number again to restore causal order

	sort.Slice(selections, func(i, j int) bool {
		return selections[i].Order < selections[j].Order
	})

	return selections
}

//*********************************************************************************

func MergeSelections(one []TextRank,two []TextRank) []TextRank{

	var merge []TextRank
	var already_selected = make(map[int]bool)

	for i := range one {
		merge = append(merge,one[i])
		already_selected[one[i].Order] = true
	}

	for i := range two {
		if !already_selected[two[i].Order] {
			merge = append(merge,two[i])
		}
	}

	// Order by line number again to restore causal order

	sort.Slice(merge, func(i, j int) bool {
		return merge[i].Order < merge[j].Order
	})

	return merge
}

// **************************************************************************

func GetChaptersByChapContext(sst PoSST,chap string,cn []string,limit int) map[string][]string {

	qstr := ""
	chap_col := ""

	chap = strings.Trim(chap,"\"")

	if chap != "any" && chap != "" {

		remove_chap_accents,chap_stripped := IsBracketedSearchTerm(chap)

		if remove_chap_accents {
			chap_search := "%"+chap_stripped+"%"
			chap_col = fmt.Sprintf("AND lower(unaccent(chap)) LIKE lower('%s')",chap_search)
		} else {
			chap_search := "%"+chap+"%"
			chap_col = fmt.Sprintf("AND lower(chap) LIKE lower('%s')",chap_search)
		}
	}

	if chap == "TableOfContents" {
		chap_col = ""
	}

	_,cn_stripped := IsBracketedSearchList(cn)
	context := FormatSQLStringArray(cn_stripped)

	qstr = fmt.Sprintf("SELECT DISTINCT chap,ctx FROM PageMap WHERE match_context(ctx,%s) %s ORDER BY Chap",context,chap_col)

	row, err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("QUERY GetChaptersByChapContext Failed",err,qstr)
	}

	var rchap string
	var rcontext ContextPtr
	var toc = make(map[string][]string)

	if row != nil {
		for row.Next() {		
			err = row.Scan(&rchap,&rcontext)

			// Each chapter can be a comma separated list

			chps := SplitChapters(rchap)

			for c := 0; c < len(chps); c++ {

				if len(toc) == limit {
					row.Close()
					return toc
				}

				rc := chps[c]

				cn := strings.Split(GetContext(&sst,rcontext),",")
				ctx_grp := ""

				for s := 0; s < len(cn); s++ {
					ctx_grp += cn[s]
					if s < len(cn)-1 {
						ctx_grp += ", "
					}
				}

				if len(ctx_grp) > 0 {
					toc[rc] = append(toc[rc],ctx_grp)
				}
			}
		}

		row.Close()
	}

	return toc
}

// *********************************************************************

func UpdateSTMContext(sst *PoSST,ambient,key string,now int64,params SearchParameters) string {

	var context []string

	if params.Sequence || params.From != nil || params.To != nil {
		// path / cone are intended
		context = append(context,params.Name...)
		context = append(context,params.From...)
		context = append(context,params.To...)
		return AddContext(sst,ambient,key,now,context)
	} else {
		// ongoing / adhoc are ambient
		context = append(context,params.Name...)

		for _,ct := range params.Context {
			if ct != "" {
				context = append(context,ct)
			}
		}

		if params.Chapter != "" {
			context = append(context,"Chapter:"+params.Chapter)
		}

		return AddContext(sst,ambient,key,now,context)
	}

	return ""
}

// *********************************************************************

func AddContext(sst *PoSST,ambient,key string,now int64,tokens []string) string {

	for t := range tokens {

		token := tokens[t]

		if len(token) == 0 || token == "%%" {
			continue
		}

		// Check for direct NPtr click, watch out for long text

		if token[0] == '(' {
			var nptr NodePtr
			fmt.Sscanf(token,"(%d,%d)",&nptr.Class,&nptr.CPtr)

			if nptr.Class > 0 {
				node := GetDBNodeByNodePtr(sst,nptr)
				if node.L > TEXT_SIZE_LIMIT + 10 {
					token = node.S[0:TEXT_SIZE_LIMIT] + "..."
				} else {
					token = node.S
				}
			} else {
				continue
			}
		}
		CommitContextToken(token,now,ambient)
	}

	var format = make(map[string]int)

	for fr := range STM_AMB_FRAG {

		if STM_AMB_FRAG[fr].Delta > FORGOTTEN {
			delete(STM_AMB_FRAG,fr)
			continue
		} 

		format[fr]++
	}

	for fr := range STM_INT_FRAG {

		if STM_INT_FRAG[fr].Delta > FORGOTTEN {
			delete(STM_INT_FRAG,fr)
			continue
		} 

		format[fr]++
	}

	full_context := List2String(Map2List(format))

	return full_context
}

// *********************************************************************

func CommitContextToken(token string,now int64,key string) {
	
	var last,obs History
	
	// Check if already known ambient
	last,already := STM_AMB_FRAG[token]
	
	// if not, then check if already seen
	if !already {
		last,already = STM_INT_FRAG[token]
	}
	
	if !already {
		last.Last = now
	}
	
	obs.Freq = last.Freq + 1
	obs.Last = now
	obs.Time = key
	obs.Delta = now - last.Last
	
	if obs.Freq > 1 {
		pr,okey := DoNowt(time.Unix(last.Last,0))
		fmt.Printf("    - last saw \"%s\" at %s (%s)\n",token,pr,okey)
	}
	
	if already {
		delete(STM_INT_FRAG,token)
		STM_AMB_FRAG[token] = obs
	} else {
		STM_INT_FRAG[token] = obs
	}
}

// **************************************************************************

func IntersectContextParts(context_clusters []string) ([]string,[][]int)  {

	// return a weighted upper triangular matrix of overlaps between frags,
	// and an idempotent list of fragments

	var idemp = make(map[string]int)
	var cluster_list []string

	for s := range context_clusters {
		idemp[context_clusters[s]]++
	}

	for each_unique_cluster := range idemp {
		cluster_list = append(cluster_list,each_unique_cluster)
	}

	sort.Strings(cluster_list)

	var adj [][]int

	for ci := 0; ci < len(cluster_list); ci++ {

		var row []int

		for cj := ci+1; cj < len(cluster_list); cj++ {			
			s,_ := DiffClusters(cluster_list[ci],cluster_list[cj])
			row = append(row,len(s))
		}

		adj = append(adj,row)
	}

	return cluster_list,adj
}

// **************************************************************************
// These functions are about text fractionation of the context strings
// which is similar to text2N4L scanning but applied to lists of phrases
// on a much smaller scale. Still looking for "mass spectrum" of fragments ..
// **************************************************************************

func DiffClusters(l1,l2 string) (string,string) {

	// The fragments arrive as comma separated strings that are
        // already composed or ordered n-grams

	spectrum1 := strings.Split(l1,", ")
	spectrum2 := strings.Split(l2,", ")

	// Get orderless idempotent directory of all 1-grams

	m1 := List2Map(spectrum1)
	m2 := List2Map(spectrum2)

	// split the lists into words into directories for common and individual ngrams

	return OverlapMatrix(m1,m2)
}

// **************************************************************************

func OverlapMatrix(m1,m2 map[string]int) (string,string) {

	var common = make(map[string]int)
	var separate = make(map[string]int)

	// sieve shared / individual parts

	for ng := range m1 {
		if m2[ng] > 0 {
			common[ng]++
		} else {
			separate[ng]++
		}
	}

	for ng := range m2 {
		if m1[ng] > 0 {
			delete(separate,ng)
			common[ng]++
		} else {
			_,exists := common[ng]
			if  !exists {
				separate[ng]++
			}
		}
	}

	return List2String(Map2List(common)),List2String(Map2List(separate))
}

// **************************************************************************

func GetContextTokenFrequencies(fraglist []string) map[string]int {

	// Count up the occurrences of different context fragments
	
	var spectrum = make(map[string]int)

	for l := range fraglist {
		fragments := strings.Split(fraglist[l],", ")
		partial := List2Map(fragments)

		// Merge all strands

		for f := range partial {
			spectrum[f] += partial[f]
		}
	}

	return spectrum
}

//
// text_intentionality.go
//
