// **************************************************************
//
// postgres_types_functions.go
//
// **************************************************************

package SSTorytime

import (
	"fmt"
	"strings"
	_ "github.com/lib/pq"

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
	I_MEXPR+"  Link[],         \n" + // Im3
	I_MCONT+"  Link[],         \n" + // Im2
	I_MLEAD+"  Link[],         \n" + // Im1
	I_NEAR +"  Link[],         \n" + // In0
	I_PLEAD+"  Link[],         \n" + // Il1
	I_PCONT+"  Link[],         \n" + // Ic2
	I_PEXPR+"  Link[]          \n" + // Ie3
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
	"First   timestamp,"+
	"Last    timestamp," +
	"Delta   real," +
	"Freq    int" +
	")"

const CONTEXT_DIRECTORY_TABLE = "CREATE TABLE IF NOT EXISTS ContextDirectory " +
	"(    " +
	"Context text,            " +
	"CtxPtr  int primary key  " +
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

	row,err := sst.DB.Query(defn)

	if err != nil {
		s := fmt.Sprintln("Failed to create datatype PGLink ",err)
		
		if strings.Contains(s,"already exists") {
			return true
		} else {
			return false
		}
	}

	row.Close();
	return true
}

// **************************************************************************

func CreateTable(sst PoSST,defn string) bool {

	row,err := sst.DB.Query(defn)
	
	if err != nil {
		s := fmt.Sprintln("Failed to create a table %.10 ...",defn,err)
		
		if strings.Contains(s,"already exists") {
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

	// NB! these functions are in "plpgsql" language, NOT SQL. They look similar but they are DIFFERENT!
	
	// This is not a pretty function, but in order to interface go-types to pg-types, we need to evaluate it 
	// like this...
	
	
	cols := I_MEXPR+","+I_MCONT+","+I_MLEAD+","+I_NEAR +","+I_PLEAD+","+I_PCONT+","+I_PEXPR


	qstr := fmt.Sprintf("CREATE OR REPLACE FUNCTION IdempInsertNode(iLi INT, iszchani INT, icptri INT, iSi TEXT, ichapi TEXT)\n" +
		"RETURNS TABLE (    \n" +
		"    ret_cptr INTEGER," +
		"    ret_channel INTEGER" +
		") AS $fn$ " +
		"DECLARE \n" +
		"BEGIN\n" +
		"  IF NOT EXISTS (SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE lower(s) = lower(iSi)) THEN\n" +
		"     INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap,%s) VALUES (iszchani,icptri,iLi,iSi,ichapi,'{}','{}','{}','{}','{}','{}','{}');" +
		"  END IF;\n" +
		"  RETURN QUERY SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;",cols);

	row,err := sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Force for managed input

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION InsertNode(iLi INT, iszchani INT, icptri INT, iSi TEXT, ichapi TEXT,sequence boolean)\n" +
		"RETURNS bool AS $fn$ " +
		"DECLARE \n" +
		"BEGIN\n" +
		"   INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap,Seq,%s) VALUES (iszchani,icptri,iLi,iSi,ichapi,sequence,'{}','{}','{}','{}','{}','{}','{}');" +
		"   RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;",cols);

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Without controlling nptr

	qstr = "CREATE OR REPLACE FUNCTION IdempAppendNode(iLi INT, iszchani INT, iSi TEXT, ichapi TEXT)\n" +
		"RETURNS TABLE (    \n" +
		"    ret_cptr INTEGER," +
		"    ret_channel INTEGER" +
		") AS $fn$ " +
		"DECLARE \n" +
		"    icptri INT = 0;" +
		"BEGIN\n" +
		"  IF NOT EXISTS (SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi) THEN\n" +
		"     SELECT max((Nptr).CPtr) INTO icptri FROM Node WHERE (Nptr).Chan=iszchani;\n"+
		"     IF icptri IS NULL THEN"+
		"         icptri = 0;"+
		"     END IF;"+
		"     INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap) VALUES (iszchani,icptri+1,iLi,iSi,ichapi);" +
		"  END IF;\n" +
		"  RETURN QUERY SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;";

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Insert Context from API

	qstr = "CREATE OR REPLACE FUNCTION IdempInsertContext(constr text,conptr int)\n" +
		"RETURNS int AS $fn$ " +
		"DECLARE \n" +
		"    cptr INT = 0;\n" +
		"    found int=-99;\n" +
		"BEGIN\n" +
		"IF conptr=-1 THEN\n"+
		"   SELECT Context,CtxPtr INTO found FROM ContextDirectory WHERE Context=constr AND CtxPtr=conptr;\n"+
		"   SELECT max(CtxPtr) INTO cptr FROM ContextDirectory;\n"+
		"   INSERT INTO ContextDirectory (Context,CtxPtr) VALUES (constr,cptr+1);\n"+
		"   RETURN cptr+1;\n" +
		"END IF;\n" +
		"IF NOT EXISTS (SELECT CtxPtr FROM ContextDirectory WHERE CtxPtr=conptr OR Context=constr) THEN\n" +
		"   INSERT INTO ContextDirectory (Context,CtxPtr) VALUES (constr,conptr);\n"+
		"   RETURN conptr;\n" +
		"END IF;"+
		"SELECT CtxPtr INTO cptr FROM ContextDirectory WHERE CtxPtr=conptr OR Context=constr;\n"+
		"RETURN cptr;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;";

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// For lookup include name,chapter,context,arrow
	
	qstr = "CREATE OR REPLACE FUNCTION NCC_match(thisnptr NodePtr,context text[],arrows int[],sttypes int[],lm3 Link[],lm2 Link[],lm1 Link[],ln0 Link[],lp1 Link[],lp2 Link[],lp3 Link[])\n"+
		"RETURNS boolean AS $fn$\n"+
		"DECLARE \n"+
		"    emptyarray Link[] := Array[] :: Link[];\n"+
		"    lnkarray Link[] := Array[] :: Link[];\n"+
		"    lnk Link;\n"+
		"    st int;\n"+
		"BEGIN\n"+
		
		// If there are no arrows, we only need to look for the Node context in lp1 for "empty" == 0

		"IF array_length(arrows,1) IS NULL THEN\n"+

		"   IF lp1 IS NOT NULL THEN"+		
		"      FOREACH lnk IN ARRAY lp1 LOOP\n"+
                "         IF lnk.Arr = 0 AND match_context(lnk.Ctx,context) THEN\n"+
		"            RETURN true;"+
		"         END IF;"+
		"      END LOOP;\n"+
		"   END IF;\n"+

		"ELSE\n"+

		// If there are arrows

		"   FOREACH st IN ARRAY sttypes LOOP\n"+
		"      CASE st \n"		
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("   WHEN %d THEN\n"+
			"         SELECT %s INTO lnkarray FROM Node WHERE Nptr=thisnptr;\n",st,STTypeDBChannel(st));
	}
	qstr +=	"      ELSE RAISE EXCEPTION 'No such sttype in NCC_match %', sttype;\n" +
		"      END CASE;\n" +
		
		"      FOREACH lnk IN ARRAY lnkarray LOOP\n"+
		"         IF match_arrows(lnk.arr,arrows) AND match_context(lnk.ctx,context) THEN\n"+
		"            RETURN true;\n"+

		"         END IF;\n"+
		"      END LOOP;\n"+
		"   END LOOP;\n"+
		"END IF;\n"+

		"RETURN false; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct an empty link pointing nowhere as a starting node

	qstr = "CREATE OR REPLACE FUNCTION GetSingletonAsLinkArray(start NodePtr)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    level Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		" lnk.Dst = start;\n"+
		" level = array_append(level,lnk);\n"+
		"RETURN level; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct an empty link pointing nowhere as a starting node

	qstr = "CREATE OR REPLACE FUNCTION GetSingletonAsLink(start NodePtr)\n"+
		"RETURNS Link AS $fn$\n"+
		"DECLARE \n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		" lnk.Dst = start;\n"+
		"RETURN lnk; \n"+
		"END ;\n"+
		"$fn$ LANGUAGE plpgsql;"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Construct search by sttype. since table names are static we need a case statement

	qstr = "CREATE OR REPLACE FUNCTION GetNeighboursByType(start NodePtr, sttype int, maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    fwdlinks Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT %s INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;\n",st,STTypeDBChannel(st));
	}
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN fwdlinks; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Get the nearest neighbours as NPtr, with respect to each of the four STtype

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetFwdNodes(start NodePtr,exclude NodePtr[],sttype int,maxlimit int)\n"+
		"RETURNS NodePtr[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours NodePtr[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNeighboursByType(start,sttype,maxlimit);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +

		"    neighbours := ARRAY[]::NodePtr[];\n" +

		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+
		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk.dst);\n" +
		"      END IF; \n" +
		"    END LOOP;\n" +

		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Basic quick neighbour probe

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetFwdLinks(start NodePtr,exclude NodePtr[],sttype int,maxlimit int)\n")
	qstr +=	"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNeighboursByType(start,sttype,maxlimit);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +
		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+
		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 
		"    END LOOP;\n" +
		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()
	
	// Get the forward cone / half-ball as NPtr

	qstr = "CREATE OR REPLACE FUNCTION FwdConeAsNodes(start NodePtr,sttype INT, maxdepth INT,maxlimit int)\n"+
		"RETURNS NodePtr[] AS $fn$\n" +
		"DECLARE \n" +
		"    nextlevel NodePtr[];\n" +
		"    partlevel NodePtr[];\n" +
		"    level NodePtr[] = ARRAY[start]::NodePtr[];\n" +
		"    exclude NodePtr[] = ARRAY['(0,0)']::NodePtr[];\n" +
		"    cone NodePtr[];\n" +
		"    neigh NodePtr;\n" +
		"    frn NodePtr;\n" +
		"    counter int := 0;\n" +

		"BEGIN\n" +

		"LOOP\n" +
		"  EXIT WHEN counter = maxdepth+1;\n" +

		"  IF level IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  nextlevel := ARRAY[]::NodePtr[];\n" +

		"  FOREACH frn IN ARRAY level "+
		"  LOOP \n"+
		"     nextlevel = array_append(nextlevel,frn);\n" +
		"  END LOOP;\n" +

		"  IF nextlevel IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  FOREACH neigh IN ARRAY nextlevel LOOP \n"+
		"    IF NOT neigh = ANY(exclude) THEN\n" +
		"      cone = array_append(cone,neigh);\n" +
		"      exclude := array_append(exclude,neigh);\n" +
		"      partlevel := GetFwdNodes(neigh,exclude,sttype,maxlimit);\n" +
		"    END IF;" +
		"    IF partlevel IS NOT NULL THEN\n" +
		"         level = array_cat(level,partlevel);\n"+
		"    END IF;\n" +
		"  END LOOP;\n" +

		// Next, continue, foreach
		"  counter = counter + 1;\n" +
		"END LOOP;\n" +
		
		"RETURN cone; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	qstr = "CREATE OR REPLACE FUNCTION FwdConeAsLinks(start NodePtr,sttype INT,maxdepth INT,maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    nextlevel Link[];\n" +
		"    partlevel Link[];\n" +
		"    level Link[] = ARRAY[]::Link[];\n" +
		"    exclude NodePtr[] = ARRAY['(0,0)']::NodePtr[];\n" +
		"    cone Link[];\n" +
		"    neigh Link;\n" +
		"    frn Link;\n" +
		"    counter int := 0;\n" +

		"BEGIN\n" +

		"level := GetSingletonAsLinkArray(start);\n"+

		"LOOP\n" +
		"  EXIT WHEN counter = maxdepth+1;\n" +

		"  IF level IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  nextlevel := ARRAY[]::Link[];\n" +

		"  FOREACH frn IN ARRAY level "+
		"  LOOP \n"+
		"     nextlevel = array_append(nextlevel,frn);\n" +
		"  END LOOP;\n" +

		"  IF nextlevel IS NULL THEN\n" +
		"     RETURN cone;\n" +
		"  END IF;\n" +

		"  FOREACH neigh IN ARRAY nextlevel LOOP \n"+
		"    IF NOT neigh.Dst = ANY(exclude) THEN\n" +
		"      cone = array_append(cone,neigh);\n" +
		"      exclude := array_append(exclude,neigh.Dst);\n" +
		"      partlevel := GetFwdLinks(neigh.Dst,exclude,sttype,maxlimit);\n" +
		"    END IF;" +
		"    IF partlevel IS NOT NULL THEN\n" +
		"         level = array_cat(level,partlevel);\n"+
		"    END IF;\n" +
		"  END LOOP;\n" +

		"  counter = counter + 1;\n" +
		"END LOOP;\n" +

		"RETURN cone; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Orthogonal (depth first) paths from origin spreading out

	qstr = "CREATE OR REPLACE FUNCTION FwdPathsAsLinks(start NodePtr,sttype INT,maxdepth INT, maxlimit INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   hop Text;\n" +
		"   path Text;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = ARRAY[start]::NodePtr[];\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;"+

		"BEGIN\n" +

		"startlnk := GetSingletonAsLink(start);\n"+
		"path := Format('%s',startlnk::Text);\n"+
		"ret_paths := SumFwdPaths(startlnk,path,sttype,1,maxdepth,exclude, maxlimit);" +

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Return end of path branches as aggregated text summaries

	qstr = "CREATE OR REPLACE FUNCTION SumFwdPaths(start Link,path TEXT, sttype INT,depth int, maxdepth INT,exclude NodePtr[], maxlimit INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"    count    int = 0;\n"+
		"    horizon  int = 0;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		"fwdlinks := GetFwdLinks(start.Dst,exclude,sttype, maxlimit);\n" +
		
		// limit recursion explosions
		"horizon := maxlimit - array_length(fwdlinks,1);"+

		"IF horizon < 0 THEN\n"+
		"  horizon = 0;\n"+
		"  maxdepth = depth + 1;"+
		"END IF;\n"+

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL OR count >= maxlimit THEN" +
		          // set end of path as return val
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"         RETURN ret_paths;"+
		"      ELSE\n"+
		"         count = count + 1;"+
		          // Add to the path and descend into new link
		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumFwdPaths(lnk,tot_path,sttype,depth+1,maxdepth,exclude,horizon);\n" +
		          // when we return, we reached the end of one path
		"         IF appendix IS NOT NULL THEN\n"+
	                     // append full path to list of all paths, separated by newlines
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"            count = count + regexp_count(appendix,';');"+
		"         ELSE"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;"+
		"      END IF;"+
		"   END IF;"+
		"END LOOP;"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Typeless cone searches

	qstr = "CREATE OR REPLACE FUNCTION AllPathsAsLinks(start NodePtr,orientation text,maxdepth INT, maxlimit INT)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   hop Text;\n" +
		"   path Text;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = ARRAY[start]::NodePtr[];\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;"+

		"BEGIN\n" +

		"startlnk := GetSingletonAsLink(start);\n"+
		"path := Format('%s',startlnk::Text);\n"+
		"ret_paths := SumAllPaths(startlnk,path,orientation,1,maxdepth,exclude, maxlimit);" +
		
		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
        // select AllPathsAsLinks('(4,1)',3)

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// SumAllPaths

	qstr = "CREATE OR REPLACE FUNCTION SumAllPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,exclude NodePtr[],maxlimit int)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    stlinks  Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"    counter int = 0;"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		// Get *All* in/out Links
		"CASE \n" +
		"   WHEN orientation = 'bwd' THEN\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   WHEN orientation = 'fwd' THEN\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   ELSE\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetFwdLinks(start.Dst,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"END CASE;\n" +

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF counter > maxlimit THEN\n"+
		"      RETURN ret_paths;"+
		"   END IF;"+
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"         RETURN ret_paths;"+
		"      ELSE\n"+
		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumAllPaths(lnk,tot_path,orientation,depth+1,maxdepth,exclude,maxlimit);\n" +
		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"         ELSE\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;\n"+
		"         counter = counter + 1;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Check if linkpath representation is just one item

	qstr = "CREATE OR REPLACE FUNCTION empty_path(path text)\n"+
		"RETURNS boolean AS $fn$\n" +
		"BEGIN \n" +
		"   IF strpos(path,';') THEN \n" + // exact match
		"      RETURN true;\n" +
		"   END IF;\n" +
		"RETURN false;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Matching context strings with fuzzy criteria. The policy/notes expression is db_set
	// the client/lookup set is user_set - both COULD use AND expressions.
	// We are looking for sets that overlap for a true result

	qstr = "CREATE OR REPLACE FUNCTION match_context(thisctxptr int,user_set text[])\n"+
		"RETURNS boolean AS $fn$\n" +
		"DECLARE\n" +
		"   ctxstr text;\n" +
		"   db_set text[];\n" +
		"   notes text[];\n" +
		"   client text[];\n" +
		"   pattern text;\n" +
		"   and_list text[];\n"+
		"   and_count int;\n"+
		"   and_result int = 0;\n"+
		"   end_result int = 0;\n"+
		"   or_list text[] = ARRAY[]::text[];\n"+
		"   item text;\n" +
		"   item_db text;\n" +
		"   item_us text;\n" +
		"   ref text;\n" +
		"   c text;\n"+
		"BEGIN \n" +

		// If no constraints at all, then match

		"IF array_length(user_set,1) IS NULL THEN\n" +  // Shouldn't happen anymore
		"   RETURN true;\n"+
		"END IF;\n"+

		"IF user_set[0] = '' THEN\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		// Convert context ptr into a list from the new factored cache
		"SELECT Context INTO ctxstr FROM ContextDirectory WHERE ctxPtr=thisctxptr;" +
		"db_set = regexp_split_to_array(ctxstr,',');\n" +

		// If there is a constraint, but no db membership, then no match
		"IF array_length(db_set,1) IS NULL AND array_length(user_set,1) IS NOT NULL THEN\n"+
		"   RETURN false;\n"+
		"END IF;\n"+

		// if both are empty, then match
		"IF array_length(db_set,1) IS NULL AND array_length(user_set,1) IS NULL THEN\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		// clean and unaccent sets

		"FOREACH item_db IN ARRAY db_set LOOP\n" +
		"   FOREACH item_us IN ARRAY user_set LOOP\n" +
		"      IF item_db = item_us THEN\n" +
		"         RETURN true;\n"+
		"      END IF;\n"+
		"   END LOOP;\n" +
		"END LOOP;\n" +

		"FOREACH item IN ARRAY db_set LOOP\n" +
		"   notes = array_append(notes,lower(unaccent(item)));\n" +
		"END LOOP;\n" +

		"FOREACH item IN ARRAY user_set LOOP\n" +
		"   client = array_append(client,lower(unaccent(item)));\n" +
		"END LOOP;\n" +

	       // First split check AND strings in the notes

		"FOREACH item IN ARRAY notes LOOP\n" +

		"   and_list = regexp_split_to_array(item, '\\.');\n" +
		"   and_count = array_length(and_list,1);\n"+

		"   IF and_count > 1 THEN\n"+

		// end_result = MatchANDExpression(and_list,client)

	        "      and_result = 0;\n"+

                       // check each and expression first

		"      FOREACH ref IN ARRAY and_list LOOP\n"+
		"         FOREACH c IN ARRAY client LOOP\n"+
		             // AND need an exact match
		"            IF ref = c THEN \n" +
	        "               and_result = and_result + 1;\n" +
		"            END IF;\n" +
		"         END LOOP;\n"+
		"      END LOOP;\n"+

		"      IF and_result = and_count THEN\n"+
		"         end_result = end_result + 1;\n"+
	        "      END IF;\n" +
		"   ELSE\n"+
		"      or_list = array_append(or_list,item);\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"IF end_result > 0 THEN\n"+
		"   RETURN true;\n" +
		"END IF;\n"+

		// if still not match, check any left overs, client AND matches are still unresolved
		"FOREACH ref IN ARRAY or_list LOOP\n" +
		    // now we can look at substring partial matches
		"   FOREACH c IN ARRAY client LOOP\n"+
		"      pattern := Format('[^.]*%s[^.]*',c);\n" +
		       // substring too greedy if there is a .
		"      IF substring(ref,pattern) IS NOT NULL THEN \n" +
	        "         return true;\n" +
		"      END IF;\n" +
		"   END LOOP;\n"+
		"END LOOP;\n" +

		"RETURN false;\n" +

		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Matching integer ranges

	qstr = "CREATE OR REPLACE FUNCTION match_arrows(arr int,user_set int[])\n"+
		"RETURNS boolean AS $fn$\n" +
		"BEGIN \n" +
		"   IF array_length(user_set,1) IS NULL THEN \n" + // empty arrows
                "      RETURN true;"+
		"   END IF;"+
		"   IF arr = ANY(user_set) THEN \n" + // exact match
		"      RETURN true;\n" +
		"   END IF;\n" +
		"RETURN false;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// Helper to find arrows by type

	qstr = "CREATE OR REPLACE FUNCTION ArrowInList(arrow int,links Link[])\n"+
		"RETURNS boolean AS $fn$\n"+
		"DECLARE \n"+
		"   lnk Link;\n"+
		"BEGIN\n"+
		"IF links IS NULL THEN\n"+
		"   RETURN false;"+
		"END IF;"+
		"FOREACH lnk IN ARRAY links LOOP\n"+
		"  IF lnk.Arr = arrow THEN\n"+
		"     RETURN true;\n"+
		"  END IF;\n"+
		"END LOOP;"+
		"RETURN false;"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// NC version

	qstr = "CREATE OR REPLACE FUNCTION ArrowInContextList(arrow int,links Link[],context text[])\n"+
		"RETURNS boolean AS $fn$\n"+
		"DECLARE \n"+
		"   lnk Link;\n"+
		"BEGIN\n"+
		"IF links IS NULL THEN\n"+
		"   RETURN false;"+
		"END IF;"+
		"FOREACH lnk IN ARRAY links LOOP\n"+
		"  IF lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN\n"+
		"     RETURN true;\n"+
		"  END IF;\n"+
		"END LOOP;"+
		"RETURN false;"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)

	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	//

	qstr =  "CREATE OR REPLACE FUNCTION UnCmp(value text,unacc boolean)\n"+
		"RETURNS text AS $fn$\n"+
		"DECLARE \n"+
		"   retval nodeptr[] = ARRAY[]::nodeptr[];\n"+
		"BEGIN\n"+
		//"  RAISE NOTICE 'VALUE= %',value;\n"+
		"  IF unacc THEN\n"+
		"    RETURN lower(unaccent(value)); \n" +
		"  ELSE\n"+
		"    RETURN lower(value); \n" +
		"  END IF;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("FAILED UnCmp definition\n",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
	// SumAllNCPaths - a filtering version of the SumAllPaths recursive helper function, slower but more powerful

	qstr = "CREATE OR REPLACE FUNCTION SumAllNCPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,chapter text,rm_acc boolean,context text[],exclude NodePtr[],maxlimit int)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    stlinks  Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"    count    int = 0;\n"+
		"    horizon  int = 0;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		// We order the link types to respect the geometry of the temporal links
		// so that (then) will always come last for visual sensemaking

		// Get *All* in/out Links
		"CASE \n" +
		"   WHEN orientation = 'bwd' THEN\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   WHEN orientation = 'fwd' THEN\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"   ELSE\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3,maxlimit);\n" +
		"     fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"END CASE;\n" +

		"horizon := maxlimit - array_length(fwdlinks,1);"+

		"IF horizon < 0 THEN\n"+
		"  horizon = 0;\n"+
		"  maxdepth = depth + 1;"+
		"END IF;\n"+

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL OR count > maxlimit THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"      ELSE\n"+
		"         count = count + 1;"+
		"         IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "            CONTINUE;\n"+
                "         END IF;\n"+

		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumAllNCPaths(lnk,tot_path,orientation,depth+1,maxdepth,chapter,rm_acc,context,exclude,horizon);\n" +

		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"            count = count + regexp_count(appendix,';');"+
		"         ELSE\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
        // with a start set of more than one node

	qstr = "CREATE OR REPLACE FUNCTION AllNCPathsAsLinks(start NodePtr[],chapter text,rm_acc boolean,context text[],orientation text,maxdepth INT,maxlimit int)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   root Text;\n" +
		"   path Text;\n"+
		"   node NodePtr;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = start;\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;\n"+
		"BEGIN\n" +

		// Aggregate array of starting set
		"FOREACH node IN ARRAY start LOOP\n"+
		"   startlnk := GetSingletonAsLink(node);\n"+
		"   path := Format('%s',startlnk::Text);\n"+
		"   root := SumAllNCPaths(startlnk,path,orientation,1,maxdepth,chapter,rm_acc,context,exclude,maxlimit);\n" +
		"   ret_paths := Format('%s\n%s',ret_paths,root);\n"+
		"END LOOP;"+

		"RETURN ret_paths;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // A more detailed path search that includes checks for chapter/context boundaries (NC/C functions)
        // with a start set of more than one node

	qstr = "CREATE OR REPLACE FUNCTION ConstraintPathsAsLinks(start NodePtr[],chapter text,rm_acc boolean,context text[],arrows int[],sttypes int[],maxdepth INT,maxlimit int)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE\n" +
		"   root Text;\n" +
		"   path Text;\n"+
		"   node NodePtr;\n"+
		"   summary_path Text[];\n"+
		"   exclude NodePtr[] = start;\n" +
		"   ret_paths Text;\n" +
		"   startlnk Link;\n"+
		"BEGIN\n" +

		"IF sttypes IS NULL OR array_length(sttypes,1) IS NULL THEN\n" +
		"   sttypes = ARRAY[-3,-2,-1,0,1,2,3];\n" +
		"END IF;\n"+

		// Aggregate array of starting set

		"FOREACH node IN ARRAY start LOOP\n"+
		"   startlnk := GetSingletonAsLink(node);\n"+
		"   path := Format('%s',startlnk::Text);\n"+
		"   root := SumConstraintPaths(startlnk,path,1,maxdepth,chapter,rm_acc,context,arrows,sttypes,exclude,maxlimit);\n" +
		"   ret_paths := Format('%s\n%s',ret_paths,root);\n"+
		"END LOOP;"+

		"RETURN ret_paths;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ...................................................................
	// Now add in the more complex context/chapter filters in searching
	// ...................................................................

        // Generalized path search
	// SumConstraintPaths - a filtering version of the SumAllPaths recursive helper function, slower but more powerful

	qstr = "CREATE OR REPLACE FUNCTION SumConstraintPaths(start Link,path TEXT,depth int,maxdepth INT,chapter text,rm_acc boolean,context text[],arrows int[],sttypes int[],exclude NodePtr[],maxlimit int)\n"+
		"RETURNS Text AS $fn$\n" +
		"DECLARE \n" + 
		"    fwdlinks Link[];\n" +
		"    stlinks  Link[];\n" +
		"    empty Link[] = ARRAY[]::Link[];\n" +
		"    lnk Link;\n" +
		"    fwd Link;\n" +
		"    ret_paths Text;\n" +
		"    appendix Text;\n" +
		"    tot_path Text;\n"+
		"    count    int = 0;\n"+
		"    horizon  int = 0;\n"+
		"    sttype   int;\n"+
		"BEGIN\n" +

		"IF depth = maxdepth THEN\n"+
		"  ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"  RETURN ret_paths;\n"+
		"END IF;\n"+

		"FOREACH sttype IN ARRAY sttypes LOOP\n"+
		"   stlinks := GetConstrainedFwdLinks(start.Dst,chapter,rm_acc,context,exclude,sttype,arrows,maxlimit);\n" +
		"   fwdlinks := array_cat(fwdlinks,stlinks);\n" +
		"END LOOP;\n"+

		"horizon := maxlimit - array_length(fwdlinks,1);"+

		"IF horizon < 0 THEN\n"+
		"  horizon = 0;\n"+
		"  maxdepth = depth + 1;"+
		"END IF;\n"+

		"FOREACH lnk IN ARRAY fwdlinks LOOP \n" +
		"   IF NOT lnk.Dst = ANY(exclude) THEN\n"+
		"      exclude = array_append(exclude,lnk.Dst);\n" +
		"      IF lnk IS NULL OR count > maxlimit THEN\n" +
		"         ret_paths := Format('%s\n%s',ret_paths,path);\n"+
		"      ELSE\n"+
		"         count = count + 1;"+
		"         IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "            CONTINUE;\n"+
                "         END IF;\n"+

		"         tot_path := Format('%s;%s',path,lnk::Text);\n"+
		"         appendix := SumConstraintPaths(lnk,tot_path,depth+1,maxdepth,chapter,rm_acc,context,arrows,sttypes,exclude,horizon);\n" +

		"         IF appendix IS NOT NULL THEN\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,appendix);\n"+
		"            count = count + regexp_count(appendix,';');"+
		"         ELSE\n"+
		"            ret_paths := Format('%s\n%s',ret_paths,tot_path);"+
		"         END IF;\n"+
		"      END IF;\n"+
		"   END IF;\n"+
		"END LOOP;\n"+

		"RETURN ret_paths; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ****
        // Fully filtering version of the neighbour scan
	// ****

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetConstrainedFwdLinks(start NodePtr,chapter text,rm_acc boolean,context text[],exclude NodePtr[],sttype int,arrows int[],maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"   fwdlinks = GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);\n"+

		"   IF fwdlinks IS NULL THEN\n" +
		"       RETURN '{}';\n" +
		"   END IF;\n" +

		"    neighbours := ARRAY[]::Link[];\n" +

		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+

		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+

		"      IF arrows IS NOT NULL AND array_length(arrows,1) > 0 AND NOT lnk.Arr=ANY(arrows) THEN\n"+
		"         CONTINUE;\n"+
		"      END IF;\n"+

                "      IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "         CONTINUE;\n"+
                "      END IF;\n"+

		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 

		"    END LOOP;\n" +

		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ****
        // An NC/C filtering version of the neighbour scan
	// ****

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCFwdLinks(start NodePtr,chapter text,rm_acc boolean,context text[],exclude NodePtr[],sttype int,maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks = GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +
		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+

		"      IF lnk.Arr = 0 THEN\n"+
		"         CONTINUE;"+
		"      END IF;\n"+

                "      IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN\n"+
                "         CONTINUE;\n"+
                "      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 
		"    END LOOP;\n" +
		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

        // This one includes an NCC chapter and context filter so slower! 

	qstr = fmt.Sprintf("CREATE OR REPLACE FUNCTION GetNCCLinks(start NodePtr,exclude NodePtr[],sttype int,chapter text,rm_acc boolean,context text[],maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n" +
		"DECLARE \n" +
		"    neighbours Link[];\n" +
		"    fwdlinks Link[];\n" +
		"    lnk Link;\n" +
		"BEGIN\n" +

		"    fwdlinks =GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);\n"+

		"    IF fwdlinks IS NULL THEN\n" +
		"        RETURN '{}';\n" +
		"    END IF;\n" +
		"    neighbours := ARRAY[]::Link[];\n" +

		"    FOREACH lnk IN ARRAY fwdlinks\n" +
		"    LOOP\n"+
                "      IF context is not NULL AND NOT match_context(lnk.Ctx,context) THEN\n"+
                "        CONTINUE;\n"+
                "      END IF;\n"+
		"      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN\n" +
		"         neighbours := array_append(neighbours, lnk);\n" +
		"      END IF; \n" + 
		"    END LOOP;\n" +
		"    RETURN neighbours; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n")
	
        // This one includes an NC chapter filter
	
	qstr = "CREATE OR REPLACE FUNCTION GetNCNeighboursByType(start NodePtr, chapter text,rm_acc boolean,sttype int,maxlimit int)\n"+
		"RETURNS Link[] AS $fn$\n"+
		"DECLARE \n"+
		"    fwdlinks Link[] := Array[] :: Link[];\n"+
		"    lnk Link := (0,1.0,0,(0,0));\n"+
		"BEGIN\n"+
		"   CASE sttype \n"
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n"+
			"     SELECT %s INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;\n",st,STTypeDBChannel(st));
	}
	
	qstr += "ELSE RAISE EXCEPTION 'No such sttype %', sttype;\n" +
		"END CASE;\n" +
		"    RETURN fwdlinks; \n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	// **************************************
	// Looking for hub / appointed node matroid search
	// **************************************
	
	qstr = "CREATE OR REPLACE FUNCTION GetAppointments(arrow int,sttype int,min int,chaptxt text,context text[],with_accents bool)\n"+

		"RETURNS Appointment[] AS $fn$\n" +
		"DECLARE \n" +
		"    app       Appointment;\n" +
		"    appointed Appointment[];\n" +
		"    this      RECORD;" +
		"    thischap  text;" +
		"    arrscalar text;"+
		"    thisarray Link[];" +
		"    count     int;"+
		"    lnk       Link;"+


		"BEGIN\n" +		
		"   CASE sttype \n"
	
	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf("WHEN %d THEN\n",st);

		qstr += "   IF with_accents THEN\n"
		qstr += fmt.Sprintf("      FOR this IN SELECT NPtr as thptr,Chap as thchap,%s as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)\n",STTypeDBChannel(st));
		qstr += "      LOOP\n" +
		"         count := 0;\n" +
		"         app.NFrom = null;"+
		"         app.NTo = this.thptr::NodePtr;\n" +
		"         app.Chap = this.thchap;\n" +
		"         app.Arr = arrow;"+
		"         app.STType = sttype;"+
		"         app.Ctx = lnk.Ctx;\n\n" +

		"         IF this.chn::Link[] IS NOT NULL THEN\n"+
		"           FOREACH lnk IN ARRAY this.chn::Link[]\n" +
		"           LOOP\n" +
		"	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN\n" +
		"  	          count = count + 1;\n" +
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN\n"+
		"  	          count = count + 1;\n" +
		"                 app.Arr = lnk.Arr;"+
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              END IF;\n" +
		"           END LOOP;\n" +

		"         END IF;\n" +
		
		"         IF count >= min THEN\n" +
		"	    appointed = array_append(appointed,app);\n" +
		"         END IF;\n" +
		"      END LOOP;\n"

		qstr += "   ELSE\n"

		qstr += fmt.Sprintf("      FOR this IN SELECT NPtr as thptr,Chap as thchap,%s as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)\n",STTypeDBChannel(st));
		qstr += "      LOOP\n" +
		"         count := 0;\n" +
		"         app.NFrom = null;"+
		"         app.NTo = this.thptr::NodePtr;\n" +
		"         app.Chap = this.thchap;\n" +
		"         app.Arr = arrow;"+
		"         app.STType = sttype;"+
		"         app.Ctx = lnk.Ctx;\n\n" +

		"         IF this.chn::Link[] IS NOT NULL THEN\n"+
		"           FOREACH lnk IN ARRAY this.chn::Link[]\n" +
		"           LOOP\n" +
		"	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN\n" +
		"  	          count = count + 1;\n" +
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN\n"+
		"  	          count = count + 1;\n" +
		"                 app.Arr = lnk.Arr;"+
		" 	          app.NFrom = array_append(app.NFrom,lnk.Dst);\n" +
		"              END IF;\n" +
		"           END LOOP;\n" +
		"         END IF;\n" +
		
		"         IF count >= min THEN\n" +
		"	    appointed = array_append(appointed,app);\n" +
		"         END IF;\n" +
		"      END LOOP;\n" +

		"   END IF;\n"
	}
	
	qstr += "END CASE;\n"
	qstr += "    RETURN appointed;\n"
	qstr += "END ;\n"
	qstr += "$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

	// **************************************
	// Maintenance/deletion transactions
	// **************************************

	qstr = "CREATE OR REPLACE FUNCTION DeleteChapter(chapter text)\n"+
		"RETURNS boolean AS $fn$\n" +
		"DECLARE\n" +
		"   marked    NodePtr[];\n"+
		"   autoset   NodePtr[];\n"+
		"   nnptr     NodePtr;\n"+
		"   lnk       Link;\n"+
		"   links     Link[];\n"+
		"   ed_list   Link[];\n"+
		"   oleft     text;\n"+
		"   oright    text;\n"+
		"   chaparray text[];\n"+
		"   chaplist  text;\n"+
	        "   ed_chap   text;\n"+
		"   chp       text;\n"+
	
		"BEGIN \n"+

		// First get all NPtrs contained in the chapter for deletion
		// To avoid deleting overlaps, select only the automorphic links

		"chp := Format('%%%s%%',chapter);\n"+
		"SELECT array_agg(NPtr) into autoset FROM Node WHERE Chap LIKE chp;\n"+

		"IF autoset IS NULL THEN\n"+
		"   RETURN false;\n"+
		"END IF;\n"+

		// Look for overlapping chapters

		"oleft := Format('%%%s,%%',chapter);\n"+
		"oright := Format('%%,%s%%',chapter);\n"+

		"SELECT array_agg(NPtr) into marked FROM Node WHERE Chap LIKE oleft OR Chap LIKE oright;\n"+

		"IF marked IS NULL THEN\n"+
		"   DELETE FROM Node WHERE Chap = chapter;\n"+
		"   RETURN true;\n"+
		"END IF;\n"+

		"FOREACH nnptr IN ARRAY marked LOOP\n"+
		"   SELECT Chap into chaplist FROM Node WHERE NPtr = nnptr;\n"+
		"   chaparray = string_to_array(chaplist,',');\n"+

		// Remove the chapter reference
		"IF chaparray IS NOT NULL AND array_length(chaparray,1) > 1 THEN"+
		"   FOREACH chp IN ARRAY chaparray LOOP\n"+
		"      IF NOT chp = chapter THEN"+
		"         IF length(ed_chap) > 0 THEN\n"+
		"            ed_chap = Format('%s,%s',ed_chap,chp);\n"+
		"         ELSE"+
		"            ed_chap = chp;"+
		"         END IF;"+
		"      END IF;"+
		"   END LOOP;"+
		"   UPDATE Node SET Chap = ed_chap WHERE NPtr = nnptr;\n"+
		"   marked = array_remove(marked,nnptr);"+
		"END IF;\n"

	for st := -EXPRESS; st <= EXPRESS; st++ {
		qstr += fmt.Sprintf(
			
			"SELECT %s into links FROM Node WHERE NPtr = nnptr;\n"+

			"   IF links IS NOT NULL THEN\n"+
			"      ed_list = ARRAY[]::Link[];\n"+         // delete reference links
			"      FOREACH lnk in ARRAY links LOOP\n"+
			"         IF NOT lnk.Dst = ANY(marked) THEN\n"+
			"            ed_list = array_append(ed_list,lnk);\n"+
			"         END IF;\n"+
			"      END LOOP;\n"+
			"      UPDATE Node SET %s = ed_list WHERE NPtr = nnptr;\n"+
			"   END IF;\n",
			STTypeDBChannel(st),STTypeDBChannel(st))
	}
	
	qstr += "END LOOP;\n"+

		"DELETE FROM Node WHERE Nptr = ANY(marked);\n"+
		"DELETE FROM Node WHERE Chap = chapter;\n"+

		"RETURN true;\n" +
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"

	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}

	row.Close()

	// ************ LAST SEEN **************'

	qstr = "CREATE OR REPLACE FUNCTION LastSawSection(this text)\n"+
		"RETURNS bool AS $fn$\n"+
		"DECLARE \n"+
		"  prev      timestamp = NOW();\n"+
		"  prevdelta int;\n"+
		"  deltat    int;\n"+
		"  avdeltat  real;\n"+
		"  nowt      int;\n"+
		"  f         int = 0;"+
		"BEGIN\n"+
		"  SELECT last,EXTRACT(EPOCH FROM NOW()-last),delta,freq INTO prev,deltat,prevdelta,f FROM LastSeen WHERE section=this;\n"+
		"  IF NOT FOUND THEN\n"+
		"     INSERT INTO LastSeen (section,first,last,delta,freq,nptr) VALUES (this,NOW(),NOW(),0,1,'(-1,-1)');\n"+
		"  ELSE\n"+
		"     avdeltat = 0.5 * deltat::real + 0.5 * prevdelta::real;\n"+
		"     f = f + 1;\n"+
		      // 1 minute dead time
		"     IF deltat > 60 THEN\n"+
		"       UPDATE LastSeen SET last=NOW(),delta=avdeltat,freq=f WHERE section = this;\n"+
		"     ELSE\n"+
		"        return false;\n"+
		"     END IF;\n"+
		"  END IF;\n"+
		"  RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()
	
	qstr = "CREATE OR REPLACE FUNCTION LastSawNPtr(this NodePtr,name text)\n"+
		"RETURNS bool AS $fn$\n"+
		"DECLARE \n"+
		"  prev      timestamp = NOW();\n"+
		"  prevdelta int;\n"+
		"  avdeltat  real;\n"+
		"  deltat    int;\n"+
		"  nowt      int;\n"+
		"  ep        int = 0;"+
		"  f         int = 0;"+
		"BEGIN\n"+
		"  SELECT last,EXTRACT(EPOCH FROM NOW()-last),delta,freq INTO prev,deltat,prevdelta,f FROM LastSeen WHERE nptr=this;\n"+
		"  IF NOT FOUND THEN\n"+
		"     INSERT INTO LastSeen (section,nptr,first,last,freq,delta) VALUES (name,this,NOW(),NOW(),1,0);\n"+
		"  ELSE\n"+
		"     avdeltat = 0.5 * deltat::real + 0.5 * prevdelta::real;\n"+
		"     f = f + 1;\n"+
		      // 1 minute dead time
		"     IF deltat > 60 THEN\n"+
		"        UPDATE LastSeen SET last=NOW(),delta=avdeltat,freq=f WHERE nptr = this;\n"+
		"     ELSE\n"+
		"        return false;\n"+
		"     END IF;\n"+
		"  END IF;\n"+
		"  RETURN true;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

	// Finally an immutable wrapper

	qstr = "CREATE OR REPLACE FUNCTION sst_unaccent(this text)\n"+
		"RETURNS text AS $fn$\n"+
		"DECLARE \n"+
		"  s text;\n"+
		"BEGIN\n"+
		"  s = unaccent(this);\n"+
		"  RETURN s;\n"+
		"END ;\n" +
		"$fn$ LANGUAGE plpgsql IMMUTABLE;\n"
	
	row,err = sst.DB.Query(qstr)
	
	if err != nil {
		fmt.Println("Error defining postgres function:",qstr,err)
	}
	
	row.Close()

}





//
// postgres_types_functions.go
//

