// **************************************************************
//
// postgres_types_functions.go
//
// **************************************************************

package sst

import (
	"fmt"
	"strings"
)

// **************************************************************

const NODEPTR_TYPE = "CREATE TYPE NodePtr AS  " +
	"(                    " +
	"Chan     int,        " +
	"CPtr     int         " +
	")"

const LINK_TYPE = "CREATE TYPE Link AS  " +
	"(                    " +
	"Arr      int,        " +
	"Wgt      real,       " +
	"Ctx      int,        " +
	"Dst      NodePtr     " +
	")"

const NODE_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS Node " +
	"( " +
	"NPtr      NodePtr,        \n" +
	"L         int,            \n" +
	"S         text,           \n" +
	"Search    TSVECTOR GENERATED ALWAYS AS (to_tsvector('english',S)) STORED,\n" +
	"UnSearch  TSVECTOR GENERATED ALWAYS AS (to_tsvector('english',sst_unaccent(S))) STORED,\n" +
	"Chap      text,           \n" +
	"Seq       boolean,        \n" +
	I_MEXPR + "  Link[],         \n" + // Im3
	I_MCONT + "  Link[],         \n" + // Im2
	I_MLEAD + "  Link[],         \n" + // Im1
	I_NEAR + "  Link[],         \n" + // In0
	I_PLEAD + "  Link[],         \n" + // Il1
	I_PCONT + "  Link[],         \n" + // Ic2
	I_PEXPR + "  Link[]          \n" + // Ie3
	")"

const PAGEMAP_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS PageMap " +
	"( " +
	"Chap     Text,  " +
	"Alias    Text,  " +
	"Ctx      int,   " +
	"Line     Int,   " +
	"Path     Link[] " +
	")"

const ARROW_DIRECTORY_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS ArrowDirectory " +
	"(    " +
	"STAindex int,           " +
	"Long text,              " +
	"Short text,             " +
	"ArrPtr int primary key  " +
	")"

const ARROW_INVERSES_TABLE = "CREATE UNLOGGED TABLE IF NOT EXISTS ArrowInverses " +
	"(    " +
	"Plus int,  " +
	"Minus int,  " +
	"Primary Key(Plus,Minus)" +
	")"

const LASTSEEN_TABLE = "CREATE TABLE IF NOT EXISTS LastSeen " +
	"(    " +
	"Section text," +
	"NPtr    NodePtr," +
	"First   timestamp," +
	"Last    timestamp," +
	"Delta   real," +
	"Freq    int" +
	")"

const CONTEXT_DIRECTORY_TABLE = "CREATE TABLE IF NOT EXISTS ContextDirectory " +
	"(    " +
	"Context text,            " +
	"CtxPtr  int primary key  " +
	")"

const BOOKMARK_TABLE = "CREATE TABLE IF NOT EXISTS Bookmarks " +
	"(    " +
	"Query text,  " +
	"Bookmark text" +
	")"

const APPOINTMENT_TYPE = "CREATE TYPE Appointment AS  " +
	"(                    " +
	"Arr    int," +
	"STType int," +
	"Chap   text," +
	"Ctx    int," +
	"NTo    NodePtr," +
	"NFrom  NodePtr[]" +
	")"

// **************************************************************************

func CreateType(sst PoSST, defn string) bool {
	_, err := sst.exec(defn)
	if err != nil {
		s := fmt.Sprintln("Failed to create datatype", err)
		return strings.Contains(s, "already exists")
	}
	return true
}

// **************************************************************************

func CreateTable(sst PoSST, defn string) bool {

	row, err := sst.query(defn)

	if err != nil {
		s := fmt.Sprintln("Failed to create a table %.10 ...", defn, err)

		if strings.Contains(s, "already exists") {
			return true
		} else {
			return false
		}
	}

	row.Close()
	return true
}

// **************************************************************************

func DefineStoredFunctions(sst PoSST) {
	// No-op: PL/pgSQL installed by migration 000002_functions.up.sql
	_ = sst
}
