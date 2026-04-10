// **************************************************************************
//
// tools.go
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	"os"
	"strings"
	"sort"
	"regexp"
	"net/http"
	"io/ioutil"
	_ "github.com/lib/pq"

)


// **************************************************************************

func SplitChapters(str string) []string {

	run := []rune(str)

	var part []rune
	var retval []string

	for r := 0; r < len(run); r++ {
		if run[r] == ',' && (r+1 < len(run) && run[r+1] != ' ') {
			retval = append(retval,string(part))
			part = nil
		} else {
			part = append(part,run[r])
		}
	}

	retval = append(retval,string(part))

	return retval
}

// **************************************************************************

func List2Map(l []string) map[string]int {

	var retvar = make(map[string]int)

	for s := range l {
		retvar[strings.TrimSpace(l[s])]++
	}

	return retvar
}

// **************************************************************************

func Map2List(m map[string]int) []string {

	var retvar []string

	for s := range m {
		retvar = append(retvar,strings.TrimSpace(s))
	}

	sort.Strings(retvar)
	return retvar
}

// **************************************************************************

func List2String(list []string) string {

	var s string

	sort.Strings(list)

	for i := 0; i < len(list); i++ {
		s += list[i]
		if i < len(list)-1 {
			s+= ","
		}
	}

	return s
}

// **************************************************************************

func SQLEscape(s string) string {

	undo := strings.ReplaceAll(s,"''","'")
	escaped := strings.ReplaceAll(undo,"'","''")

	return string(escaped)
}

// **************************************************************************

func Array2Str(arr []string) string {

	var s string

	for a := 0; a < len(arr); a++ {
		s += arr[a]
		if a < len(arr)-1 {
			s += ", "
		}
	}

	return s
}

// **************************************************************************

func Str2Array(s string) ([]string,int) {

	var non_zero int
	s = strings.Replace(s,"{","",-1)
	s = strings.Replace(s,"}","",-1)
	s = strings.Replace(s,"\"","",-1)

	arr := strings.Split(s,",")

	for a := 0; a < len(arr); a++ {
		arr[a] = strings.TrimSpace(arr[a])
		if len(arr[a]) > 0 {
			non_zero++
		}
	}

	return arr,non_zero
}

//******************************************************************

func ParseLiteralNodePtrs(names []string) ([]NodePtr,[]string) {

	var current []rune
	var rest []string
	var nodeptrs []NodePtr

	// Note that, when we get here (a,b) is already splut into (a and b)

	for n := range names {

		line := []rune(names[n])
		
		for i := 0; i < len(line); i++ {
			
			if line[i] == '(' {

				rs := strings.TrimSpace(string(current))

				if len(rs) > 0 {
					rest = append(rest,string(current))
					current = nil
				}
				continue
			}
			
			if line[i] == ')' {
				np := string(current)
				var nptr NodePtr
				var a,b int = -1,-1
				fmt.Sscanf(np,"%d,%d",&a,&b)
				if a >= 0 && b >= 0 {
					nptr.Class = a
					nptr.CPtr = ClassedNodePtr(b)
					nodeptrs = append(nodeptrs,nptr)
					current = nil
				} else {
					rest = append(rest,"("+np+")")
					current = nil
				}
				continue
			}
			current = append(current,line[i])
		}
		rs := strings.TrimSpace(string(current))

		if len(rs) > 0 {
			rest = append(rest,rs)
		}
		current = nil
	}

	return nodeptrs,rest
}

// **************************************************************************

func ParseSQLNPtrArray(s string) []NodePtr {

	stringify := ParseSQLArrayString(s)

	var retval []NodePtr
	var nptr NodePtr

	for n := 0; n < len(stringify); n++ {
		fmt.Sscanf(stringify[n],"(%d,%d)",&nptr.Class,&nptr.CPtr)
		retval = append(retval,nptr)
	}

	return retval
}

// **************************************************************************

