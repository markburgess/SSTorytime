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
	_ "github.com/lib/pq"
)

// **************************************************************************

func LooksLikeChapSection(org,filename string) (bool,string,string) {

	if len(org) == 0 {
		return false,"unnumbered",""
	}

	org = strings.TrimSpace(org)

	if strings.HasPrefix(org,"I ") {
		return false,"unnumbered",""
	}

	if strings.HasPrefix(strings.ToLower(org),"chapter ") {
		name := org[len("chapter "):]
		fmt.Println("\n Found chapter:",name)

		name = name + " [" + filename + "]"

		namestart := strings.Split(name," ")[0]
		hasnumber := LooksLikeNumber(namestart)

		if hasnumber {
			return true,namestart,name
		} else {
			return true,"unnumbered",name
		}
	}
	
	// Is the line ALL CAPS? This could be a Note To Self, but fine.

	const concept_length = 10
	
	if IsAllUpper(org) {
		name := org + " ["+filename+"]"
		if len(org) > concept_length {
			fmt.Println("\n Found possible chapter:",name)
		}
		return true,name,"unnumbered"
	}

	line := strings.Split(org," ")
	linestart := line[0]

	rest := filename + " " + strings.Join(line[1:]," ")
	
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

	var hasupper bool

	for _, r := range s {

		if unicode.IsLower(r) {
			return false
		}

		if unicode.IsUpper(r) {
			hasupper = true
		}

		if r == '-' || r == '_' || r == '/' || r == '\\' {
			continue
		}
		
		if unicode.IsPunct(r) {
			return false
		}
	}

	if hasupper {
		return true
	}
	
	return false
}

//*****************************************************************
// Build ancillary structures
//*****************************************************************

func CompileTOC(fp *os.File, filealias string, sections []ExtSentence) {

	// map to n4l
	
	var done = make(map[string]bool)

	fmt.Fprintf(fp,"\n # BEG table of contents ############################\n")
	fmt.Fprintf(fp,"\n :: _sequence_, table of contents [%s] :: \n",filealias)
	
	for _,sec := range sections {

		title := strings.ToLower(sec.Title)
		
		if !done[title] {
			if LooksLikeProperName(title) || strings.Contains(title," ") {
				done[title] = true
				fmt.Fprintf(fp,"\n%s [%s] (%s) Table of Contents - %s\n",sec.Title,filealias,INV_CONT_NAME_L,filealias)
			}
		}
	}

	fmt.Fprintf(fp,"\n -:: _sequence_ :: \n")
	fmt.Fprintf(fp,"\n # END table of contents ############################\n")
}

//*****************************************************************

func ListProperNames(fp *os.File) {

	fmt.Println("\nCompiling proper names ....\n\n")

	fmt.Fprintf(fp,"\n\n :: proper names :: \n\n")
	
	for name := range PROPER_NAME {

		fmt.Fprintf(fp,"\n %s (is explained by) %s \n",name,PROPER_NAME[name])
		
	}
}

//*****************************************************************

func CompileTabular(fp *os.File, alias string, tables []MDTable) {

	// Map to n4l

	for _, table := range tables {
		
		fmt.Fprintf(fp,"\n\n :: %s tabular spec, %s :: \n\n",alias,strings.ToLower(table.Context))
		
		for _,row := range table.Rows {

			var first string

			for pos, cell := range row.Cells {

				head := strings.ToLower(cell.Header)
				colval := strings.ToLower(cell.Value)

				if !CheckTabularValues(head,colval) {
					continue
				}

				// Cache and come back to proper names later
				
				if pos == 0 && (head == "Term" || strings.Contains(head,"col")) {
					if LooksLikeProperName(colval) {
						if val,ok := PROPER_NAME[colval]; ok {
							if len(val) == 0 {
								// save it until we know rhs
								PROPER_NAME[colval] = ""
							}
						}
					}
					first = colval
					continue
				}
				
				// start of row, subject title?
				
				if pos == 0 {
					if IsNumber(colval) {
						// random numbers have no semantic value
						continue
					}
					
					if strings.Contains(head,"col") || strings.Contains(head,alias) {
						fmt.Fprintf(fp,"%s (has concept) %s \n",head,colval)
					} else {
						fmt.Fprintf(fp,"%s [%s] (has concept) %s \n",head,alias,colval)
					}
				} else {
					fmt.Fprintf(fp,"%s [%s] %s (has policy) %s \n",first,alias,head,colval)
				}
				
				// rest of row
				
				if pos > 0 && head == "Meaning" && len(colval) > 0 {
					if val,ok := PROPER_NAME[colval]; ok {
						if len(val) == 0 {
							PROPER_NAME[first] = colval
						}
					} else {
						PROPER_NAME[first] = colval
					}
					fmt.Fprintf(fp,"%s [%s] (stands for) %s\n",first,alias,colval)
				}
			}
		}
	}
}
	
