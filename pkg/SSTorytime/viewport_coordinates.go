// **************************************************************************
//
// viewport_coordinates.go
//
// **************************************************************************

package SSTorytime

import (
       "math"
)

// **************************************************************************

const R0 = 0.4    // radii should not overlap
const R1 = 0.3
const R2 = 0.1

// **************************************************************************

func RelativeOrbit(origin Coords,radius float64,n int,max int) Coords {

	var xyz Coords
	var offset float64

	// splay the vector positions so links not collinear
	switch radius {
	case R1:
		offset = -math.Pi/6.0
	case R2:
		offset = +math.Pi/6.0
	}

	angle := offset + 2 * math.Pi * float64(n)/float64(max)

	xyz.X = origin.X + float64(radius * math.Cos(angle))
	xyz.Y = origin.Y + float64(radius * math.Sin(angle))
	xyz.Z = origin.Z

	return xyz
}

// **************************************************************************

func SetOrbitCoords(xyz Coords,orb [ST_TOP][]Orbit) [ST_TOP][]Orbit {
	
	var r1max,r2max int
	
	// Count all the orbital nodes at this location to calc space
	
	for sti := 0; sti < ST_TOP; sti++ {
		
		for o := range orb[sti] {
			switch orb[sti][o].Radius {
			case 1:
				r1max++
			case 2:
				r2max++
			}
		}
	}
	
	// Place + and - cones on opposite sides, by ordering of sti
	
	var r1,r2 int
	
	for sti := 0; sti < ST_TOP; sti++ {
		
		for o := 0; o < len(orb[sti]); o++ {
			if orb[sti][o].Radius == 1 {
				anchor := RelativeOrbit(xyz,R1,r1,r1max)
				orb[sti][o].OOO = xyz
				orb[sti][o].XYZ = anchor
				r1++
				for op := o+1; op < len(orb[sti]) && orb[sti][op].Radius == 2; op++ {
					orb[sti][op].OOO = anchor
					orb[sti][op].XYZ = RelativeOrbit(anchor,R2,r2,r2max)
					r2++
					o = op-1
				}
			}
		}
	}

	return orb
}

// **************************************************************************

func AssignConeCoordinates(cone [][]Link,nth,swimlanes int) map[NodePtr]Coords {

	var unique = make([][]NodePtr,0)
	var already = make(map[NodePtr]bool)
	var maxlen_tz int

	// If we have multiple cones, each needs a separate name/graph space in X

	if swimlanes == 0 {
		swimlanes = 1
	}

	// Find the longest path length

	for x := 0; x < len(cone); x++ {
		if len(cone[x]) > maxlen_tz {
			maxlen_tz = len(cone[x])
		}
	}

	// Count the expanding wavefront sections for unique node entries

	XChannels := make([]float64,maxlen_tz) // node widths along each path step

	// Find the total number of parallel swimlanes

	for tz := 0; tz < maxlen_tz; tz++ {
		var unique_section = make([]NodePtr,0)
		for x := 0; x < len(cone); x++ {
			if tz < len(cone[x]) {
				if !already[cone[x][tz].Dst] {
					unique_section = append(unique_section,cone[x][tz].Dst)
					already[cone[x][tz].Dst] = true
					XChannels[tz]++
				}
			}
		}
		unique = append(unique,unique_section)
	}

	return MakeCoordinateDirectory(XChannels,unique,maxlen_tz,nth,swimlanes)
}

// **************************************************************************

func AssignStoryCoordinates(axis []Link,nth,swimlanes int,limit int, already map[NodePtr]bool) map[NodePtr]Coords {

	var unique = make([][]NodePtr,0)

	// Nth is segment nth of swimlanes, which has range (width=1.0)/swimlanes * [nth-nth+1]

	if swimlanes == 0 {
		swimlanes = 1
	}

	maxlen_tz := len(axis)

	if limit < maxlen_tz {
		maxlen_tz = limit
	}

	XChannels := make([]float64,maxlen_tz)        // node widths along the path

	for tz := 0; tz < maxlen_tz; tz++ {

		var unique_section = make([]NodePtr,0)	

		if !already[axis[tz].Dst] {
			unique_section = append(unique_section,axis[tz].Dst)
			already[axis[tz].Dst] = true
			XChannels[tz]++
		}

		unique = append(unique,unique_section)
	}

	return MakeCoordinateDirectory(XChannels,unique,maxlen_tz,nth,swimlanes)
}

// **************************************************************************

func AssignPageCoordinates(maplines []PageMap) map[NodePtr]Coords {

	// Make a quasi causal cone [width][depth] to span the geometry

	var directory = make(map[NodePtr]Coords)
	var already = make(map[NodePtr]bool)
	var axis []NodePtr
	var satellites = make(map[NodePtr][]NodePtr)
	var allnotes int

	// Order unique axial leads and satellite notes

	for depth := 0; depth < len(maplines); depth++ {

		axial_nptr := maplines[depth].Path[0].Dst

		if !already[axial_nptr] {
			allnotes++
			already[axial_nptr] = true
			axis = append(axis,axial_nptr)
		}

		axis := maplines[depth].Path[0].Dst

		for sat := 1; sat < len(maplines[depth].Path); sat++ {
			orbit := maplines[depth].Path[sat].Dst
			if !already[orbit] {
				satellites[axis] = append(satellites[axis],orbit)
				already[orbit] = true
			}
		}
	}

	const screen = 2.0
	const z_start = -1.0
	var zinc = screen / float64(allnotes)

	for tz := 0; tz < len(axis); tz++ {

		var leader Coords

		leader.X = 0
		leader.Y = 0
		leader.Z = z_start + float64(tz) * zinc // [-1,1]

		directory[axis[tz]] = leader

		// Arrange the notes orbitally around the leader

		satrange := float64(len(satellites[axis[tz]]))

		for i,sat := range(satellites[axis[tz]]) {

			pos := float64(i)
			radius := 0.5 + (0.2*leader.Z) // heuristic scaling to fit extrema
			var satc Coords
			nptr := sat
			satc.X = radius * math.Cos(2.0 * pos * math.Pi/satrange)
			satc.Y = radius * math.Sin(2.0 * pos * math.Pi/satrange)
			satc.Z = leader.Z

			directory[nptr] = satc
		}

	}

	return directory
}

