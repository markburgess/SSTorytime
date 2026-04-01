// ******************************************************************
//
// service_search_cmd.go
//
// ******************************************************************

package SSTorytime

import (
	"fmt"
	"strings"
	"regexp"
	_ "github.com/lib/pq"

)


// ******************************************************************

type SearchParameters struct {

	Name     []string
	From     []string
	To       []string
	Chapter  string
	Context  []string
	Arrows   []string
	PageNr   int
	Range    int
	Min      []int
	Max      []int
	Finds    []string
	Sequence bool
	Stats    bool
	Horizon  int
}

// ******************************************************************
// Short Term Memory (STM) intent event capture
// ******************************************************************

type STM struct {

	Query SearchParameters

}

// ******************************************************************

const (

	CMD_ON = "\\on"
	CMD_ON_2 = "on"    // _2 are too short to be intentional
	CMD_FOR = "\\for"  // so double these for "smarter" accident avoidance
	CMD_FOR_2 = "for"
	CMD_ABOUT = "\\about"
	CMD_NOTES = "\\notes"
	CMD_BROWSE = "\\browse"
	CMD_PAGE = "\\page"
	CMD_PATH = "\\path"
	CMD_PATH2 = "\\paths"
	CMD_SEQ1 = "\\sequence"
	CMD_SEQ2 = "\\seq"
	CMD_STORY = "\\story"
	CMD_STORIES = "\\stories"
	CMD_FROM = "\\from"
	CMD_TO = "\\to"
	CMD_TO_2 = "to"
	CMD_CTX = "\\ctx"
	CMD_CONTEXT = "\\context"
	CMD_AS = "\\as"
	CMD_AS_2 = "as"
	CMD_CHAPTER = "\\chapter"
	CMD_CONTENTS = "\\contents"
	CMD_TOC = "\\toc"
	CMD_TOC_2 = "toc"
	CMD_MAP = "\\map"
	CMD_SECTION = "\\section"
	CMD_IN = "\\in"
	CMD_IN_2 = "in"
	CMD_ARROW = "\\arrow"
	CMD_ARROWS = "\\arrows"
	CMD_LIMIT = "\\limit"
	CMD_DEPTH = "\\depth"
	CMD_RANGE = "\\range"
	CMD_DISTANCE = "\\distance"
	CMD_STATS = "\\stats"
	CMD_STATS_2 = "stats"
	CMD_REMIND = "\\remind"
	CMD_HELP = "\\help"
	CMD_HELP_2 = "help"
	// What to find in orbit
	CMD_FINDS = "\\finds"
	CMD_FINDING = "\\finding"
	// bounding linear path and parallel arrows
	CMD_GT = "\\gt"
	CMD_LT = "\\lt"
	CMD_MIN = "\\min"	
	CMD_MAX = "\\max"
	CMD_ATLEAST = "\\atleast"
	CMD_ATMOST = "\\atmost"
	CMD_NEVER = "\\never"
	CMD_NEW = "\\new"

	RECENT = 4  // Four hours between a morning and afternoon
        NEVER = -1   // Haven't seen in this long
)

//******************************************************************
// Decoding local receiver (-) intent
//******************************************************************

func DecodeSearchField(cmd string) SearchParameters {

	var keywords = []string{ 
		CMD_NOTES, CMD_BROWSE,
		CMD_PATH,CMD_FROM,CMD_TO,CMD_TO_2,
		CMD_SEQ1,CMD_SEQ2,CMD_STORY,CMD_STORIES,
		CMD_CONTEXT,CMD_CTX,CMD_AS,CMD_AS_2,
		CMD_CHAPTER,CMD_IN,CMD_IN_2,CMD_SECTION,CMD_CONTENTS,CMD_TOC,CMD_TOC_2,CMD_MAP,
		CMD_ARROW,CMD_ARROWS,
		CMD_GT,CMD_MIN,CMD_ATLEAST,
		CMD_LT,CMD_MAX,CMD_ATMOST,
		CMD_ON,CMD_ON_2,CMD_ABOUT,CMD_FOR,CMD_FOR_2,
		CMD_PAGE,
		CMD_LIMIT,CMD_RANGE,CMD_DISTANCE,CMD_DEPTH,
		CMD_STATS,CMD_STATS_2,
		CMD_REMIND,CMD_NEVER,CMD_NEW,
		CMD_HELP,CMD_HELP_2,
		CMD_FINDS,CMD_FINDING,
        }
	
	// parentheses are reserved for unaccenting

	cmd = strings.ToLower(cmd)

	m := regexp.MustCompile("[ \t]+") 
	cmd = m.ReplaceAllString(cmd," ") 

	cmd = strings.TrimSpace(cmd)
	pts := SplitQuotes(cmd)

	var parts [][]string
	var part []string

	for p := 0; p < len(pts); p++ {

		subparts := SplitQuotes(pts[p])

		for w := 0; w < len(subparts); w++ {

			if IsCommand(subparts[w],keywords) {
				// special case for TO with implicit FROM, and USED AS
				if p > 0 && subparts[w] == "to" {
					part = append(part,subparts[w])
					continue
				}
				if w > 0 && strings.HasPrefix(subparts[w],"to") {
					part = append(part,subparts[w])
				} else {
					parts = append(parts,part)
					part = nil
					part = append(part,subparts[w])
				}
			} else {
				// Try to override command line splitting behaviour
				part = append(part,subparts[w])
			}
		}
	}

	parts = append(parts,part) // add straggler to complete

	// command is now segmented

	param := FillInParameters(parts,keywords)

	for arg := range param.Name {

		isdirac,beg,end,cnt := DiracNotation(param.Name[arg])
		
		if isdirac {
			param.Name = nil
			param.From = []string{beg}
			param.To = []string{end}
			param.Context = []string{cnt}
			break
		}
	}

	return param
}

