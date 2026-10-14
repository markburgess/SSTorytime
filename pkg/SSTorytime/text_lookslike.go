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
	fmt.Fprintf(fp,"\n :: _sequence_, table of contents, overview :: \n")
	
	for _,sec := range sections {
		if !done[sec.Title] {
			if LooksLikeProperName(sec.Title) || strings.Contains(sec.Title," ") {
				done[sec.Title] = true
				fmt.Fprintf(fp,"\n%s (%s) Table of Contents - %s\n  \" (%s) Overview - %s\n",sec.Title,INV_CONT_NAME_L,filealias,INV_CONT_NAME_L,filealias)
			}
		}
	}

	fmt.Fprintf(fp,"\n -:: _sequence_ :: \n")
	fmt.Fprintf(fp,"\n # END table of contents ############################\n")
}

//*****************************************************************

func CompileTabular(fp *os.File, alias string, tables []MDTable) {

	// Map to n4l

	fmt.Println("TABLES......",alias)
	for _, table := range tables {
		
		fmt.Printf(" :: %s, table, %s :: \n",alias,strings.ToLower(table.Context))
		
		for _,h := range table.Headers {

			fmt.Println("HEADER:",h)
			if len(h) > 0 && LooksLikeProperName(h) || strings.Contains(h," ") {
				fmt.Printf("\n # table %s  ############## \n",table.Context)
				fmt.Printf(" %s  (%s) %s\n",h,INV_EXPR_HEADER_L,table.Context)	
			}
		}
		
		for _,row := range table.Rows {

			var first string

			for pos, cell := range row.Cells {

				head := cell.Header
				colval := cell.Value

				CheckTabularValues(colval)
				
				if pos == 0 && (head == "Term" || strings.Contains(head,"col")) {
					if LooksLikeProperName(colval) {
						fmt.Printf("PROPER NAME: %s \n",colval)
						if val,ok := PROPER_NAME[colval]; ok {
							if len(val) == 0 {
								PROPER_NAME[colval] = ""
							}
						}
					}
					first = colval
				}

				if pos > 0 && (head == "Meaning"  || strings.Contains(head,"col")) {
					head = "Value"
					fmt.Printf("%s --> %s (stands for) %s\n",alias, first,colval)
					//deliver
				}
				
				if pos == 0 {
					head = ""
					fmt.Printf("%s ITEM: %s %s \n",alias, head,colval)
					first = head + " " + colval
				} else {
					fmt.Printf("%s %s (has policy) %s ****`%s` \n",alias, first,head,colval)
				}
			}
		}
	}
	
	fmt.Println("END TABLES......")
}
	
// *********************************************************************

func CheckTabularValues(s string) {

	list := strings.Split(s,",")

	for _,v := range list {
		name,value := NameValue(v)
		if LooksLikeProperName(name) {
			fmt.Printf("PROPER NAME: %s WITH VALUE %s\n",name,value)
			PROPER_NAME[name] = value
		}
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

func OutputProperNameAnnotation(fp *os.File, filealias string,parent string,list []string) {

	for _,child := range list {
		
		fmt.Fprintf(fp,"\n%s (stands for) %s\n",child,PROPER_NAME[child])
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
