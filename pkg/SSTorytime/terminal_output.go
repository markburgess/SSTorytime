//****************************************************************************
//
// terminal_output.go
//
//****************************************************************************

package SSTorytime

import (
	"fmt"
	"unicode"
	_ "github.com/lib/pq"

)

//****************************************************************************

func ShowText(s string, width int) {

	var spacecounter int
	var linecounter int
	var indent string = Indent(LEFTMARGIN)

	if width < 40 {
		width = SCREENWIDTH
	}

	// Check is the string has a large number of spaces, in which case it's
	// probably preformatted,

	runes := []rune(s)

	for r := 0; r < len(runes); r++ {
		if unicode.IsSpace(runes[r]) {
			spacecounter++
		}
	} 

	if len(runes) > SCREENWIDTH - LEFTMARGIN - RIGHTMARGIN {
		if spacecounter > len(runes) / 3 {
			fmt.Println()
			fmt.Println(s)
			return
		}
	}

	// Format

	linecounter = 0

	for r := 0; r < len(runes); r++ {

		if unicode.IsSpace(runes[r]) && linecounter > width-RIGHTMARGIN {
			if runes[r] != '\n' {
				fmt.Print("\n",indent)
				linecounter = 0
				continue
			} else {
				linecounter = 0
			}
		}
		if unicode.IsPunct(runes[r]) && linecounter > width-RIGHTMARGIN {
			fmt.Print(string(runes[r]))
			r++
			if r < len(runes) && runes[r] != '\n' {
				fmt.Print("\n",indent)
				linecounter = 0
				continue
			} else {
				linecounter = 0
			}
		}

		if r < len(runes) {
			fmt.Print(string(runes[r]))
		}
		linecounter++
		
	}
}

// *********************************************************************

func ShowContext(amb,intent,key string) {

	fmt.Println()
	fmt.Println("  .......................................................")
	fmt.Printf("    Recurrent now: %s\n",key)
	fmt.Printf("    Intentional  : %s\n",intent)
	fmt.Printf("    Ambient      : %s\n",amb)
	fmt.Println("  .......................................................")

}

//****************************************************************************

func Indent(indent int) string {

	spc := ""

	for i := 0; i < indent; i++ {
		spc += " "
	}

	return spc
}

//****************************************************************************

func NewLine(n int) {

	if n % 6 == 0 {
		fmt.Print("\n    ",)
	}
}

// **************************************************************************

func Waiting(output bool,total int) {

	if !output {
		return
	}

	percent := float64(SILLINESS_COUNTER) / float64(total) * 100

	var propaganda = []string{"\n1) JOT IT DOWN WHEN YOU THINK OF IT. . .\n","\n2) TYPE IT INTO N4L AS SOON AS YOU CAN. . .\n","\n3) ORGANIZE AND TIDY YOUR NOTES EVERY DAY. . .\n","\n4) UPLOAD AND BROWSE THEM ONLINE. . .\n","\n5) AND REMEMBER, IT ISN'T KNOWLEDGE IF YOU DON'T ACTUALLY KNOW IT !!\n"}

	const interval = 2

	if SILLINESS {
		if SILLINESS_COUNTER % interval != 0 {
			fmt.Print(" ")
		} else {
			fmt.Print(string(propaganda[SILLINESS_SLOGAN][SILLINESS_POS]))
			SILLINESS_POS++

			if SILLINESS_POS > len(propaganda[SILLINESS_SLOGAN])-1 {
				SILLINESS_POS = 0
				SILLINESS = false
				SILLINESS_SLOGAN++
				if SILLINESS_SLOGAN >= len(propaganda) {
					SILLINESS_SLOGAN = 0
				}
			}
		}
	} else {
		fmt.Print(".")
	}

	if SILLINESS_COUNTER % (2000) == 0 {
		SILLINESS = !SILLINESS
		if percent > 100 {
			fmt.Printf("\n(%.1f%% - oops, have to work overtime!)\n",percent)
		} else {
			fmt.Printf("\n\n(%.1f%%) uploading . . .\n",percent)
		}
	}

	SILLINESS_COUNTER++
}