//******************************************************************

func FillInParameters(cmd_parts [][]string,keywords []string) SearchParameters {

	var param SearchParameters 

	for c := 0; c < len(cmd_parts); c++ {

		lenp := len(cmd_parts[c])

		for p := 0; p < lenp; p++ {

			switch SomethingLike(cmd_parts[c][p],keywords) {

			case CMD_STATS, CMD_STATS_2:
				param.Stats = true
				continue

			case CMD_HELP, CMD_HELP_2:
				param.Chapter = "SSTorytime help"
				param.Name = []string{"any"}
				continue

			case CMD_CHAPTER,CMD_SECTION,CMD_IN,CMD_IN_2,CMD_CONTENTS,CMD_TOC,CMD_TOC_2,CMD_MAP:

				if lenp > p+1 {
					str := cmd_parts[c][p+1]
					str = strings.TrimSpace(str)
					str = strings.Trim(str,"'")
					str = strings.Trim(str,"\"")
					if str == "any" {
						str = "%%"
					}
					param.Chapter = str
					break
				} else {
					param.Chapter = "TableOfContents"
					break					
				}
				continue

			case CMD_NOTES, CMD_BROWSE:
				if param.PageNr < 1 {
					param.PageNr = 1
				}

				if lenp > p+1 {
					if cmd_parts[c][p+1] == "any" {
						param.Chapter = "%%"
					} else {
						param.Chapter = cmd_parts[c][p+1]
					}
				} else {
					if lenp > 1 {
						param = AddOrphan(param,cmd_parts[c][p+1])
					}
				}
				continue

			case CMD_PAGE:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.PageNr = no
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_RANGE,CMD_DEPTH,CMD_LIMIT,CMD_DISTANCE:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.Range = no
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_GT,CMD_MIN,CMD_ATLEAST:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.Min = append(param.Min,no)
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_LT,CMD_MAX,CMD_ATMOST:
				// if followed by a number, else could be search term
				if lenp > p+1 {
					p++
					var no int = -1
					fmt.Sscanf(cmd_parts[c][p],"%d",&no)
					if no > 0 {
						param.Max = append(param.Max,no)
					} else {
						param = AddOrphan(param,cmd_parts[c][p-1])
						param = AddOrphan(param,cmd_parts[c][p])
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_ARROW,CMD_ARROWS:
				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							param.Arrows = append(param.Arrows,DeQ(ult[u]))
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue
				
				case CMD_CONTEXT,CMD_CTX,CMD_AS,CMD_AS_2:
				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							class := strings.TrimSpace(DeQ(ult[u]))
							if len(class) > 0 {
								param.Context = append(param.Context,class)
							}
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

		case CMD_PATH,CMD_PATH2,CMD_FROM:
				if lenp-1 == p {
					// redundant word if empty
					continue
				}

				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						if !IsLiteralNptr(cmd_parts[c][pp]) {
							ult := strings.Split(cmd_parts[c][pp],",")
							for u := range ult {
								param.From = append(param.From,DeQ(ult[u]))
							}
						} else {
							param.From = append(param.From,cmd_parts[c][pp])
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_TO,CMD_TO_2:
				if p > 0 && lenp > p+1 {
					if param.From == nil {
						param.From = append(param.From,cmd_parts[c][p-1])
					}

					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						if !IsLiteralNptr(cmd_parts[c][pp]) {
							ult := strings.Split(cmd_parts[c][pp],",")
							for u := range ult {
								param.To = append(param.To,DeQ(ult[u]))
							}
						} else {
							param.To = append(param.From,cmd_parts[c][pp])
						}
					}
					continue
				}
				// TO is too short to be an independent search term

				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						ult := strings.Split(cmd_parts[c][pp],",")
						for u := range ult {
							param.To = append(param.To,DeQ(ult[u]))
						}
					}
					continue
				}

			case CMD_SEQ1,CMD_SEQ2,CMD_STORY,CMD_STORIES:
				param.Sequence = true
				continue

			case CMD_NEW:
				param.Horizon = RECENT
				continue
			case CMD_NEVER:
				param.Horizon = NEVER
				continue

			case CMD_ON,CMD_ON_2,CMD_ABOUT,CMD_FOR,CMD_FOR_2:
				if lenp > p+1 {
					for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
						p++
						if param.PageNr > 0 {
							param.Chapter = cmd_parts[c][pp]
						} else {
							ult := strings.Split(cmd_parts[c][pp]," ")
							for u := range ult {
								param.Name = append(param.Name,DeQ(ult[u]))
							} 
						}
					}
				} else {
					param = AddOrphan(param,cmd_parts[c][p])
				}
				continue

			case CMD_FINDS,CMD_FINDING:

				for pp := p+1; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
					p++
					ult := SplitQuotes(cmd_parts[c][pp])
					for u := range ult {
						if ult[u] == "any" {
							ult[u] = "%%"
						}
						param.Finds = append(param.Finds,DeQ(ult[u]))
					}
				}
				continue

			default:

				if lenp > p+1 && cmd_parts[c][p+1] == CMD_TO {
					continue
				}

				for pp := p; IsParam(pp,lenp,cmd_parts[c],keywords); pp++ {
					p++
					ult := SplitQuotes(cmd_parts[c][pp])
					for u := range ult {
						if ult[u] == "any" {
							ult[u] = "%%"
						}
						param.Name = append(param.Name,DeQ(ult[u]))
					}
				}
				continue
			}
			break
		}
	}

	var rnames []string
	var wildcards bool

	// If there are wildcards AND other matches, these are redundant so remove any/%%

	for _,term := range param.Name {
		if term == "%%" || term == "any" {
			wildcards = true
		} else {
			rnames = append(rnames,term)
		}
	}

	if wildcards && len(rnames) > 0 {
		param.Name = rnames
	}

	return param
}

