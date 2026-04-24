
package SSTorytime

import (
	"database/sql"
	"fmt"
	"os"
	"io/ioutil"
	_ "github.com/lib/pq"

)


//**************************************************************
//
// session.go
//
//**************************************************************

func Open(load_arrows bool) PoSST {

	var sst PoSST
	var err error
	
	// Replace credentials with a private file

	var (
		user     = "sstoryline"
		password = "sst_1234"
		dbname   = "sstoryline"
	)

	user,password,dbname = OverrideCredentials(user,password,dbname)

        connect_str := "user="+user+" dbname="+dbname+" password="+password+" sslmode=disable"

	// Check environment variable
        // postgresql://<USER>:<PASSWORD>@<HOST>:<PORT>/<DATABASE>
	// export POSTGRESQL_URI=postgresql://sstoryline:sst1234@localhost:5432/sstoryline?sslmode=disable

	env := os.Getenv("POSTGRESQL_URI")

	if len(env) == 0 {
		sst.DB, err = sql.Open("postgres", connect_str)
	} else {
		sst.DB, err = sql.Open("postgres",env)
	}

	if err != nil {
		fmt.Println("Error connecting to the database: ", err)
		os.Exit(-1)
	}

	// Basic test
	
	err = sst.DB.Ping()
	
	if err != nil {
		fmt.Println("Error pinging the database: ", err)
		os.Exit(-1)
	}

	MemoryInit(&sst)
	Configure(sst,load_arrows)

	DownloadArrowsFromDB(&sst)
	DownloadContextsFromDB(&sst)
	SynchronizeNPtrs(&sst)
	
	NO_NODE_PTR.Class = 0
	NO_NODE_PTR.CPtr =  -1
	NONODE.Class = 0
	NONODE.CPtr = 0

	return sst
}

// **************************************************************************

func OverrideCredentials(u,p,d string) (string,string,string) {

	// Store database/postgres credentials in a system file instead of hardcoding

	dirname, err := os.UserHomeDir()

	if err != nil && len(dirname) > 1 {
		fmt.Println("Unable to determine user's home directory")
		os.Exit(-1)
	}

	filename := dirname+"/"+CREDENTIALS_FILE
	content,err := ioutil.ReadFile(filename)

	if err != nil {
		return u,p,d
	}

	/* format
          dbname: sstoryline 
          user:sstoryline 
          passwd: sst_1234
        */

	var (
		offset,delta int
		user=u
		password=p
		dbname=d
	)

	for offset = 0; offset < len(content); offset = offset {

		var conf string
		fmt.Sscanf(string(content[offset:]),"%s",&conf)

		if len(conf) > 0 && conf[len(conf)-1] != ':' { // missing space

			for delta = 0; delta < len(conf); delta++ {
				if conf[delta] == ':' {
					conf = conf[:delta+1]
				}
			}
		}

		switch(conf) {
		case "user:":
			delta = len(conf)
			user,offset = GetLine(content,offset+delta)
		case "passwd:","password:":
			delta = len(conf)
			password,offset = GetLine(content,offset+delta)
		case "db:","dbname:":
			delta = len(conf)
			dbname,offset = GetLine(content,offset+delta)
		default:
			offset++
		}
	}

	return user,password,dbname
}

// **************************************************************************

func GetLine(s []byte,i int) (string,int) {

	// For parsing the password credential file

	var result []byte

	for o := i; o < len(s); o++ {

		if s[o] == '\n' {
			i = o
			break
		}

		result = append(result,s[o])
	}

	return string(result),i
}

// **************************************************************************

func MemoryInit(sst *PoSST) {

//  When opening a connection, restore config and allocate maps

        if sst.NODE_DIRECTORY.N1grams == nil {
		sst.NODE_DIRECTORY.N1grams = make(map[string]ClassedNodePtr)
	}

	if sst.NODE_DIRECTORY.N2grams == nil {
		sst.NODE_DIRECTORY.N2grams = make(map[string]ClassedNodePtr)
	}

	if sst.NODE_DIRECTORY.N3grams == nil {
		sst.NODE_DIRECTORY.N3grams = make(map[string]ClassedNodePtr)
	}

	if sst.NODE_DIRECTORY.LT128 == nil {
		sst.NODE_DIRECTORY.LT128 = make(map[string]ClassedNodePtr)
	}

	sst.NODE_CACHE = make(map[NodePtr]NodePtr)
	sst.INVERSE_ARROWS = make(map[ArrowPtr]ArrowPtr)
	sst.ARROW_SHORT_DIR = make(map[string]ArrowPtr)
	sst.ARROW_LONG_DIR = make(map[string]ArrowPtr)
	sst.ARROW_DIRECTORY_TOP = 0
	sst.CONTEXT_DIR = make(map[string]ContextPtr)
}

