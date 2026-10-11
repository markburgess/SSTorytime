//*****************************************************************
//
// text_lookslike.go
//
//*****************************************************************

package SSTorytime

import (
	"fmt"
	"strings"
	"unicode"
	//"sort"
	//"regexp"
	_ "github.com/lib/pq"
)

// **************************************************************************

func LooksLikeChapSection(org string) (bool,string,string) {

	if len(org) == 0 {
		return false,"unnumbered",""
	}

	org = strings.TrimSpace(strings.ToLower(org))

	if strings.HasPrefix(org,"I ") {
		return false,"unnumbered",""
	}

	if strings.HasPrefix(org,"chapter ") {
		name := org[len("chapter "):]
		fmt.Println("\n Found chapter:",name)

		namestart := strings.Split(name," ")[0]
		hasnumber := LooksLikeNumber(namestart)

		if hasnumber {
			return true,namestart,name
		} else {
			return true,"unnumbered",name
		}
	}
	
	// Is the line ALL CAPS? This could be a Note To Self, but fine.
	
	if IsAllUpper(org) {
		return true,org,"unnumbered"
	}

	line := strings.Split(org," ")
	linestart := line[0]

	rest := strings.Join(line[1:]," ")
	
	if len(linestart) > 2 && linestart[len(linestart)-1] == ')' {
		rest = linestart[:len(linestart)-1]
	}
	
	// check start

	hasnumber := LooksLikeNumber(linestart)

	if hasnumber {
		return true,linestart,rest
	}
	
	return false,"unnumbered",org
}

//*****************************************************************

func LooksLikeNumber(s string) bool {
	
	var numbers = []rune{'i','v','x','c','l'}
	var havedigit int

	if len(s) < 2 {
		return false
	}
	
	for _,c := range strings.ToLower(s) {

		goaround := false
		
		if unicode.IsSpace(c) {
			return false
		}

		if unicode.IsDigit(c) {
			havedigit++
			continue
		}

		if havedigit > 0 && c == '.' {
			havedigit++
			continue
		}		

		for _,t := range numbers {
			if t == c {
				goaround = true
				break
			}
		}

		if goaround {
			continue
		}

		if unicode.IsLetter(c) || unicode.IsPunct(c) {
			return false
		}
	}

	return true
}

//*****************************************************************

func LooksLikeProperName(s string) bool {


	return false
}



//
// text_lookslike.go
//
