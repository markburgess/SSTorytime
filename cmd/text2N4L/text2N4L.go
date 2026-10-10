//
// Scan a document and pick out the sentences that are measured to
// be high in "intentionality" or potential knowledge significance
// using two methods: dynamic running and static posthoc assessment
//

package main

import (
	"os"
	"fmt"
	"flag"
	"strings"

	SST "github.com/markburgess/SSTorytime/pkg/SSTorytime"
)

var TARGET_PERCENT float64 = 50.0
var MARKDOWN bool = false

//**************************************************************
// BEGIN
//**************************************************************

func main() {

	const max_class = 100

	input := GetArgs()

	if MARKDOWN {
		RipMarkdown(input,TARGET_PERCENT)
	} else {
		RipFile2File(input,TARGET_PERCENT)
	}
}

//**************************************************************

func GetArgs() string {

	flag.Usage = Usage

	limitPtr := flag.Float64("%", 50, "approximate percentage of file to skim (overestimates for small values)")
	md := flag.Bool("md", false, "interpret file as markdown text")
	
	flag.Parse()
	args := flag.Args()

	TARGET_PERCENT = *limitPtr
	MARKDOWN = *md
	
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

func RipFile2File(filename string,percentage float64) {

	SST.InitSSTorytime()

	fmt.Println("Fractionating txt file...",filename)
	psf,L := SST.FractionateTextFile(filename)
	fmt.Println("Analyzing longitudinal patterns")
	ranking1 := SST.SelectByRunningIntent(psf,L,percentage)
	fmt.Println("Analyzing transverse statistical patterns")
	ranking2 := SST.SelectByStaticIntent(psf,L,percentage)
	fmt.Println("Merging selections")
	selection := SST.MergeSelections(ranking1,ranking2)

	fmt.Println("Extracting ambient phrases for context")

	// We only want short fragments for context, else we're repeating
	// significant context info from teh actual samples

	const minN = 1 // >= N_GRAM_MIN
	const maxN = 4 // <= N_GRAM_MAX

	f,s,ff,ss := SST.ExtractIntentionalTokens(L,selection,minN,maxN)

	WriteOutput(filename,selection,L,percentage,f,s,ff,ss)
}

//*******************************************************************

func RipMarkdown(filename string,percentage float64) {

	for i := 1; i < SST.N_GRAM_MAX; i++ {
		
		SST.STM_NGRAM_FREQ[i] = make(map[string]float64)
		SST.STM_NGRAM_LOCA[i] = make(map[string][]int)
		SST.STM_NGRAM_LAST[i] = make(map[string]int)
	}

	fmt.Println("Fractionating markdown file...",filename)
	psf,L := SST.FractionateMarkdown(filename)
	fmt.Println("Analyzing frequencies")
	SST.AnnotateFractionIntent(psf)

	fmt.Println("Analyzing longitudinal patterns")
	ranking1 := SST.SelectByRunningIntent(psf,L,percentage)
	fmt.Println("Analyzing transverse statistical patterns")
	ranking2 := SST.SelectByStaticIntent(psf,L,percentage)
	fmt.Println("Merging selections")
	selection := SST.MergeSelections(ranking1,ranking2)

	fmt.Println("Extracting ambient phrases for context")

	// We only want short fragments for context, else we're repeating
	// significant context info from teh actual samples

	const minN = 1 // >= N_GRAM_MIN
	const maxN = 4 // <= N_GRAM_MAX

	f,s,ff,ss := SST.ExtractIntentionalTokens(L,selection,minN,maxN)
	
	WriteOutput(filename,selection,L,percentage,f,s,ff,ss)

}


//*******************************************************************

func WriteOutput(filename string,selection []SST.TextRank,L int, percentage float64,anom_by_part[][]string,ambi_by_part[][]string,all_anom[]string,all_ambi[]string) {

	// See AddMandatory() in N4L.go for reserved names (TBD, collect these one day as const)

	var collected_fragments = make(map[string][]string)

	outputfile := filename + "_edit_me.n4l"

	fp, err := os.Create(outputfile)

	if err != nil {
		fmt.Println("Failed to open file for writing: ",outputfile)
		os.Exit(-1)
	}

	defer fp.Close()

	fmt.Fprintf(fp," - Excerpts from %s\n",filename)

	fmt.Fprintf(fp,"\n # TABLE OF CONTENTS ...")
	fmt.Fprintf(fp,"\n # themes and topics ")
	fmt.Fprintf(fp,"\n # selected samples ")
	fmt.Fprintf(fp,"\n # final fraction %.2f of requested %.2f\n",float64(len(selection)*100)/float64(L),percentage)
	fmt.Fprintf(fp,"\n # selected samples ")
	fmt.Fprintf(fp,"\n # concepts by part/region \n")

	fmt.Fprintf(fp,"\n# (begin) ************\n")

	filealias := strings.Split(filename,".")[0]


	// Lookup tables of contents from Markdown

	SST.CompileTabular(fp,filealias,SST.TABLES)
	SST.CompileTOC(fp,filealias,SST.SECTIONS)

	// 

	fmt.Fprintf(fp,"\n#######################################")
	fmt.Fprintf(fp,"\n :: _sequence_ , %s::\n", filealias)
	fmt.Fprintf(fp,"#######################################\n")
	
	var partcheck = make(map[string]bool)
	var parts []string
	var lastpart string
	var already = make(map[string]bool)

	for i := range selection {

		if len(selection[i].Fragment) < 1 {
			continue
		}
		
		var part string
		context := SST.SpliceSet(ambi_by_part[selection[i].Partition])
		
		if len(selection[i].Title) > 0 {
			part = selection[i].Title
		} else {
			part = PartName(selection[i].Partition,filealias,context)
		}
		
		// Add context from n = 2,3 fractions
		
		if part != lastpart {
			if len(context) > 0 {
				fmt.Fprintf(fp,"\n#######################################")
				fmt.Fprintf(fp,"\n :: %s ::\n",context)
				fmt.Fprintf(fp,"#######################################\n")
			}
			
			lastpart = part
			
		}

		fmt.Fprintf(fp,"\n@sen%d   %s\n\n",selection[i].Order,SST.SanitizeParen(selection[i].Fragment))		
		fmt.Fprintf(fp,"              \" (%s) %s\n",SST.INV_CONT_FOUND_IN_S,part)
		
		AddIntentionalContext(collected_fragments,part,anom_by_part[selection[i].Partition],already)
		
		if !partcheck[part] {
			parts = append(parts,part)
			partcheck[part] = true
		}
	}
	
	fmt.Fprintf(fp,"\n -:: _sequence_ , %s::\n", filealias)
	fmt.Fprintf(fp,"\n# (end) ************\n")

	// some stats

	fmt.Fprintf(fp,"\n# Final fraction %.2f of requested %.2f\n",float64(len(selection)*100)/float64(L),percentage)

	WriteSampleSelections(fp,selection,L)

	// add the parts' fragments

	fmt.Fprintf(fp,"\n #\n # Concepts by part / region \n #\n")

	for key := range collected_fragments {

		fmt.Fprintf(fp,"\n\n %s\n",key)
		for _,s := range collected_fragments[key] {
			fmt.Fprintf(fp,"              \" (%s) %s\n",SST.CONT_FRAG_S,s)
		}
	}

	// document the parts

	fmt.Fprintf(fp,"\n #\n # SUMMARY OF CONTEXT AND SIGNIFICANT FRAGMENTS\n #\n")

	fmt.Fprintf(fp,"\n :: themes and topics you might want to annotate/replace ::\n")

	fmt.Fprintf(fp,"\n :: parts, sections ::\n")

	for p := range parts {
		
		container,exists := SST.DOC_DIRECTORY[parts[p]]

		if exists {
			fmt.Fprintf(fp,"\n %s (%s) %s \n",parts[p],SST.INV_CONT_FRAG_IN_S,container)
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


	fmt.Println("Wrote file",outputfile)
	fmt.Printf("Final fraction %.2f of requested %.2f sampled\n",float64(len(selection)*100)/float64(L),percentage)
		
}

//*******************************************************************

func WriteSampleSelections(fp *os.File, selection []SST.TextRank, L int) {

	fmt.Fprintf(fp,"\n# Selected %d samples of %d: ",len(selection),L)

	for i := range selection {
		fmt.Fprintf(fp,"%d ",selection[i].Order)
		}

	fmt.Fprintf(fp,"\n#\n")
}

//*******************************************************************

func PartName(p int,file string,context string) string {

	basename := fmt.Sprintf("part %d of %s",p,file)

	// include ambient context in the section name

	if len(context) > 0 {
		return fmt.Sprintf("%s about: %s",basename,context)
	} else {
		return basename
	}
}

//*******************************************************************

func AddIntentionalContext(collected map[string][]string,key string,ctx []string,already map[string]bool) {

	for w := 0; w < len(ctx); w++ {

		if !already[ctx[w]] {
			collected[key] = append(collected[key],ctx[w])
			already[ctx[w]] = true
		}
	}
}


//
// end text2N4L
//

