//******************************************************************
//
// Try out neighbour search for all ST stypes together
//
// Prepare:
// cd examples
// ../src/N4L-db -u chinese.n4l
//
//******************************************************************

package main

import (
	"os"
	"fmt"
	"flag"

        SST "SSTorytime"
)

var path [8][]string

//******************************************************************

func main() {

	load_arrows := false
	sst := SST.Open(load_arrows)

	dbchapters := SST.GetDBChaptersMatchingName(sst,"")

	fmt.Println("\nThe database currently caches the following chapters:\n")
	for c := range dbchapters {
		fmt.Printf("%d. - chapter: \"%s\"\n",c+1,dbchapters[c])
	}
	fmt.Println()

	args := Init()
	chapter := args[0]

	DeleteChapter(sst,chapter)

	SST.Close(sst)
}

//**************************************************************

func Init() []string {

	flag.Usage = Usage

	forcePtr := flag.Bool("force", false,"force remove")

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		Usage()
		os.Exit(1);
	}

	if !*forcePtr {
		fmt.Println("Are you sure you want to remove a chapter? Use -force to confirm.")
		os.Exit(1);
	}

	SST.MemoryInit()

	return args
}

//**************************************************************

func Usage() {
	
	fmt.Printf("\n\nusage: removeN4L \"chapter name\"\n")
	flag.PrintDefaults()
	os.Exit(2)
}

//******************************************************************

func DeleteChapter(sst SST.PoSST,chapter string) {

	qstr := fmt.Sprintf("select DeleteChapter('%s')",chapter)

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error running deletechapter function:",qstr,err)
	} else {
		fmt.Println("Deleted",chapter)
	}

	row.Close()

}









