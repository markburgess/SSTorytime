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

func CheckNewContext(fp *os.File, last_context string, ctxset []string, filealias string) (string,string) {

	context := SpliceSet(ctxset)

	if context != last_context {
		fmt.Fprintf(fp,"\n-:: _sequence_ ::")
		fmt.Fprintf(fp,"\n\n :: %s / %s :: \n\n",context,filealias)
		last_context = context
	}
	
	return context,last_context
}

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
		fmt.Println(" Found chapter:",name)

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
				havedigit++
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
		KeepProperName(s,"")
		return true
	}
	
	return false
}

//*****************************************************************

func KeepProperName(name,explain string) {

	const minimum_name_length = 3

	if strings.Contains(name,"date") {
		return
	}
	
	if len(name) < minimum_name_length {
		return
	}

	_,ok := PROPER_NAME[name]

	if IsEmptyExplainer(explain) && !ok {
		PROPER_NAME[name] = ""
		return
	}

	// What if explain is a list with parentheses?
	// NO (Norway), UK (United Kingdom), DE (Germany) ... etc
	// Now name is not the important part

	if !IsListExplainer(explain) {
		if PROPER_NAME[name] == "" && !IsEmptyExplainer(explain) {
			PROPER_NAME[name] = explain
		}
	}
}

//*****************************************************************

func IsEmptyExplainer(str string) bool {

	if str == "" {
		return true
	}

	const reasonable_word_length = 50

	if len(str) > reasonable_word_length && !strings.Contains(str," ") {
		return true
	}
	
	return false
}

//*****************************************************************

func IsListExplainer(str string) bool {

	if strings.Contains(str,",") && strings.Contains(str,"[") {

		list := strings.Split(str,",")

		for _,each := range list {
			NameValue(each,"[")
		}
		
		return true
	}

	return false
}

//*****************************************************************

func CompileTabular(fp *os.File, alias string, tables []MDTable) {

	// Map to n4l

	for _, table := range tables {
		
		fmt.Fprintf(fp,"\n\n :: %s tabular spec, %s :: \n\n",alias,strings.ToLower(table.Context))
		
		for _,row := range table.Rows {

			var term string
			var meaning string

			for _,cell := range row.Cells {

				head := cell.Header
				colval := cell.Value

				if IsNumber(colval) {
					// random numbers have no semantic value
					break
				}

				if head == "Term" || strings.Contains(head,"col") {
					if term == "" {
						term = colval
						continue
					}
				}

				// rest of row

				if term == "" {
					term = colval
					continue
				}

				meaning = colval
				
				if term != "" && meaning != "" {
					KeepProperName(term,meaning)
					term = ""
					meaning = ""
				}
			}
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

	searchlist := GetTableStakes(tables)
	
	for _, str := range searchlist {
		SearchNames(str)
		ConsecutiveWordPhrase(str)
	}
}

// *********************************************************************

func GetTableStakes(tables []MDTable) []string {

	// Look through identifiers

	var idemplist = make(map[string]int)
	
	for _, table := range tables {

		idemplist[table.Context]++
		
		for _,h := range table.Headers {
			idemplist[h]++
		}
		
		for _, row := range table.Rows {
			for _, cell := range row.Cells {
				idemplist[cell.Value]++
			}
		}
	}
	
	return Map2List(idemplist)
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

//*******************************************************************

func SearchStringForProperNames(text string) []string {

	// first break the string by punctuated regions

	fields := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsPunct(r)
	})

	const minimum_string = 2
	var retval = make(map[string]int)
	
	for _, part := range fields {

		if len(part) < minimum_string {
			continue
		}

		// Now split each field on spaces

		words := strings.Split(part," ")
		
		if len(words) < minimum_string {
			continue
		}
		
		var keep []string
		
		for _,w := range words {

			l := strings.ToLower(w)
			
			if ExcludedByBindings(l,l,l) {
				continue
			}
			
			if IsCapitalized(w) {
				keep = append(keep,w)
			} else {
				if len(keep) > 1 {
					name := strings.Join(keep," ")
					KeepProperName(name,"")
					retval[name]++
					keep = nil
				}
			}
		}
	}

	return Map2List(retval)
}

// *********************************************************************

func IsCapitalized(s string) bool {

	word := []rune(s)

	if len(word) > 2 {
		if unicode.IsUpper(word[0]) && unicode.IsLower(word[1]) {
			return true
		}
	}
	
	return false
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
