// **************************************************************************
//
// json_marshalling.go
//
// **************************************************************************


package SSTorytime

import (
	"fmt"
	"sync"
	"encoding/json"
	_ "github.com/lib/pq"

)

// **************************************************************************

func JSONNodeEvent(sst PoSST, nptr NodePtr,xyz Coords,orbits [ST_TOP][]Orbit) NodeEvent {

	node := GetDBNodeByNodePtr(&sst,nptr)

	var event NodeEvent
	event.Text = node.S
	event.L = node.L
	event.Chap = node.Chap
	event.Context = GetNodeContextString(&sst,node)
	event.NPtr = nptr
	event.XYZ = xyz
	event.Orbits = orbits
	return event
}

// **************************************************************************

func LinkWebPaths(sst *PoSST,cone [][]Link,nth int,chapter string,context []string,swimlanes,limit int) [][]WebPath {

	// This is dealing in good faith with one of swimlanes cones, assigning equal width to all
	// The cone is a flattened array, we can assign spatial coordinates for visualization

	var conepaths [][]WebPath

	directory := AssignConeCoordinates(cone,nth,swimlanes)

	// JSONify the cone structure, converting []Link into []WebPath

	for p := 0; p < len(cone); p++ {

		path_start := GetDBNodeByNodePtr(sst,cone[p][0].Dst)		
		
		start_shown := false

		var path []WebPath
		
		for l := 1; l < len(cone[p]); l++ {

			if !MatchContexts(sst,context,cone[p][l].Ctx) {
				break
			}

			nextnode := GetDBNodeByNodePtr(sst,cone[p][l].Dst)

			if !SimilarString(nextnode.Chap,chapter) {
				break
			}
			
			if !start_shown {
				var ws WebPath
				ws.Name = path_start.S
				ws.NPtr = cone[p][0].Dst
				ws.Chp = nextnode.Chap
				ws.XYZ = directory[cone[p][0].Dst]
				ws.Wgt = cone[p][0].Wgt
				path = append(path,ws)
				start_shown = true
			}

			arr := GetDBArrowByPtr(sst,cone[p][l].Arr)
	
			if l < len(cone[p]) {
				var wl WebPath
				wl.Name = arr.Long
				wl.Arr = cone[p][l].Arr
				wl.STindex = arr.STAindex
				wl.XYZ = directory[cone[p][l].Dst]
				wl.Wgt = cone[p][l].Wgt
				path = append(path,wl)
			}

			var wn WebPath
			wn.Name = nextnode.S
			wn.Chp = nextnode.Chap
			wn.NPtr = cone[p][l].Dst
			wn.XYZ = directory[cone[p][l].Dst]
			path = append(path,wn)

		}
		conepaths = append(conepaths,path)
	}

	return conepaths
}

// **************************************************************************

func JSONPage(sst PoSST, maplines []PageMap) string {

	var webnotes PageView
	var lastchap,lastctx string
	var signalchap, signalctx, signalchange string
	var warned bool = false

	directory := AssignPageCoordinates(maplines)

	for n := 0; n < len(maplines); n++ {

		var path []WebPath

		txtctx := GetContext(&sst,maplines[n].Context)

		// Format superheader aggregate summary

		if lastchap != maplines[n].Chapter {
			if !warned {
				webnotes.Title = webnotes.Title
				warned = true
			}
			webnotes.Title += maplines[n].Chapter + ", "
			lastchap = maplines[n].Chapter
			signalchap = maplines[n].Chapter
		} else {
			signalchap = ""
		}

		if lastctx != txtctx {
			webnotes.Context += txtctx + ", " 
			lastctx = txtctx
			signalctx = txtctx
		} else {
			signalctx = txtctx
		}

		signalchange = signalchap + " :: " + signalctx

		// Next line item

		for lnk := 0; lnk < len(maplines[n].Path); lnk++ {

			text := GetDBNodeByNodePtr(&sst,maplines[n].Path[lnk].Dst)

			if lnk == 0 {
				var ws WebPath
				ws.Name = text.S
				ws.NPtr = maplines[n].Path[lnk].Dst
				ws.XYZ = directory[ws.NPtr]
				ws.Chp = maplines[n].Chapter
				ws.Line = maplines[n].Line
				ws.Ctx = GetContext(&sst,maplines[n].Context)
				path = append(path,ws)
				
			} else {// ARROW
				arr := GetDBArrowByPtr(&sst,maplines[n].Path[lnk].Arr)
				var wl WebPath
				wl.Name = arr.Long
				wl.Arr = maplines[n].Path[lnk].Arr
				wl.STindex = arr.STAindex
				path = append(path,wl)
				// NODE
				var ws WebPath
				ws.Name = text.S
				ws.NPtr = maplines[n].Path[lnk].Dst
				ws.XYZ = directory[ws.NPtr]
				ws.Chp = maplines[n].Chapter
				ws.Ctx = signalchange
				path = append(path,ws)
			}
		}
		// Next line
		webnotes.Notes = append(webnotes.Notes,path)
	}
	
	encoded, _ := json.Marshal(webnotes)
	jstr := fmt.Sprintf("%s",string(encoded))

	return jstr
}