func ParseSQLArrayString(whole_array string) []string {

	// array as {"(1,2,3)","(4,5,6)",spacelessstring}

      	var l []string

    	whole_array = strings.Replace(whole_array,"{","",-1)
    	whole_array = strings.Replace(whole_array,"}","",-1)

	uni_array := []rune(whole_array)

	var items []string
	var item []rune
	var protected = false

	for u := range uni_array {

		if uni_array[u] == '"' {
			protected = !protected
			continue
		}

		if !protected && uni_array[u] == ',' {
			items = append(items,string(item))
			item = nil
			continue
		}

		item = append(item,uni_array[u])
	}

	if item != nil {
		items = append(items,string(item))
	}

	for i := range items {

	    s := strings.TrimSpace(items[i])

	    l = append(l,s)
	    }

	return l
}

// **************************************************************************

func FormatSQLIntArray(array []int) string {

        if len(array) == 0 {
		return "'{ }'"
        }

	sort.Slice(array, func(i, j int) bool {
		return array[i] < array[j]
	})

	var ret string = "'{ "
	
	for i := 0; i < len(array); i++ {
		ret += fmt.Sprintf("%d",array[i])
	    if i < len(array)-1 {
	    ret += ", "
	    }
        }

	ret += " }' "

	return ret
}

// **************************************************************************

func FormatSQLStringArray(array []string) string {

        if len(array) == 0 {
		return "'{ }'"
        }

	sort.Strings(array) // Avoids ambiguities in db comparisons

	var ret string = "'{ "
	
	for i := 0; i < len(array); i++ {

		if len(array[i]) == 0 {
			continue
		}

		ret += fmt.Sprintf("\"%s\"",SQLEscape(array[i]))
	    if i < len(array)-1 {
	    ret += ", "
	    }
        }

	ret += " }' "

	return ret
}

// **************************************************************************

func FormatSQLNodePtrArray(array []NodePtr) string {

        if len(array) == 0 {
		return "'{ }'"
        }

	var ret string = "'{ "
	
	for i := 0; i < len(array); i++ {
		ret += fmt.Sprintf("\"(%d,%d)\"",array[i].Class,array[i].CPtr)
	    if i < len(array)-1 {
	    ret += ", "
	    }
        }

	ret += " }' "

	return ret
}

// **************************************************************************

func FormatSQLLinkArray(array []Link) string {

	// {"(81,1,2,\"(1,0)\")","(108,1,2,\"(3,11)\")","(118,1,2,\"(2,1348)\")"}

	var s string

	for _,lnk := range array {

		l := fmt.Sprintf("(%d, %f, %d, \\\"(%d,%d)\\\")",lnk.Arr,lnk.Wgt,lnk.Ctx,lnk.Dst.Class,lnk.Dst.CPtr)
		s += fmt.Sprintf("\"%s\",",l)
	}

	s = "{" + strings.Trim(s,",") + "}"

	return s
}

// **************************************************************************

func ParseSQLLinkString(s string) Link {

        // e.g. (77,0.34,334,"(4,2)")

      	var l Link

	s = strings.Replace(s,"\"","",-1)
	s = strings.Replace(s,"\\","",-1)
	s = strings.Replace(s,"(","",-1)
	s = strings.Replace(s,")","",-1)
	
        items := strings.Split(s,",")

	for i := 0; i < len(items); i++ {
		items[i] = strings.Replace(items[i],";","",-1)
		items[i] = strings.TrimSpace(items[i])
	}

	// Arrow type
	fmt.Sscanf(items[0],"%d",&l.Arr)

	// Link weight
	fmt.Sscanf(items[1],"%f",&l.Wgt)

	// Context pointer
	fmt.Sscanf(items[2],"%d",&l.Ctx)

	// DstNPtr
	fmt.Sscanf(items[3],"%d",&l.Dst.Class)
	fmt.Sscanf(items[4],"%d",&l.Dst.CPtr)

	return l
}

//**************************************************************

func ParseLinkArray(s string) []Link {

	var array []Link

	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s,"{","")
	s = strings.ReplaceAll(s,"}","")

	if len(s) <= 2 {
		return array
	}

	strarray := strings.Split(s,"\",\"")

	for i := 0; i < len(strarray); i++ {
		link := ParseSQLLinkString(strarray[i])
		array = append(array,link)
	}
	
	return array
}

//**************************************************************

func ParseMapLinkArray(s string) []Link {

	var array []Link

	s = strings.TrimSpace(s)

	if len(s) <= 2 {
		return array
	}

	strarray := strings.Split(s,"\",\"")

	for i := 0; i < len(strarray); i++ {
		link := ParseSQLLinkString(strarray[i])
		array = append(array,link)
	}
	
	return array
}

