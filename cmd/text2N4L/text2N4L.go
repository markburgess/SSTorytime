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
	"path/filepath"
	
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

	dir, file := filepath.Split(filename)
	fmt.Println("Fractionating txt file...",filename)
	psf,L := SST.FractionateTextFile(dir,file)
	fmt.Println("Analyzing frequencies")
	SST.AnalyzeFractionIntent(psf)
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

	const with_title = true
	alias,_ := SST.MakeTitleFromFileName(filename)

	WriteOutput(filename,alias,selection,L,percentage,f,s,ff,ss,with_title)
}

//*******************************************************************

func RipMarkdown(filename string,percentage float64) {

	for i := 1; i < SST.N_GRAM_MAX; i++ {
		
		SST.STM_NGRAM_FREQ[i] = make(map[string]float64)
		SST.STM_NGRAM_LOCA[i] = make(map[string][]int)
		SST.STM_NGRAM_LAST[i] = make(map[string]int)
	}

	dir, file := filepath.Split(filename)
	fmt.Println("Fractionating markdown file...",filename)
	psf,L,_ := SST.FractionateMarkdown(dir,file)
	fmt.Println("Analyzing frequencies")
	SST.AnalyzeFractionIntent(psf)

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

	const with_title = true
	alias,_ := SST.MakeTitleFromFileName(filename)
	
	WriteOutput(filename,alias,selection,L,percentage,f,s,ff,ss,with_title)
}

//*******************************************************************

func WriteOutput(filename,alias string,selection []SST.TextRank,L int, percentage float64,f[][]string,s[][]string,ff[]string,ss[]string,with_title bool) {

	outputfile := filename + "_edit_me.n4l"

	fp, err := os.Create(outputfile)

	if err != nil {
		fmt.Println("Failed to open file for writing: ",outputfile)
		os.Exit(-1)
	}

	defer fp.Close()

	SST.GenerateOutput(fp,alias,selection,L,percentage,f,s,ff,ss,with_title)
}


//
// end text2N4L
//