// **************************************************************************

func AssignChapterCoordinates(nth,swimlanes int) Coords {

	// Place chapters uniformly over the surface of a sphere, using
	// the Fibonacci lattice

	N := float64(swimlanes)
	n := float64(nth)
	const fibratio = 1.618
	const rho = 0.75

	latitude := math.Asin(2 * n / (2 * N + 1))
	longitude := 2 * math.Pi * n/fibratio

	if longitude < -math.Pi {
		longitude += 2 * math.Pi
	}

	if longitude > math.Pi {
		longitude -= 2 * math.Pi
	}

	var fxyz Coords

	fxyz.X = float64(-rho * math.Sin(longitude))
	fxyz.Y = float64(rho * math.Sin(latitude))
	fxyz.Z = float64(rho * math.Cos(longitude) * math.Cos(latitude))

	fxyz.R = rho
	fxyz.Lat = latitude
	fxyz.Lon = longitude

	return fxyz
}

// **************************************************************************

func AssignContextSetCoordinates(origin Coords,nth,swimlanes int) Coords {

	N := float64(swimlanes)
	n := float64(nth)
	latitude := float64(origin.Lat)
	longitude := float64(origin.Lon)
	rho := 0.85

	orbital_angle := math.Pi / 8

	var fxyz Coords

	if N == 1 {
		fxyz.X = -rho * math.Sin(longitude)
		fxyz.Y = rho * math.Sin(latitude)
		fxyz.Z = rho * math.Cos(longitude) * math.Cos(latitude)
		return fxyz
	}

	delta_lon := orbital_angle * math.Sin(2 * math.Pi * n / N)
	delta_lat := orbital_angle * math.Cos(2 * math.Pi * n / N)

	fxyz.X = -rho * math.Sin(longitude+delta_lon)
	fxyz.Y = rho * math.Sin(latitude+delta_lat)
	fxyz.Z = rho * math.Cos(longitude+delta_lon) * math.Cos(latitude+delta_lat)

	return fxyz
}

// **************************************************************************

func AssignFragmentCoordinates(origin Coords,nth,swimlanes int) Coords {

	// These are much more crowded, so stagger radius

	N := float64(swimlanes)
	n := float64(nth)
	latitude := float64(origin.Lat)
	longitude := float64(origin.Lon)

	rho := 0.3 + float64(nth % 2) * 0.1

	orbital_angle := math.Pi / 12

	var fxyz Coords

	if N == 1 {
		fxyz.X = -rho * math.Sin(longitude)
		fxyz.Y = rho * math.Sin(latitude)
		fxyz.Z = rho * math.Cos(longitude) * math.Cos(latitude)
		return fxyz
	}

	delta_lon := orbital_angle * math.Sin(2 * math.Pi * n / N)
	delta_lat := orbital_angle * math.Cos(2 * math.Pi * n / N)

	fxyz.X = -rho * math.Sin(longitude+delta_lon)
	fxyz.Y = rho * math.Sin(latitude+delta_lat)
	fxyz.Z = rho * math.Cos(longitude+delta_lon) * math.Cos(latitude+delta_lat)

	return fxyz
}

// **************************************************************************

func MakeCoordinateDirectory(XChannels []float64, unique [][]NodePtr,maxzlen,nth,swimlanes int) map[NodePtr]Coords {

	var directory = make(map[NodePtr]Coords)

	const totwidth = 2.0 // This is the width dimenion of the paths -1 to +1
	const totdepth = 2.0 // This is the depth dimenion of the paths -1 to +1
	const arbitrary_elevation = 0.0

	x_lanewidth := totwidth / (float64(swimlanes))
	tz_steplength := totdepth / float64(maxzlen) 

	x_lane_start := float64(nth) * x_lanewidth - totwidth/2.0

	// Start allocating swimlane into XChannels parallel spaces
	// x now runs from (x_lane_start to += x_lanewidth)

	for tz := 0; tz < maxzlen && tz < len(unique); tz++ {

		x_increment := x_lanewidth / (XChannels[tz]+1)

		z_left := -float64(totwidth/2)
		x_left := float64(x_lane_start) + x_increment 

		var xyz Coords

		xyz.X = x_left
		xyz.Y = arbitrary_elevation
		xyz.Z = z_left + tz_steplength * float64(tz)

		// Each cross section, at depth tz

		for uniqptr := 0; uniqptr < len(unique[tz]); uniqptr++ {
			directory[unique[tz][uniqptr]] = xyz
			xyz.X += x_increment
		}
	}

	return directory
}


//
// viewport_coordinates.go
//