//**************************************************************

func ParseLinkPath(s string) [][]Link {

	// Each path will start on a new line, with comma sep Link encodings

	var array [][]Link
	var index int = 0
	s = strings.TrimSpace(s)

	lines := strings.Split(s,"\n")

	for line := range lines {

		if len(lines[line]) > 0 {

			links := strings.Split(lines[line],";")

			// Actual paths need len > 1, but this is also used to seed longer paths

			if len(links) < 1 {
				continue
			}

			array = append(array,make([]Link,0))

			for l := 0; l < len(links); l++ {
				lnk := ParseSQLLinkString(links[l])
				array[index] = append(array[index],lnk)
			}
			index++
		}
	}

	if index < 1 {
		return nil
	}
	return array
}

//**************************************************************

func StorageClass(s string) (int,int) {
	
	var spaces int = 0

	var l = len(s)
	
	for i := 0; i < l; i++ {
		
		if s[i] == ' ' {
			spaces++
		}
		
		if spaces > 2 {
			break
		}
	}
	
	// Text usage tends to fall into a number of different roles, with a power law
	// frequency of occurrence in a text, so let's classify in order of likely usage
	// for small and many, we use a hashmap/btree
	
	switch spaces {
	case 0:
		return l,N1GRAM
	case 1:
		return l,N2GRAM
	case 2:
		return l,N3GRAM
	}
	
	// For longer strings, a linear search is probably fine here
        // (once it gets into a database, it's someone else's problem)
	
	if l < 128 {
		return l,LT128
	}
	
	if l < 1024 {
		return l,LT1024
	}
	
	return l,GT1024
}

// **************************************************************************

func DiracNotation(s string) (bool,string,string,string) {

	var begin,end,context string

	if s == "" {
		return false,"","",""
	}

	if s[0] == '<' && s[len(s)-1] == '>' {
		matrix := s[1:len(s)-1]
		params := strings.Split(matrix,"|")
		
		switch len(params) {
			
		case 2: 
			end = params[0]
			begin = params[1]
		case 3:
			end = params[0]
			context = params[1]
			begin = params[2]			
		default:
			fmt.Println("Bad Dirac notation, should be <a|b> or <a|context|b>")
			os.Exit(-1)
		}
	} else {
		return false,"","",""
	}

	return true,begin,end,context
}



//****************************************************************************

func IsBracketedSearchList(list []string) (bool,[]string) {

	var stripped_list []string
	retval := false

	for i := range list {

		isbrack,stripped := IsBracketedSearchTerm(list[i])

		if isbrack {
			retval = true
			stripped_list = append(stripped_list,"|"+stripped+"|")
		} else {
			stripped_list = append(stripped_list,list[i])
		}

	}

	return retval,stripped_list
}

//****************************************************************************

func IsBracketedSearchTerm(src string) (bool,string) {

	retval := false
	stripped := src

	decomp := strings.TrimSpace(src)

	if len(decomp) == 0 {
		return false, ""
	}

	if decomp[0] == '(' && decomp[len(decomp)-1] == ')' {
		retval = true
		stripped = decomp[1:len(decomp)-1]
		stripped = strings.TrimSpace(stripped)
	}

	return retval,SQLEscape(stripped)
}

//****************************************************************************

func IsExactMatch(org string) (bool,string) {

	org = strings.TrimSpace(org)

	if len(org) == 0 {
		return false,org
	}

	if org[0] == '!' && org[len(org)-1] == '!' {
		tr := strings.Trim(org,"!")
		return true,strings.ToLower(tr)
	}

	if org[0] == '|' && org[len(org)-1] == '|' {
		tr := strings.Trim(org,"|")
		return true,strings.ToLower(tr)
	}

	return false,org
}

//****************************************************************************

