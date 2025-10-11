//
// Scan a document and pick out the sentences that are measured to
// be high in "intentionality" or potential knowledge significance
// using two methods: dynamic running and static posthoc assessment
//

package main

import (
	"os"
	"fmt"
	"sort"
	"flag"
	"strings"
        SST "SSTorytime"
)

var TARGET_PERCENT float64 = 50.0

//**************************************************************
// BEGIN
//**************************************************************

func main() {

	const max_class = 100

	input := GetArgs()

	RipFile2File(input,TARGET_PERCENT)

}

//**************************************************************

func GetArgs() string {

	flag.Usage = Usage

	limitPtr := flag.Float64("%", 50, "approximate percentage of file to skim (overestimates for small values)")

	flag.Parse()
	args := flag.Args()

	TARGET_PERCENT = *limitPtr

	if len(args) != 1 {
		fmt.Println("Missing pure text filename to scan")
		os.Exit(-2)
	} 

	return args[0]
}

//**************************************************************

func Usage() {
	
	fmt.Println("usage: Text2N4L [-% percent] filename\n")
	flag.PrintDefaults()

	os.Exit(2)
}

//*******************************************************************

func RipFile2File(filename string,percentage float64){

	SST.MemoryInit()

	fmt.Println("Fractionating file...",filename)
	psf,L := SST.FractionateTextFile(filename)

	fmt.Println("Analyzing longitudinal patterns")
	ranking1 := SelectByRunningIntent(psf,L,percentage)
	fmt.Println("Analyzing statistical patterns")
	ranking2 := SelectByStaticIntent(psf,L,percentage)
	fmt.Println("Merging selections")
	selection := MergeSelections(ranking1,ranking2)

	fmt.Println("Extracting ambient phrases for context")

	// We only want short fragments for context, else we're repeating
	// significant context info from teh actual samples

	const minN = 1 // >= N_GRAM_MIN
	const maxN = 3 // <= N_GRAM_MAX

	f,s,ff,ss := SST.ExtractIntentionalTokens(L,selection,minN,maxN)

	WriteOutput(filename,selection,L,percentage,f,s,ff,ss)
}

//*******************************************************************

func WriteOutput(filename string,selection []SST.TextRank,L int, percentage float64,anom_by_part[][]string,ambi_by_part[][]string,all_anom[]string,all_ambi[]string) {

	// See AddMandatory() in N4L.go for reserved names (TBD, collect these one day as const)

	outputfile := filename + "_edit_me.n4l"

	fp, err := os.Create(outputfile)

	if err != nil {
		fmt.Println("Failed to open file for writing: ",outputfile)
		os.Exit(-1)
	}

	defer fp.Close()

	fmt.Fprintf(fp," - Samples from %s\n",filename)

	fmt.Fprintf(fp,"\n# (begin) ************\n")

	filealias := strings.Split(filename,".")[0]
	fmt.Fprintf(fp,"\n :: _sequence_ , %s::\n", filealias)

	var partcheck = make(map[string]bool)
	var parts []string
	var lastpart string
	
	for i := range selection {

		context := SpliceSet(ambi_by_part[selection[i].Partition])
		part := PartName(selection[i].Partition,filealias,context)

		// Add context from n = 2,3 fractions

		if part != lastpart {
			if len(context) > 0 {
				fmt.Fprintf(fp,"\n :: %s ::\n",context)
				lastpart = part
			}
		}

		fmt.Fprintf(fp,"\n@sen%d   %s\n",selection[i].Order,Sanitize(selection[i].Fragment))

		fmt.Fprintf(fp,"              \" (%s) %s\n",SST.INV_CONT_FOUND_IN_L,part)

		AddIntentionalContext(fp,anom_by_part[selection[i].Partition])

		if !partcheck[part] {
			parts = append(parts,part)
			partcheck[part] = true
		}
	}
	
	fmt.Fprintf(fp,"\n# (end) ************\n")

	// some stats
	
	fmt.Fprintf(fp,"\n# Final fraction %.2f of requested %.2f\n",float64(len(selection)*100)/float64(L),percentage)
	
	fmt.Fprintf(fp,"\n# Selected %d samples of %d: ",len(selection),L)
	
	for i := range selection {
		fmt.Fprintf(fp,"%d ",selection[i].Order)
		}
	
	fmt.Fprintf(fp,"\n#\n")

	// document the parts

	fmt.Fprintf(fp,"\n :: themes and topics you might want to annotate/replace ::\n")

	fmt.Fprintf(fp,"\n :: parts, sections ::\n")

	for p := range parts {
		fmt.Fprintf(fp,"\n %s\n",parts[p])
		for w := range ambi_by_part[p] {
			fmt.Fprintf(fp,"  #AMBI %s\n",ambi_by_part[p][w])
		}

		for w := range anom_by_part[p] {
			fmt.Fprintf(fp,"   #INTENT %s\n",anom_by_part[p][w])
		}
	}

	// whole document summary

	for w := range all_ambi {
		fmt.Fprintf(fp," # %s\n",all_ambi[w])
	}
	
	for w := range all_anom {
		fmt.Fprintf(fp,"  # %s\n",all_anom[w])
	}

	fmt.Println("Wrote file",outputfile)
	fmt.Printf("Final fraction %.2f of requested %.2f sampled\n",float64(len(selection)*100)/float64(L),percentage)

}

