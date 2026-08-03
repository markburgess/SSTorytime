package sst

import "github.com/markburgess/SSTorytime/internal/db/sqlc"

// StTypeFromInt maps Go ST type (-3..+3) to the Postgres st_type enum.
func StTypeFromInt(sttype int) sqlc.StType {
	switch sttype {
	case -EXPRESS:
		return sqlc.StTypeMExpress
	case -CONTAINS:
		return sqlc.StTypeMContains
	case -LEADSTO:
		return sqlc.StTypeMLeads
	case NEAR:
		return sqlc.StTypeNear
	case LEADSTO:
		return sqlc.StTypePLeads
	case CONTAINS:
		return sqlc.StTypePContains
	case EXPRESS:
		return sqlc.StTypePExpress
	default:
		return sqlc.StTypeNear
	}
}

// StTypeToInt is the inverse of StTypeFromInt.
func StTypeToInt(t sqlc.StType) int {
	switch t {
	case sqlc.StTypeMExpress:
		return -EXPRESS
	case sqlc.StTypeMContains:
		return -CONTAINS
	case sqlc.StTypeMLeads:
		return -LEADSTO
	case sqlc.StTypeNear:
		return NEAR
	case sqlc.StTypePLeads:
		return LEADSTO
	case sqlc.StTypePContains:
		return CONTAINS
	case sqlc.StTypePExpress:
		return EXPRESS
	default:
		return NEAR
	}
}