func IsStringFragment(s string) bool {

	tsvec_patterns := []string{"|","&","!","<->","<1>","<2>","<3>","<4>"}

	// if this is a ts_vec pattern, it's not for us

	for _,p := range tsvec_patterns {
		if strings.Contains(s,p) {
			return false
		}
	}

	// The tsvector cannot handle spaces or apostrophes(!), so fall back on LIKE %%

	str_patterns := []string{" ","-","_","'","\""}

	for _,p := range str_patterns {
		if strings.Contains(s,p) {
			return true
		}
	}

	const theshold_for_uniqueness = 12 // skjønn

	if len(s) > theshold_for_uniqueness {
		return true
	}

	return false
}

//****************************************************************************

func IsQuote(r rune) bool {

	switch r {
	case '"','\'',NON_ASCII_LQUOTE,NON_ASCII_RQUOTE:
		return true
	}

	return false
}

//****************************************************************************

func ReadToNext(array []rune,pos int,r rune) (string,int) {

	var buff []rune

	for i := pos; i < len(array); i++ {

		buff = append(buff,array[i])

		if i > pos && array[i] == r {
			ret := string(buff)
			return ret,len(ret)
		}
	}

	ret := string(buff)
	return ret,len(ret)
}


// **************************************************************************

func SearchTermLen(names []string) int {

	var maxlen int

	for _,s := range names {
		if !IsNPtrStr(s) && len(s) > maxlen {
			maxlen = len(s)
		}
	}

	return maxlen
}

// **************************************************************************

func IsNPtrStr(s string) bool {

	s = strings.TrimSpace(s)

	if s[0] == '(' && s[len(s)-1] == ')' {
		var a,b int = -1,-1
		fmt.Sscanf(s,"(%d,%d)",&a,&b)
		if a >= 0 && b >= 0 {
			return true
		}
	}
	return false
}

// **************************************************************************

func RunErr(message string) {

	const red = "\033[31;1;1m"
	const endred = "\033[0m"

	fmt.Println("SSTorytime",message,endred)

}

// **************************************************************************

func EscapeString(s string) string {

	run := []rune(s)
	var res []rune

	for r := range run {
		if run[r] == '\n' {
		} else if run[r] == '"' {
			res = append(res,'\\')
			res = append(res,'"')
		} else {
			res = append(res,run[r])
		}
	}

	s = string(res)
	return s
}

//******************************************************************

func ContextString(context []string) string {

	var s string

	for c := 0; c < len(context); c++ {

		s += context[c] + " "
	}

	return s
}

//****************************************************************************

func InList(s string, list []string) (int,bool) {

	for i,v := range list {
		if s == v {
			return i,true
		}
	}

	return -1,false
}

//****************************************************************************

func MatchArrows(arrows []ArrowPtr,arr ArrowPtr) bool {

	for a := range arrows {
		if arrows[a] == arr {
			return true
		}
	}

	return false
}

//****************************************************************************

func Arrow2Int(arr []ArrowPtr) []int {

	var ret []int

	for a := range arr {
		ret = append(ret,int(arr[a]))
	}

	return ret
}

//****************************************************************************

func MatchContexts(context1 []string,context2ptr ContextPtr) bool {

	if context1 == nil || context2ptr == 0 {
		return true
	}

	context2 := strings.Split(GetContext(context2ptr),",")

	for c := range context1 {

		if MatchesInContext(context1[c],context2) {
			return true
		}
	}

	return false 
}

//****************************************************************************

func MatchesInContext(s string,context []string) bool {
	
	for c := range context {
		if SimilarString(s,context[c]) {
			return true
		}
	}
	return false 
}


// **************************************************************************

func SimilarString(full,like string) bool {

	// Placeholder
	// Need to handle pluralisation patterns etc... multi-language

	if full == like {
		return true
	}

	if full == "" || like == "" || full == "any" || like == "any" {  // same as any
		return true
	}

	if strings.Contains(full,like) {
		return true
	}

	return false
}

// **************************************************************************

func SanitizePath(s string) string {

	re := regexp.MustCompile("[^a-zA-Z0-9]")
	s = re.ReplaceAllString(s, "_")

	brk := 0
	
	for i := 0; i < len(s); i++  {
		if s[i] != '_' {
			brk = i
			break
		}
	}

	return s[brk:]
}

// **************************************************************************

func GetURIFile(url string) (string,error) {

	// Get a remote file
	
	resp, err := http.Get(url)

	defer resp.Body.Close()

	if err != nil {
		return "",err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body),nil
}

//
// tools.go
//

