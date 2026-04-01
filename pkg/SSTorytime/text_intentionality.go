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
	"time"
	_ "github.com/lib/pq"

)

// *********************************************************************

func UpdateSTMContext(sst PoSST,ambient,key string,now int64,params SearchParameters) string {

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

func AddContext(sst PoSST,ambient,key string,now int64,tokens []string) string {

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
				if node.L < TEXT_SIZE_LIMIT {
					token = node.S
				} else {
					token = node.S[0:TEXT_SIZE_LIMIT] + "..."
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

func IntersectContextParts(context_clusters []string) (int,[]string,[][]int)  {

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

	return len(cluster_list),cluster_list,adj
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