//*******************************************************************

func PartName(p int,file string,context string) string {

	// include ambient context in the section name

	return fmt.Sprintf("part %d of %s with %s",p,file,context)
}

//*******************************************************************

func SpliceSet(ctx []string) string {
	return strings.Join(ctx, ", ")
}

//*******************************************************************

func AddIntentionalContext(fp *os.File,ctx []string) {
	
	for w := 0; w < len(ctx); w++ {
		fmt.Fprintf(fp,"              \" (%s) %s\n",SST.NEAR_FRAG_L,ctx[w])
	}
}

//*******************************************************************

func Sanitize(s string) string {
	replacer := strings.NewReplacer("(", "[", ")", "]")
	
	return replacer.Replace(s)
}

//*******************************************************************

func SelectByRunningIntent(psf [][][]string,L int,percentage float64) []SST.TextRank {

	// Rank sentences

	const coherence_length = SST.DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var sentences []SST.TextRank
	var sentence_counter int

	for p := range psf {

		for s := range psf[p] {

			score := 0.0
			text := ""

			for f := 0; f < len(psf[p][s]); f++ {

				score += SST.RunningIntentionality(sentence_counter,psf[p][s][f])

				text += psf[p][s][f]

				if f < len(psf[p][s])-1 {
					text += ", "
				}
			}

			var this SST.TextRank
			this.Fragment = text
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

func SelectByStaticIntent(psf [][][]string,L int,percentage float64) []SST.TextRank {

	// Rank sentences

	const coherence_length = SST.DUNBAR_30   // approx narrative range or #sentences before new point/topic

	var sentences []SST.TextRank
	var sentence_counter int

	for p := range psf {

		for s := range psf[p] {

			score := 0.0
			text := ""

			for f := 0; f < len(psf[p][s]); f++ {

				score += SST.AssessStaticIntent(psf[p][s][f],L,SST.STM_NGRAM_FREQ,1)

				text += psf[p][s][f]

				if f < len(psf[p][s])-1 {
					text += ", "
				}
			}

			var this SST.TextRank
			this.Fragment = text
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

func OrderAndRank(sentences []SST.TextRank,percentage float64) []SST.TextRank {

	var selections []SST.TextRank

	// Order by intentionality first to skim cream

	sort.Slice(sentences, func(i, j int) bool {
		return sentences[i].Significance > sentences[j].Significance
	})

	// Measure relative threshold for percentage of document
	// the lower the threshold, the lower the significance of the document

	threshold := percentage / 100.0

	limit := int(threshold * float64(len(sentences)))

	// Skim

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

func MergeSelections(one []SST.TextRank,two []SST.TextRank) []SST.TextRank{

	var merge []SST.TextRank
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


