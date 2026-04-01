//****************************************************************************
//
// expand_node_vars.go
//
//****************************************************************************

package SSTorytime

import (
	"fmt"
	"strings"
	"time"
	_ "github.com/lib/pq"

)

//****************************************************************************

func ExpandDynamicFunctions(s string) string {

	if !strings.Contains(s,"{") {
		return s
	}

	if !strings.Contains(s,"}") {
		return s
	}

	chars := []rune(s[len("Dynamic:"):])

	var news string

	for pos := 0; pos < len(chars); pos++ {

		if chars[pos] != '{' {
			news += string(chars[pos])
		} else {
			newpos,result := EvaluateInBuilt(chars,pos)
			news += result
			pos = newpos
		}
	}

	return news
}

//****************************************************************************

func EvaluateInBuilt(chars []rune,pos int) (int,string) {

	var fntext string
	var endpos int

	for r := pos; chars[r] != '}' && r < len(chars); r++ {
		fntext += string(chars[r])
		endpos = r+1
	}

	fntext = fntext[1:len(fntext)]

	delim := func(c rune) bool {
		return c == ' ' || c == ',' || c == ';'
	}

	fn := strings.FieldsFunc(fntext,delim)
	result := DoInBuiltFunction(fn)
	return endpos,result
}

//****************************************************************************

func DoInBuiltFunction(fn []string) string {

	// Placeholder - this needs to support sandboxed read only user functions

	var result string

	switch fn[0] {
	case "TimeUntil":
		result = InBuiltTimeUntil(fn)
	case "TimeSince":
		result = InBuiltTimeSince(fn)
	}

	return result
}

//****************************************************************************

func InBuiltTimeUntil(fn []string) string {

	now := time.Now().Local()
	intended_time := GetTimeFromSemantics(fn,now)
	duration := intended_time.Sub(now)

	interval := int(duration / 1000000000)  // nanoseconds -> seconds

	years := interval / (365 * 24 * 3600)
	r1 := interval % (365 * 24 * 3600)

	days := r1 / (24 * 3600)
	r2 := r1 % (24 * 3600)

	hours := r2 / 3600
	r3 :=  r2 % 3600

	mins := r3 / 60

	return ShowTime(years,days,hours,mins)
}

//****************************************************************************

func InBuiltTimeSince(fn []string) string {

	now := time.Now().Local()
	intended_time := GetTimeFromSemantics(fn,now)

	duration := now.Sub(intended_time)

	interval := int(duration / 1000000000)  // nanoseconds -> seconds

	years := interval / (365 * 24 * 3600)
	r1 := interval % (365 * 24 * 3600)

	days := r1 / (24 * 3600)
	r2 := r1 % (24 * 3600)

	hours := r2 / 3600
	r3 :=  r2 % 3600

	mins := r3 / 60

	return ShowTime(years,days,hours,mins)
}

//****************************************************************************

func ShowTime(years,days,hours,mins int) string {

	var s string

	if years > 0 {
		s += fmt.Sprintf("%d Years, ",years)
	}

	if days > 0 {
		s += fmt.Sprintf("%d Days, ",days)
	}

	if hours > 0 {
		s += fmt.Sprintf("%d Hours, ",hours)
	}

	s += fmt.Sprintf("%d Mins ",mins)

	if mins < 0 {
		s += " [already passed or waiting for next occurrence]"
	}

	return s
}



//
// expand_node_vars.go
//
