// **************************************************************************
//
// eval_context.go
//
// **************************************************************************

package SSTorytime

import (
	"fmt"
	"strings"
	"time"
	_ "github.com/lib/pq"

)


// **************************************************************************

func GetContext(contextptr ContextPtr) string {

	exists := int(contextptr) < len(CONTEXT_DIRECTORY)

	if exists {
		return CONTEXT_DIRECTORY[contextptr].Context
	}

	return "unknown context"
}

// ****************************************************************************

func RegisterContext(parse_state map[string]bool,context []string) ContextPtr {

	ctxstr := NormalizeContextString(parse_state,context)

	if len(ctxstr) == 0 {
		return 0
	}

	ctxptr,exists := CONTEXT_DIR[ctxstr] 

	if !exists {
		var cd ContextDirectory
		cd.Context = ctxstr
		cd.Ptr = CONTEXT_TOP
		CONTEXT_DIRECTORY = append(CONTEXT_DIRECTORY,cd)
		CONTEXT_DIR[ctxstr] = CONTEXT_TOP
		ctxptr = CONTEXT_TOP
		CONTEXT_TOP++
	}

	return ctxptr
}

// **************************************************************************

func TryContext(sst PoSST,context []string) ContextPtr {

	ctxstr := CompileContextString(context)
	str,ctxptr := GetDBContextByName(sst,ctxstr)

	if ctxptr == -1 || str != ctxstr {
		ctxptr = UploadContextToDB(sst,ctxstr,-1)
		RegisterContext(nil,context)
	}

	return ctxptr
}

// **************************************************************************

func CompileContextString(context []string) string {

	// Ensure idempotence

	var merge = make(map[string]int)

	for c := range context {
		merge[context[c]]++
	}

	return List2String(Map2List(merge))	
}

// **************************************************************************

func NormalizeContextString(contextmap map[string]bool,ctx []string) string {

	// Mitigate combinatoric explosion

	var merge = make(map[string]bool)
	var clist []string

	// Merge sources into single map
	
	if contextmap != nil {
		for c := range contextmap {
			merge[c] = true
		}
	}

	for c := range ctx {
		merge[ctx[c]] = true
	}

	for c := range merge {
		s := strings.Split(c,",")
		for i := range s {
			s[i] = strings.TrimSpace(s[i])
			if s[i] != "_sequence_" {
				clist = append(clist,s[i])
			}
		}
	}

	return List2String(clist)
}

// **************************************************************************

func GetNodeContext(sst PoSST,node Node) []string {

	str := GetNodeContextString(sst,node)

	if str != "" {
		return strings.Split(str,",")
	}

	return nil
}

// **************************************************************************

func GetNodeContextString(sst PoSST,node Node) string {

	// This reads the ghost link planted for the purpose of attaching
	// a context to floating nodes

	empty := GetDBArrowByName(sst,"empty")

	for _,lnk := range node.I[ST_ZERO+LEADSTO] {

		if lnk.Arr == empty {
			return GetContext(lnk.Ctx)
		}
	}

	return ""
}

// **************************************************************************
// Dynamic context
// **************************************************************************

var GR_DAY_TEXT = []string{
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
        "Sunday",
    }
        
var GR_MONTH_TEXT = []string{
	"NONE",
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
}
        
var GR_SHIFT_TEXT = []string{
        "Night",
        "Morning",
        "Afternoon",
        "Evening",
    }

// For second resolution Unix time

const CF_MONDAY_MORNING = 345200
const CF_MEASURE_INTERVAL = 5*60
const CF_SHIFT_INTERVAL = 6*3600

const MINUTES_PER_HOUR = 60
const SECONDS_PER_MINUTE = 60
const SECONDS_PER_HOUR = (60 * SECONDS_PER_MINUTE)
const SECONDS_PER_DAY = (24 * SECONDS_PER_HOUR)
const SECONDS_PER_WEEK = (7 * SECONDS_PER_DAY)
const SECONDS_PER_YEAR = (365 * SECONDS_PER_DAY)
const HOURS_PER_SHIFT = 6
const SECONDS_PER_SHIFT = (HOURS_PER_SHIFT * SECONDS_PER_HOUR)
const SHIFTS_PER_DAY = 4
const SHIFTS_PER_WEEK = (4*7)

// ****************************************************************************
// Semantic spacetime timeslots, CFEngine style
// ****************************************************************************