//******************************************************************

func IsParam(i,lenp int,keys []string,keywords []string) bool {

	// Make sure the next item is not the start of a new token

	const min_sense = 4

	if i >= lenp {
		return false
	}

	key := keys[i]

	if IsCommand(key,keywords) {
		return false
	}

	return true
}

//******************************************************************

func MinMaxPolicy(search SearchParameters) (int,int) {

	// The min max doubles as context dependent role as
	// i) limits on path length and ii) limits on matches arrow matches

	minlimit := 1
	maxlimit := 0
	from := search.From != nil
	to := search.To != nil

	// Validate

	if len(search.Min) > 4 {
		fmt.Println("\nWARNING: minimum arrow matches exceeds the number of ST-types")
	}

	if len(search.Max) > 4 {
		fmt.Println("\nWARNING: maximum arrow matches exceeds the number of ST-types")
	}

	if len(search.Min) != 4 && len(search.Max) != 4 {

		// Only "abusing" the min max for linear path length or search depth

		if len(search.Min) > 0 && len(search.Max) > 0 {
			if search.Min[0] > search.Max[0] {
				fmt.Println("\nWARNING: minimum arrow limit greater than maximum limit!")
				fmt.Println("Depth/range:","min =",search.Min[0],", max =",search.Max[0])
			}
			
			if len(search.Min) == 1 {
				minlimit = search.Min[0]
			}
			
		} else if len(search.Max) == 1 && search.Range > 0 {
			fmt.Println("\nWARNING: conflict between \\depth,\\range and \\max,\\lt,\\atmost ")
		}
	} else {
		// Full ST-type arrow match limits

		for i := 0; i < 4; i++ {
			if search.Min[i] > search.Max[i] {
				fmt.Println("\nWARNING: minimum arrow limit greater than maximum limit!")
				fmt.Println("ST-type:",i,"min =",search.Min[i],", max =",search.Max[i])
			}
		}
	}

	// Defaults

	if search.Chapter == "TableOfContents" {

		// We want to see all contents
		maxlimit = 50

	} else if search.Range > 0 {

		maxlimit = search.Range

	} else if len(search.Max) == 1 {  // if only one, we probably meant Range

		maxlimit = search.Max[0]

	} else {

		if from || to || search.Sequence {
			maxlimit = 30 // many paths make hard work
		} else {
			const common_word = 5

			if SearchTermLen(search.Name) < common_word {
				maxlimit = 5
			} else {
				maxlimit = 10
			}

			if len(search.Name) < 3 && AllExact(search.Name) {
				maxlimit = 30
			}
		}
	}

	return minlimit,maxlimit
}