// **************************************************************************

func GetNodeOrbit(sst *PoSST,nptr NodePtr,exclude_vector string,limit int) [ST_TOP][]Orbit {

	// radius = 0 is the starting node

	const probe_radius = 3

	// Find the orbiting linked nodes of NPtr, start with properties of node

	sweep,_ := GetEntireConePathsAsLinks(sst,"any",nptr,probe_radius,limit)

	var satellites [ST_TOP][]Orbit
	var thread_wg sync.WaitGroup

	for stindex := 0; stindex < ST_TOP; stindex++ {

		// Go routines remain a mystery
		thread_wg.Add(1)
		
		go func(idx int) {
			defer thread_wg.Done()  // threading
			
			satellites[idx] = AssembleSatellitesBySTtype(sst,idx,satellites[idx],sweep,exclude_vector,probe_radius,limit)
			
		} (stindex)
	}
	
	thread_wg.Wait()

	return satellites
}

// **************************************************************************

func AssembleSatellitesBySTtype(sst *PoSST,stindex int,satellite []Orbit,sweep [][]Link,exclude_vector string,probe_radius int,limit int) []Orbit {

	var already = make(map[string]bool)

	// Sweep different radial paths [angle][depth]

	for angle := 0; angle < len(sweep); angle++ {
		
		// len(sweep[angle]) is the length of the probe path at angle
		
		if sweep[angle] != nil && len(sweep[angle]) > 1 {
			
			const nearest_satellite = 1
			start := sweep[angle][nearest_satellite]
			
			arrow := GetDBArrowByPtr(sst,start.Arr)
			
			if arrow.STAindex == stindex {

				txt := GetDBNodeByNodePtr(sst,start.Dst)

				var nt Orbit				
				nt.Arrow = arrow.Long
				nt.STindex = arrow.STAindex
				nt.Dst = start.Dst
				nt.Wgt = start.Wgt
				nt.Text = txt.S
				nt.Ctx = GetContext(sst,start.Ctx)
				nt.Radius = 1
				if arrow.Long == exclude_vector || arrow.Short == exclude_vector {
					continue
				}

				satellite = IdempAddSatellite(satellite,nt,already)
				
				// are there more satellites at this angle?
				
				for depth := 2; depth < probe_radius && depth < len(sweep[angle]); depth++ {
					
					arprev := STIndexToSTType(arrow.STAindex)
					next := sweep[angle][depth]
					arrow = GetDBArrowByPtr(sst,next.Arr)
					subtxt := GetDBNodeByNodePtr(sst,next.Dst)
					
					if arrow.Long == exclude_vector || arrow.Short == exclude_vector {
						break
					}

					nt.Arrow = arrow.Long
					nt.STindex = arrow.STAindex
					nt.Dst = next.Dst
					nt.Wgt = next.Wgt
					nt.Ctx = GetContext(sst,next.Ctx)
					nt.Text = subtxt.S
					nt.Radius = depth
					
					arthis := STIndexToSTType(arrow.STAindex)
					// No backtracking
					if arthis != -arprev {	
						satellite = IdempAddSatellite(satellite,nt,already)
						arprev = arthis
					}
				}
			}
		}
	}

	return satellite
}

// **************************************************************************

func IdempAddSatellite(list []Orbit, item Orbit,already map[string]bool) []Orbit {

	// crude check but effective, since the list is fairly short unless the graph is sick

	key := fmt.Sprintf("%v,%s",item.Dst,item.Arrow)

	if already[key] {
		return list
	} else {
		already[key] = true
		return append(list,item)
	}
}

// **************************************************************************

func GetLongestAxialPath(sst *PoSST,nptr NodePtr,arrowptr ArrowPtr,limit int) []Link {

	// Used in story search along extended STtype paths

	var max int = 1

	sttype := STIndexToSTType(sst.ARROW_DIRECTORY[arrowptr].STAindex)

	paths,dim := GetFwdPathsAsLinks(sst,nptr,sttype,limit,limit)

	for pth := 0; pth < dim; pth++ {

		var depth int
		paths[pth],depth = TruncatePathsByArrow(paths[pth],arrowptr)

		if len(paths[pth]) == 1 {
			paths[pth] = nil
		}

		if depth > max {
			max = pth
		}
	}

	return paths[max]
}

// **************************************************************************

func TruncatePathsByArrow(path []Link,arrow ArrowPtr) ([]Link,int) {

	for hop := 1; hop < len(path); hop++ {

		if path[hop].Arr != arrow {
			return path[:hop],hop
		}
	}

	return path,len(path)
}





//
// json_marshalling.go
//