// **************************************************************************

func PrintNodeOrbit(sst *PoSST, nptr NodePtr,limit int) {

	node := GetDBNodeByNodePtr(sst,nptr)		
	fmt.Print("\"")
	ShowText(node.S,SCREENWIDTH)
	fmt.Print("\"")
	fmt.Println("\tin chapter:",node.Chap)
	fmt.Println()

	satellites := GetNodeOrbit(sst,nptr,"",limit)

	PrintLinkOrbit(satellites,EXPRESS,0)
	PrintLinkOrbit(satellites,-EXPRESS,0)
	PrintLinkOrbit(satellites,-CONTAINS,0)
	PrintLinkOrbit(satellites,LEADSTO,0)
	PrintLinkOrbit(satellites,-LEADSTO,0)
	PrintLinkOrbit(satellites,NEAR,0)

	fmt.Println()
}

// **************************************************************************

func PrintLinkOrbit(satellites [ST_TOP][]Orbit,sttype int,indent_level int) {

	t := STTypeToSTIndex(sttype)

	for n := range satellites[t] {		

		r := satellites[t][n].Radius + indent_level

		if satellites[t][n].Ctx != "" {
			txt := fmt.Sprintf(" -    (%s) - %s  \t.. in the context of %s\n",satellites[t][n].Arrow,satellites[t][n].Text,satellites[t][n].Ctx)
			text := Indent(LEFTMARGIN * r) + txt
			ShowText(text,SCREENWIDTH)
		} else {
			txt := fmt.Sprintf(" -    (%s) - %s\n",satellites[t][n].Arrow,satellites[t][n].Text)
			text := Indent(LEFTMARGIN * r) + txt
			ShowText(text,SCREENWIDTH)
		}

	}

}

// **************************************************************************

func PrintLinkPath(sst *PoSST, cone [][]Link, p int, prefix string,chapter string,context []string) {

	PrintSomeLinkPath(sst,cone, p,prefix,chapter,context,10000)
}

// **************************************************************************

func PrintSomeLinkPath(sst *PoSST, cone [][]Link, p int, prefix string,chapter string,context []string,limit int) {

	count := 0

	if len(cone[p]) > 1 {

		path_start := GetDBNodeByNodePtr(sst,cone[p][0].Dst)		
		
		start_shown := false

		var format int
		var stpath []string
		
		for l := 1; l < len(cone[p]); l++ {

			if !MatchContexts(sst,context,cone[p][l].Ctx) {
				return
			}

			NewLine(format)

			count++

			if count > limit {
				return
			}

			if !start_shown {

				if len(cone) > 1 {
					fmt.Printf("%s (%d) %s",prefix,p+1,path_start.S)
				} else {
					fmt.Printf("%s %s",prefix,path_start.S)
				}
				start_shown = true
			}

			nextnode := GetDBNodeByNodePtr(sst,cone[p][l].Dst)

			if !SimilarString(nextnode.Chap,chapter) {
				break
			}

			arr := GetDBArrowByPtr(sst,cone[p][l].Arr)

			if arr.Short == "then" {
				fmt.Print("\n   >>> ")
				format = 0
			}

			if arr.Short == "prior" {
				fmt.Print("\n   <<< ")
			}

			stpath = append(stpath,STTypeName(STIndexToSTType(arr.STAindex)))
	
			if l < len(cone[p]) {
				fmt.Print("  -(",arr.Long,")->  ")
			}
			
			fmt.Print(nextnode.S)
			format += 2
		}

		fmt.Print("\n     -  [ Link STTypes:")

		for s := range stpath {
			fmt.Print(" -(",stpath[s],")-> ")
		}
		fmt.Println(". ]\n")
	}
}



//**************************************************************

func ShowPsi(etc Etc) string {

	result := ""

	if etc.E {
		result += "event,"
	}
	if etc.T {
		result += "thing,"
	}
	if etc.C {
		result += "concept,"
	}
	return result
}



//
// terminal_output.go
//