// *********************************************************************

func CheckTabularValues(head,value string) bool {

	if head == "" && IsNumber(value) {
		return false
	}

	// Numbered header items?
	
	if IsNumber(value) {
		value = head + " " + value
	}
	
	if len(value) < 3 {
		NameValue(value)
	} else {

		list := strings.Split(value,",")

		for _,v := range list {
			NameValue(v)
		}
	}

	return true
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

	//fmt.Fprintf(fp,"\n\n :: proper names, identifiers  :: \n\n")

	searchlist := GetTableStakes(tables)
	
	for _, str := range searchlist {

		// Look for ID-like common strings
		
		names := SearchNames(str)
		OutputProperNameAnnotation(fp,filealias,str,names)

		// Consecutive Word phrases

		names = ConsecutiveWordPhrase(str)
		OutputProperNameAnnotation(fp,filealias,str,names)

	}
}

// *********************************************************************

func OutputProperNameAnnotation(fp *os.File, alias string,parent string,list []string) {

	for _,child := range list {

		if len(PROPER_NAME[child]) > 0 {
			fmt.Fprintf(fp,"\n %s %s (stands for) %s\n",alias,child,PROPER_NAME[child])
		} else {
			fmt.Fprintf(fp,"\n %s %s\n",alias,child)
		}
	}
}

// *********************************************************************

func GetTableStakes(tables []MDTable) []string {

	var list []string
	
	for _, table := range tables {

		list = append(list,table.Context)
		
		for _,h := range table.Headers {

			list = append(list,h)
		}
		
		for _, row := range table.Rows {
			for _, cell := range row.Cells {

				list = append(list,cell.Value)
			}
		}
	}
	
	return list
}

// *********************************************************************

func ConsecutiveWordPhrase(s string) []string {

	words := strings.Split(s," ")

	var phrases []string
	var tmp []string
	
	for _,w := range words {

		if len(w) < 3 {
			continue
		}

		if ExcludedByBindings(w,w,w) {
			continue
		}
		
		runes := []rune(w)

		if unicode.IsUpper(runes[0]) && !unicode.IsUpper(runes[1]) {
			
			if unicode.IsPunct(runes[len(runes)-1]) {
				tmp = append(tmp,w[0:len(w)-1])
				if len(tmp) > 1 {
					newphrase := strings.Join(tmp," ")
					newphrase = strings.TrimSpace(newphrase)
					if len(newphrase) > 0 {
						phrases = append(phrases,newphrase)
					}
				}
				tmp = nil
			} else {
				tmp = append(tmp,w)
			}

		} else {
			if len(tmp) > 1 {
				newphrase := strings.Join(tmp," ")
				newphrase = strings.TrimSpace(newphrase)
				if len(newphrase) > 0 {
					phrases = append(phrases,newphrase)
				}
			}
			tmp = nil
		}
	}

	return phrases
}

// *********************************************************************

func SearchNames(text string) []string {

	if text == "" {
		return nil
	}

	var list []string
	
	pbsf_table,_ := FractionateText(text,"proper names")
	
	pieces := WalkPBSF(pbsf_table)

	// Special looking identifiers "[A-Z0-9/-_]"

	for _,txt := range pieces {

		for _,word := range strings.Split(txt," ") {

			word = strings.Trim(word,".")
			
			if LooksLikeProperName(word) {
				list = append(list,word)
			}
		}
	}

	// A sequencee of worlds starting with CAPS
	
	return list
}

// *********************************************************************

func WalkPBSF(pbsf [][]Sentence) []string {

	// Concatenate significant strings from the decomposition
	
	var list []string

	if pbsf == nil {
		return nil
	}

	for _,p := range pbsf {
		for _,s := range p {
			list = append(list,s.S)
			if s.Title != "" {
				list = append(list,s.Title)
			}
		}

	}

	return list
} 

//
// text_lookslike.go
//