// **************************************************************************

func Configure(sst PoSST,load_arrows bool) {

	// Tmp reset

	if WIPE_DB {

		fmt.Println("***********************")
		fmt.Println("* WIPING DB")
		fmt.Println("***********************")
		
		sst.DB.QueryRow("DROP INDEX sst_nan")
		sst.DB.QueryRow("DROP INDEX sst_type")
		sst.DB.QueryRow("DROP INDEX sst_gin")
		sst.DB.QueryRow("DROP INDEX sst_ungin")
		sst.DB.QueryRow("DROP INDEX sst_s")
		sst.DB.QueryRow("DROP INDEX sst_n")
		sst.DB.QueryRow("DROP INDEX sst_cnt")

		sst.DB.QueryRow("drop function fwdconeaslinks")
		sst.DB.QueryRow("drop function fwdconeasnodes")
		sst.DB.QueryRow("drop function fwdpathsaslinks")
		sst.DB.QueryRow("drop function getfwdlinks")
		sst.DB.QueryRow("drop function getfwdnodes")
		sst.DB.QueryRow("drop function getneighboursbytype")
		sst.DB.QueryRow("drop function getsingletonaslink")
		sst.DB.QueryRow("drop function AllNCPathsAsLinks")
		sst.DB.QueryRow("drop function AllSuperNCPathsAsLinks")
		sst.DB.QueryRow("drop function SumAllNCPaths")
		sst.DB.QueryRow("drop function GetNCFwdLinks")
		sst.DB.QueryRow("drop function GetNCCLinks")

		sst.DB.QueryRow("drop function getsingletonaslinkarray")
		sst.DB.QueryRow("drop function idempinsertnode")
		sst.DB.QueryRow("drop function sumfwdpaths")
		sst.DB.QueryRow("drop function match_context")
		sst.DB.QueryRow("drop function empty_path")
		sst.DB.QueryRow("drop function match_arrows")
		sst.DB.QueryRow("drop function ArrowInList")
		sst.DB.QueryRow("drop function GetNCCStoryStartNodes")
		sst.DB.QueryRow("drop function GetStoryStartNodes")
		sst.DB.QueryRow("drop function GetAppointments")
		sst.DB.QueryRow("drop function UnCmp")
		sst.DB.QueryRow("drop function DeleteChapter")

		sst.DB.QueryRow("drop function lastsawsection(text)")
		sst.DB.QueryRow("drop function lastsawnptr(nodeptr)")

		sst.DB.QueryRow("drop type NodePtr")
		sst.DB.QueryRow("drop type Link")
		sst.DB.QueryRow("drop type Appointment")

		sst.DB.QueryRow("drop table Node")
		sst.DB.QueryRow("drop table PageMap")
		sst.DB.QueryRow("drop table NodeArrowNode")
		sst.DB.QueryRow("drop table ArrowDirectory")
		sst.DB.QueryRow("drop table ArrowInverses")
		sst.DB.QueryRow("drop table ContextDirectory")
		sst.DB.QueryRow("drop table LastSeen")

	}

	// Create functions, some we use in autocreating index columns

	sst.DB.QueryRow("CREATE EXTENSION unaccent")

	if !CreateType(sst,NODEPTR_TYPE) {
		fmt.Println("Unable to create type as, ",NODEPTR_TYPE)
		os.Exit(-1)
	}

	if !CreateType(sst,LINK_TYPE) {
		fmt.Println("Unable to create type as, ",LINK_TYPE)
		os.Exit(-1)
	}

	if !CreateType(sst,APPOINTMENT_TYPE) {
		fmt.Println("Unable to create type as, ",APPOINTMENT_TYPE)
		os.Exit(-1)
	}

	if !CreateTable(sst,CONTEXT_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",CONTEXT_DIRECTORY_TABLE)
		os.Exit(-1)
	}

	DefineStoredFunctions(sst)

	if !CreateTable(sst,PAGEMAP_TABLE) {
		fmt.Println("Unable to create table as, ",PAGEMAP_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,NODE_TABLE) {
		fmt.Println("Unable to create table as, ",NODE_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,ARROW_INVERSES_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_INVERSES_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,ARROW_DIRECTORY_TABLE) {
		fmt.Println("Unable to create table as, ",ARROW_DIRECTORY_TABLE)
		os.Exit(-1)
	}

	if !CreateTable(sst,LASTSEEN_TABLE) {
		fmt.Println("Unable to create table as, ",LASTSEEN_TABLE)
		os.Exit(-1)
	}

	// Find ignorable arrows
}


// **************************************************************************

func Close(sst PoSST) {
	sst.DB.Close()
}



//
// session.go
//


