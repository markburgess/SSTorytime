//**************************************************************
//
// text_heuristics.go
//
//**************************************************************

package SSTorytime

import (
	"strings"
)

//**************************************************************

func ExcludedByBindings(firstword,whole,lastword string) bool {

        // This is the extent of grammatical understanding we need to parse the text
	// In principle, it is determined by training, but we can summarize it like this

	// An empirical standalone fragment can't start/end with these words, because they
	// Promise to bind to something else...
	// Rather than looking for semantics, look at spacetime promises only - words that bind strongly
	// to a prior or posterior word.

	// Promise bindings in English only. This domain knowledge saves us a lot of training analysis
	// So how to replace this with something generic?
	
	var forbidden_ending = []string{"but", "and", "the", "or", "a", "an", "its", "it's", "their", "your", "my", "our", "of", "as", "are", "is", "was", "has", "be", "been", "with", "using", "that", "who", "to" ,"no", "not", "because", "at", "but", "yes", "no", "yeah", "yay", "in", "which", "what", "as", "he", "him", "she", "her","they", "all", "I", "my", "they", "from", "for", "then", "any", "however", "its", "it's", "get", "don't", "this", "one", "shall"}

	var forbidden_starter = []string{"its", "it's", "and", "or", "of", "the", "it", "because", "in", "that", "these", "those", "is", "are", "was", "were", "but", "yes", "no", "yeah", "yay", "also", "me", "them", "him", "his", "her", "but", "been", "however", "get", "do", "don't", "soon", "own", "all", "their", "suppose", "for", "said", "shall", "will"}

	if (len(firstword) <= 2) || len(lastword) <= 2 {
		return true
	}

        // Adverbs don't end

        if strings.HasSuffix(lastword,"ly") {
                return true
        }

        if strings.Contains(whole,"--") {
	        return true
	}

	for s := range forbidden_ending {
		if strings.ToLower(lastword) == forbidden_ending[s] {
			return true
		}
	}
	
	for s := range forbidden_starter {
		if strings.ToLower(firstword) == forbidden_starter[s] {
			return true
		}
	}

	return false 
}



//
// text_heuristics.go
//
