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
	"os"
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

//*****************************************************************
// Build ancillary structures
//*****************************************************************

func CompileTOC(fp *os.File, filealias string, sections []ExtSentence) {

	// map to n4l

	fmt.Fprintf(fp,"\n # BEG table of contents ############################\n")
	fmt.Fprintf(fp,"\n :: _sequence_, table of contents :: \n")
	
	for _,sec := range sections {
		fmt.Fprintf(fp,"\n%s (%s) part/section in %s\n",sec.Title,EXPR_TAB_HEADER_L,filealias)
		fmt.Fprintf(fp,"\n%s (%s) part in %s\n",sec.Title,EXPR_TABNAME_S,sec.Content)
	}

	fmt.Fprintf(fp,"\n -:: _sequence_ :: \n")
	fmt.Fprintf(fp,"\n # END table of contents ############################\n")
}

//*****************************************************************

func CompileTabular(fp *os.File, filealias string, tables []MDTable) {

	// Map to n4l

	for _, table := range tables {

		fmt.Fprintf(fp," :: _sequence_, %s :: \n",table.Context)

		for _,h := range table.Headers {
			if len(h) > 0 {
				fmt.Fprintf(fp,"\n # table %s  ############## \n",table.Context)
				fmt.Fprintf(fp," %s  (%s) %s\n",h,INV_EXPR_TAB_HEADER_L,table.Context)	
			}
		}

		for _, row := range table.Rows {

			last := ""
			
			for _, cell := range row.Cells {
				fmt.Fprintf(fp," %s (%s) %s \n", table.Context,CONT_TABLE_NAME_L,cell.Value)
				if !strings.HasPrefix(cell.Header,"col_") {
					fmt.Fprintf(fp," %s (%s) %s \n", cell.Header,INV_EXPR_TAB_HEADER_L, cell.Value)
				}
				if last != "" {
					fmt.Fprintf(fp," %s (%s) %s \n", last,EXPR_TABNAME_S, cell.Value)
					last = cell.Value
				}
			}
		}

		fmt.Fprintf(fp," -:: _sequence_ :: \n")
	}

}

// *********************************************************************

func CompileProperNames(fp *os.File, filealias string,selection []TextRank,tables []MDTable) {

	// Run through selected sentences and link to meaningful names
	// in tables and other places

	// First look through tables, since these may have explanations

	// type MDTable struct {
	// Headers []string
	// Rows    []MDRow
	// Context string
        // }

	for _, table := range tables {
		
		fmt.Printf(" :: CONTEXT %s :: \n",table.Context)
		
		if LooksLikeProperName(table.Context) {
			
		}
		
		for _,h := range table.Headers {
			if len(h) > 0 {
				fmt.Printf("HEADER %s\n",h)	
			}
		}
		
		for _, row := range table.Rows {
			
			for _, cell := range row.Cells {
				fmt.Printf("CELL %s \n",cell.Value)
			}
		}
	}
}



//
// text_lookslike.go
//
