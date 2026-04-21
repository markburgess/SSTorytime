//**************************************************************
//
// globals.go
//
//**************************************************************

package SSTorytime

const (
	CREDENTIALS_FILE = ".SSTorytime" // user's home directory

	ERR_ST_OUT_OF_BOUNDS = "Link STtype is out of bounds (must be -3 to +3)"
	ERR_ILLEGAL_LINK_CLASS = "ILLEGAL LINK CLASS"
	ERR_NO_SUCH_ARROW = "No such arrow has been declared in the configuration: "
	ERR_MEMORY_DB_ARROW_MISMATCH = "Arrows in database are not in synch (shouldn't happen)"
	ERR_MEMORY_DB_CONTEXT_MISMATCH = "Contexts in database are not in synch (shouldn't happen)"
	WARN_DIFFERENT_CAPITALS = "WARNING: Another capitalization exists"

	SCREENWIDTH = 120
	RIGHTMARGIN = 5
	LEFTMARGIN = 5

	NEAR = 0
	LEADSTO = 1   // +/-
	CONTAINS = 2  // +/-
	EXPRESS = 3   // +/-

	// Letting a cone search get too large is unresponsive
	CAUSAL_CONE_MAXLIMIT = 100

	// And shifted indices for array indicesin Go

	ST_ZERO = EXPRESS
	ST_TOP = ST_ZERO + EXPRESS + 1

	// For the SQL table, as 2d arrays not good

	I_MEXPR = "Im3"
	I_MCONT = "Im2"
	I_MLEAD = "Im1"
	I_NEAR  = "In0"
	I_PLEAD = "Il1"
	I_PCONT = "Ic2"
	I_PEXPR = "Ie3"

	// For separating text types

	N1GRAM = 1
	N2GRAM = 2
	N3GRAM = 3
	LT128 = 4
	LT1024 = 5
	GT1024 = 6

	// mandatory relations used in text processing, but we should never follow these links
        // when presenting results

	EXPR_INTENT_L = "has contextual theme"
	EXPR_INTENT_S = "has_theme"
        INV_EXPR_INTENT_L = "is a context theme in"
	INV_EXPR_INTENT_S ="theme_of"

	EXPR_AMBIENT_L = "has contextual highlight"
	EXPR_AMBIENT_S = "has_highlight"
        INV_EXPR_AMBIENT_L = "is contextual highlight of"
	INV_EXPR_AMBIENT_S = "highlight_of"

	CONT_FINDS_L = "contains extract/quote"
	CONT_FINDS_S = "has-extract"
	INV_CONT_FOUND_IN_L = "extract/quote from"
	INV_CONT_FOUND_IN_S = "extract-fr"

	// This is a "contained-by something that expresses" shortcut => NEAR

	CONT_FRAG_L = "contains intented characteristic"      // intentional characteristic
	CONT_FRAG_S = "has-frag"
	INV_CONT_FRAG_IN_L = "characteristic of"   // explains intentional context
	INV_CONT_FRAG_IN_S = "charct-in"

	NON_ASCII_LQUOTE = '“'
	NON_ASCII_RQUOTE = '”'

        FORGOTTEN = 10800
        TEXT_SIZE_LIMIT = 30

)


// **************************************************************************

var ( 

	NO_NODE_PTR NodePtr // see Init()
	NONODE NodePtr

	// Uploading

	WIPE_DB bool = false
        SILLINESS_COUNTER int
        SILLINESS_POS int
        SILLINESS_SLOGAN int
	SILLINESS bool


        // Text analysis

        STM_INT_FRAG = make(map[string]History) // for intentional (exceptional) fragments
        STM_AMB_FRAG = make(map[string]History) // for ambient (repeated) fragments
        STM_INV_GROUP = make(map[string]History) // look for invariants

)


//
// end globals.go
//


