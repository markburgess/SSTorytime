
package main

import (
	"fmt"
	"strings"
	"regexp"
        SST "SSTorytime"
)

//**************************************************************
// BEGIN
//**************************************************************

func main() {

	SST.MemoryInit()
	file := SST.ReadFile("../../examples/example_data/NDA.txt")
	//file := SST.ReadFile("../../examples/example_data/MobyDick.dat")

	pbsf := SplitIntoParaSentences(file)

	for p := range pbsf {
		for s := range pbsf[p] {
			for f:= range pbsf[p][s] {
				fmt.Println("-",pbsf[p][s][f])
			}
			fmt.Println("..........")
		}
		fmt.Println("============")
	}
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

			frags := SST.SplitPunctuationText(sentences[s])

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
