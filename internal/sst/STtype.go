// **************************************************************************
//
// STtype.go
//
// *************************************************************************

package sst

// *************************************************************************

func GetSTIndexByName(stname, pm string) int {

	var encoding int
	var sign int

	switch pm {
	case "+":
		sign = 1
	case "-":
		sign = -1
	}

	switch stname {

	case "leadsto":
		encoding = ST_ZERO + LEADSTO*sign
	case "contains":
		encoding = ST_ZERO + CONTAINS*sign
	case "properties":
		encoding = ST_ZERO + EXPRESS*sign
	case "similarity":
		encoding = ST_ZERO + NEAR
	}

	return encoding

}

//**************************************************************

func PrintSTAIndex(stindex int) string {

	sttype := stindex - ST_ZERO
	var ty string

	switch sttype {
	case -EXPRESS:
		ty = "-(expressed by)"
	case -CONTAINS:
		ty = "-(part of)"
	case -LEADSTO:
		ty = "-(arriving from)"
	case NEAR:
		ty = "(close to)"
	case LEADSTO:
		ty = "+(leading to)"
	case CONTAINS:
		ty = "+(containing)"
	case EXPRESS:
		ty = "+(expressing)"
	default:
		ty = "unknown relation!"
	}

	const green = "\x1b[36m"
	const endgreen = "\x1b[0m"

	return green + ty + endgreen
}

// **************************************************************************

func STIndexToSTType(stindex int) int {

	// Convert shifted array index to symmetrical type

	return stindex - ST_ZERO
}

// **************************************************************************

func STTypeToSTIndex(stindex int) int {

	// Convert shifted array index to symmetrical type

	return stindex + ST_ZERO
}

// **************************************************************************

func STTypeName(sttype int) string {

	switch sttype {
	case -EXPRESS:
		return "-is property of"
	case -CONTAINS:
		return "-contained by"
	case -LEADSTO:
		return "-comes from"
	case NEAR:
		return "=Similarity"
	case LEADSTO:
		return "+leads to"
	case CONTAINS:
		return "+contains"
	case EXPRESS:
		return "+property"
	}

	return "Unknown ST type"
}

//
// STtype.go
//