//******************************************************************

func AllExact(list []string) bool {

	is_exact := false

	for _,s := range list {
		is,_ := IsExactMatch(s)
		is_exact = is_exact || is
	}

	return is_exact
}

//******************************************************************

func IsLiteralNptr(s string) bool {
	
	var a,b int = -1,-1

	s = strings.TrimSpace(s)
	
	fmt.Sscanf(s,"(%d,%d)",&a,&b)
	
	if a >= 0 && b >= 0 {
		return true
	}
	
	return false
}

//******************************************************************

func SomethingLike(s string,keywords []string) string {

	const min_sense = 4

	for k := 0; k < len(keywords); k++ {

		if s == keywords[k] {
			return keywords[k]
		}

		if len(s) > min_sense && len(keywords[k]) > min_sense {
			if strings.HasPrefix(s,keywords[k]) {
				return keywords[k]
			}
		}
	}
	return s
}

//******************************************************************

func CheckHelpQuery(name string) string {

	if name == "\\help" {
		name = "\\notes \\chapter \"help and search\" \\limit 40"
	}
	
	return name
}

//******************************************************************

func CheckNPtrQuery(name,nclass,ncptr string) string {

	if name == "" && len(nclass) > 0 && len(ncptr) > 0 {
		// direct click on an item
		var a, b int
		fmt.Sscanf(nclass, "%d", &a)
		fmt.Sscanf(ncptr, "%d", &b)
		nstr := fmt.Sprintf("(%d,%d)", a, b)
		name = name + nstr
	}

	return name
}

//******************************************************************

func CheckRemindQuery(name string) string {

	if len(name) == 0 || name == "\\remind" {
		ambient, key, _ := GetTimeContext()
		name = "any \\chapter reminders \\context any, " + key + " " + ambient + " \\limit 20"
	}

	return name
}

//******************************************************************

func CheckConceptQuery(name string) string {

	if strings.Contains(name,"\\dna ") {
		repl := "any \\arrow " + INV_CONT_FRAG_IN_S + " \\limit 20 "
		name = strings.Replace(name, "\\dna ",repl,-1)
		return name
	}

	if strings.Contains(name,"\\concept ") {
		repl := "any \\arrow " + INV_CONT_FRAG_IN_S + " \\limit 20 "
		name = strings.Replace(name, "\\concept ",repl,-1)
		return name
	}

	if strings.Contains(name,"\\concepts ") {
		repl := "any \\arrow " + INV_CONT_FRAG_IN_S + " \\limit 20 "
		name = strings.Replace(name, "\\concepts ",repl,-1)
		return name
	}

	if strings.Contains(name,"\\terms ") {
		repl := "any \\arrow " + INV_CONT_FRAG_IN_S + " \\limit 20 "
		name = strings.Replace(name, "\\terms ",repl,-1)
		return name
	}

	return name
}

//******************************************************************

func IsCommand(s string,list []string) bool {

	const min_sense = 5

	for w := range list {
		if list[w] == s {
			return true
		}

		// Allow likely abbreviations ?

		if len(list[w]) > min_sense && strings.HasPrefix(s,list[w]) {
			return true
		}
	}
	return false
}

//******************************************************************

func AddOrphan(param SearchParameters,orphan string) SearchParameters {

	// if a keyword isn't followed by the right param it was possibly
	// intended as a search term not a command, so add back

	if param.To != nil {
		param.To = append(param.To,orphan)
		return param
	}

	if param.From != nil {
		param.From = append(param.From,orphan)
		return param
	}

	param.Name = append(param.Name,orphan)

	return param
}

//******************************************************************

func SplitQuotes(s string) []string {

	var items []string
	var upto []rune
	cmd := []rune(s)

	for r := 0; r < len(cmd); r++ {

		if IsQuote(cmd[r]) {
			if len(upto) > 0 {
				items = append(items,string(upto))
			}

			qstr,offset := ReadToNext(cmd,r,cmd[r])

			if len(qstr) > 0 {
				items = append(items,qstr)
				r += offset
			}
			continue
		}

		switch cmd[r] {
		case ' ':
			if len(upto) > 0 {
				items = append(items,string(upto))
			}
			upto = nil
			continue

		case '(':
			if len(upto) > 0 {
				items = append(items,string(upto))
			}

			qstr,offset := ReadToNext(cmd,r,')')

			if len(qstr) > 0 {
				items = append(items,qstr)
				r += offset
			}
			continue

		}

		upto = append(upto,cmd[r])
	}

	if len(upto) > 0 {
		items = append(items,string(upto))
	}

	return items
}

// **************************************************************************

func DeQ(s string) string {

	return strings.Trim(s,"\"")
}



//
// service_search_cmd.go
//