func DoNowt(then time.Time) (string,string) {

	//then := given.UnixNano()

	// Time on the torus (donut/doughnut) (CFEngine style)
	// The argument is a Golang time unit e.g. then := time.Now()
	// Return a db-suitable keyname reflecting the coarse-grained SST time
	// The function also returns a printable summary of the time

	// In this version, we need less accuracy and greater semantic distinction
	// so prefix temporal classes with a :

	year := fmt.Sprintf("Yr%d",then.Year())
	month := GR_MONTH_TEXT[int(then.Month())]
	day := then.Day()
	hour := fmt.Sprintf("Hr%02d",then.Hour())
	quarter := fmt.Sprintf("Qu%d",then.Minute()/15 + 1)
	shift :=  fmt.Sprintf("%s",GR_SHIFT_TEXT[then.Hour()/6])

	//secs := then.Second()
	//nano := then.Nanosecond()
	//mins := fmt.Sprintf("Min%02d",then.Minute())

	n_season,s_season := Season(month)

	dayname := then.Weekday()
	dow := fmt.Sprintf("%.3s",dayname)
	daynum := fmt.Sprintf("Day%d",day)

	// 5 minute resolution capture is too fine grained for most human interest
        interval_start := (then.Minute() / 5) * 5
        interval_end := (interval_start + 5) % 60
        minD := fmt.Sprintf("Min%02d_%02d",interval_start,interval_end)

	// Don't include the time key in general context, as it varies too fast to be meaningful

	var when string = fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s, %s, %s",n_season,s_season,shift,dayname,daynum,month,year,hour,quarter)
	var key string = fmt.Sprintf("%s:%s:%s-%s",dow,hour,quarter,minD)

	return when, key
}

// ****************************************************************************

func GetTimeContext() (string,string,int64) {

	now := time.Now()
	context,keyslot := DoNowt(now)

	return context,keyslot,now.Unix()
}

// ****************************************************************************

func Season (month string) (string,string) {

	switch month {

	case "December","January","February":
		return "N_Winter","S_Summer"
	case "March","April","May":
		return "N_Spring","S_Autumn"
	case "June","July","August":
		return "N_Summer","S_Winter"
	case "September","October","November":
		return "N_Autumn","S_Spring"
	}

	return "hurricane","typhoon"
}

// ****************************************************************************
 
func GetTimeFromSemantics(speclist []string,now time.Time) time.Time {

	day := 0
	hour := 0
	mins := 0
	weekday := 0
	month := time.Month(0)
	year := 0
	days_to_next := 0

	hasweekday := false
	hasmonth := false

	// Parse semantic time array

	for i,v := range speclist {

		if i == 0 {
			continue
		}

		if strings.HasPrefix(v,"Day") {
			fmt.Sscanf(v[3:],"%d",&day)
			continue
		}

		if strings.HasPrefix(v,"Yr") {
			fmt.Sscanf(v[2:],"%d",&year)
			continue
		}

		if strings.HasPrefix(v,"Min") {
			fmt.Sscanf(v[3:],"%d",&mins)
			continue
		}

		if strings.HasPrefix(v,"Hr") {
			fmt.Sscanf(v[2:],"%d",&hour)
			continue
		}

		if !hasweekday {
			weekday,hasweekday = InList(v,GR_DAY_TEXT)
			if hasweekday {
				intended := weekday
				todayis := fmt.Sprintf("%s",now.Weekday())
				actual,_ := InList(todayis,GR_DAY_TEXT)
				days_to_next = (intended - actual + 7) % 7
				continue
			}
		}

		if !hasmonth {
			var index int
			index,hasmonth = InList(v,GR_MONTH_TEXT)
			if hasmonth {
				month = time.Month(index)
				continue
			}
		}

		fmt.Println("Semantic time parameter without semantic prefix (Day,Hr,Min, etc)",v)
	}

	if hasweekday && (day > 0 || hasmonth || year > 0) {
		fmt.Println("Weekday only makes sense as the next applicable occurrence, without a date")
	} else if hasweekday {

		// We're looking for the next upcoming day
		day = now.Day()
		month = now.Month()
		year = now.Year()
		newnow := time.Date(year,month,day,0,0,0,0,time.UTC)
		newnow = newnow.AddDate(0,0,days_to_next)
		return newnow
	}

	if year == 0 {
		year = now.Year()
	}

	if day == 0 {
		day = now.Day()
	}

	if month == 0 {
		month = now.Month()
	}

	if hour == 0 {
		hour = now.Hour()
	}

	// Note the local timezone is very problematic in Go

	_,offset_secs := now.Zone()

	offset := offset_secs / 3600

	newnow := time.Date(year,month,day,hour-offset,mins,0,0,time.UTC)
	return newnow
}



//
// END eval_context.go
//
