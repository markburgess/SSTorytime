-- Generated from DefineStoredFunctions — exact function bodies
CREATE OR REPLACE FUNCTION IdempInsertNode(iLi INT, iszchani INT, icptri INT, iSi TEXT, ichapi TEXT)
RETURNS TABLE (    
    ret_cptr INTEGER,    ret_channel INTEGER) AS $fn$ DECLARE 
BEGIN
  IF NOT EXISTS (SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE lower(s) = lower(iSi)) THEN
     INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap,Im3,Im2,Im1,In0,Il1,Ic2,Ie3) VALUES (iszchani,icptri,iLi,iSi,ichapi,'{}','{}','{}','{}','{}','{}','{}');  END IF;
  RETURN QUERY SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION InsertNode(iLi INT, iszchani INT, icptri INT, iSi TEXT, ichapi TEXT,sequence boolean)
RETURNS bool AS $fn$ DECLARE 
BEGIN
   INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap,Seq,Im3,Im2,Im1,In0,Il1,Ic2,Ie3) VALUES (iszchani,icptri,iLi,iSi,ichapi,sequence,'{}','{}','{}','{}','{}','{}','{}');   RETURN true;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION IdempAppendNode(iLi INT, iszchani INT, iSi TEXT, ichapi TEXT)
RETURNS TABLE (    
    ret_cptr INTEGER,    ret_channel INTEGER) AS $fn$ DECLARE 
    icptri INT = 0;BEGIN
  IF NOT EXISTS (SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi) THEN
     SELECT max((Nptr).CPtr) INTO icptri FROM Node WHERE (Nptr).Chan=iszchani;
     IF icptri IS NULL THEN         icptri = 0;     END IF;     INSERT INTO Node (Nptr.Chan,Nptr.Cptr,L,S,chap) VALUES (iszchani,icptri+1,iLi,iSi,ichapi);  END IF;
  RETURN QUERY SELECT (NPtr).Chan,(NPtr).CPtr FROM Node WHERE s = iSi;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION IdempInsertContext(constr text,conptr int)
RETURNS int AS $fn$ DECLARE
    cptr INT = 0;
BEGIN
IF conptr=-1 THEN
   SELECT COALESCE(max(CtxPtr), 0) INTO cptr FROM ContextDirectory;
   INSERT INTO ContextDirectory (Context,CtxPtr) VALUES (constr,cptr+1);
   RETURN cptr+1;
END IF;
IF NOT EXISTS (SELECT CtxPtr FROM ContextDirectory WHERE CtxPtr=conptr OR Context=constr) THEN
   INSERT INTO ContextDirectory (Context,CtxPtr) VALUES (constr,conptr);
   RETURN conptr;
END IF;
SELECT CtxPtr INTO cptr FROM ContextDirectory WHERE CtxPtr=conptr OR Context=constr;
RETURN cptr;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION NCC_match(thisnptr NodePtr,context text[],arrows int[],sttypes int[],lm3 Link[],lm2 Link[],lm1 Link[],ln0 Link[],lp1 Link[],lp2 Link[],lp3 Link[])
RETURNS boolean AS $fn$
DECLARE 
    emptyarray Link[] := Array[] :: Link[];
    lnkarray Link[] := Array[] :: Link[];
    lnk Link;
    st int;
BEGIN
IF array_length(arrows,1) IS NULL THEN
   IF lp1 IS NOT NULL THEN      FOREACH lnk IN ARRAY lp1 LOOP
         IF lnk.Arr = 0 AND match_context(lnk.Ctx,context) THEN
            RETURN true;         END IF;      END LOOP;
   END IF;
ELSE
   FOREACH st IN ARRAY sttypes LOOP
      CASE st 
   WHEN -3 THEN
         SELECT Im3 INTO lnkarray FROM Node WHERE Nptr=thisnptr;
   WHEN -2 THEN
         SELECT Im2 INTO lnkarray FROM Node WHERE Nptr=thisnptr;
   WHEN -1 THEN
         SELECT Im1 INTO lnkarray FROM Node WHERE Nptr=thisnptr;
   WHEN 0 THEN
         SELECT In0 INTO lnkarray FROM Node WHERE Nptr=thisnptr;
   WHEN 1 THEN
         SELECT Il1 INTO lnkarray FROM Node WHERE Nptr=thisnptr;
   WHEN 2 THEN
         SELECT Ic2 INTO lnkarray FROM Node WHERE Nptr=thisnptr;
   WHEN 3 THEN
         SELECT Ie3 INTO lnkarray FROM Node WHERE Nptr=thisnptr;
      ELSE RAISE EXCEPTION 'No such sttype in NCC_match %', sttype;
      END CASE;
      FOREACH lnk IN ARRAY lnkarray LOOP
         IF match_arrows(lnk.arr,arrows) AND match_context(lnk.ctx,context) THEN
            RETURN true;
         END IF;
      END LOOP;
   END LOOP;
END IF;
RETURN false; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetSingletonAsLinkArray(start NodePtr)
RETURNS Link[] AS $fn$
DECLARE 
    level Link[] := Array[] :: Link[];
    lnk Link := (0,1.0,0,(0,0));
BEGIN
 lnk.Dst = start;
 level = array_append(level,lnk);
RETURN level; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetSingletonAsLink(start NodePtr)
RETURNS Link AS $fn$
DECLARE 
    lnk Link := (0,1.0,0,(0,0));
BEGIN
 lnk.Dst = start;
RETURN lnk; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetNeighboursByType(start NodePtr, sttype int, maxlimit int)
RETURNS Link[] AS $fn$
DECLARE 
    fwdlinks Link[] := Array[] :: Link[];
    lnk Link := (0,1.0,0,(0,0));
BEGIN
   CASE sttype 
WHEN -3 THEN
     SELECT Im3 INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;
WHEN -2 THEN
     SELECT Im2 INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;
WHEN -1 THEN
     SELECT Im1 INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;
WHEN 0 THEN
     SELECT In0 INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;
WHEN 1 THEN
     SELECT Il1 INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;
WHEN 2 THEN
     SELECT Ic2 INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;
WHEN 3 THEN
     SELECT Ie3 INTO fwdlinks FROM Node WHERE Nptr=start AND NOT L=0 LIMIT maxlimit;
ELSE RAISE EXCEPTION 'No such sttype %', sttype;
END CASE;
    RETURN fwdlinks; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetFwdNodes(start NodePtr,exclude NodePtr[],sttype int,maxlimit int)
RETURNS NodePtr[] AS $fn$
DECLARE 
    neighbours NodePtr[];
    fwdlinks Link[];
    lnk Link;
BEGIN
    fwdlinks = GetNeighboursByType(start,sttype,maxlimit);
    IF fwdlinks IS NULL THEN
        RETURN '{}';
    END IF;
    neighbours := ARRAY[]::NodePtr[];
    FOREACH lnk IN ARRAY fwdlinks
    LOOP
      IF lnk.Arr = 0 THEN
         CONTINUE;      END IF;
      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN
         neighbours := array_append(neighbours, lnk.dst);
      END IF; 
    END LOOP;
    RETURN neighbours; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetFwdLinks(start NodePtr,exclude NodePtr[],sttype int,maxlimit int)
RETURNS Link[] AS $fn$
DECLARE 
    neighbours Link[];
    fwdlinks Link[];
    lnk Link;
BEGIN
    fwdlinks = GetNeighboursByType(start,sttype,maxlimit);
    IF fwdlinks IS NULL THEN
        RETURN '{}';
    END IF;
    neighbours := ARRAY[]::Link[];
    FOREACH lnk IN ARRAY fwdlinks
    LOOP
      IF lnk.Arr = 0 THEN
         CONTINUE;      END IF;
      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN
         neighbours := array_append(neighbours, lnk);
      END IF; 
    END LOOP;
    RETURN neighbours; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION FwdConeAsNodes(start NodePtr,sttype INT, maxdepth INT,maxlimit int)
RETURNS NodePtr[] AS $fn$
DECLARE 
    nextlevel NodePtr[];
    partlevel NodePtr[];
    level NodePtr[] = ARRAY[start]::NodePtr[];
    exclude NodePtr[] = ARRAY['(0,0)']::NodePtr[];
    cone NodePtr[];
    neigh NodePtr;
    frn NodePtr;
    counter int := 0;
BEGIN
LOOP
  EXIT WHEN counter = maxdepth+1;
  IF level IS NULL THEN
     RETURN cone;
  END IF;
  nextlevel := ARRAY[]::NodePtr[];
  FOREACH frn IN ARRAY level   LOOP 
     nextlevel = array_append(nextlevel,frn);
  END LOOP;
  IF nextlevel IS NULL THEN
     RETURN cone;
  END IF;
  FOREACH neigh IN ARRAY nextlevel LOOP 
    IF NOT neigh = ANY(exclude) THEN
      cone = array_append(cone,neigh);
      exclude := array_append(exclude,neigh);
      partlevel := GetFwdNodes(neigh,exclude,sttype,maxlimit);
    END IF;    IF partlevel IS NOT NULL THEN
         level = array_cat(level,partlevel);
    END IF;
  END LOOP;
  counter = counter + 1;
END LOOP;
RETURN cone; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION FwdConeAsLinks(start NodePtr,sttype INT,maxdepth INT,maxlimit int)
RETURNS Link[] AS $fn$
DECLARE 
    nextlevel Link[];
    partlevel Link[];
    level Link[] = ARRAY[]::Link[];
    exclude NodePtr[] = ARRAY['(0,0)']::NodePtr[];
    cone Link[];
    neigh Link;
    frn Link;
    counter int := 0;
BEGIN
level := GetSingletonAsLinkArray(start);
LOOP
  EXIT WHEN counter = maxdepth+1;
  IF level IS NULL THEN
     RETURN cone;
  END IF;
  nextlevel := ARRAY[]::Link[];
  FOREACH frn IN ARRAY level   LOOP 
     nextlevel = array_append(nextlevel,frn);
  END LOOP;
  IF nextlevel IS NULL THEN
     RETURN cone;
  END IF;
  FOREACH neigh IN ARRAY nextlevel LOOP 
    IF NOT neigh.Dst = ANY(exclude) THEN
      cone = array_append(cone,neigh);
      exclude := array_append(exclude,neigh.Dst);
      partlevel := GetFwdLinks(neigh.Dst,exclude,sttype,maxlimit);
    END IF;    IF partlevel IS NOT NULL THEN
         level = array_cat(level,partlevel);
    END IF;
  END LOOP;
  counter = counter + 1;
END LOOP;
RETURN cone; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION FwdPathsAsLinks(start NodePtr,sttype INT,maxdepth INT, maxlimit INT)
RETURNS Text AS $fn$
DECLARE
   hop Text;
   path Text;
   summary_path Text[];
   exclude NodePtr[] = ARRAY[start]::NodePtr[];
   ret_paths Text;
   startlnk Link;BEGIN
startlnk := GetSingletonAsLink(start);
path := Format('%s',startlnk::Text);
ret_paths := SumFwdPaths(startlnk,path,sttype,1,maxdepth,exclude, maxlimit);RETURN ret_paths; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION SumFwdPaths(start Link,path TEXT, sttype INT,depth int, maxdepth INT,exclude NodePtr[], maxlimit INT)
RETURNS Text AS $fn$
DECLARE 
    fwdlinks Link[];
    empty Link[] = ARRAY[]::Link[];
    lnk Link;
    fwd Link;
    ret_paths Text;
    appendix Text;
    tot_path Text;
    count    int = 0;
    horizon  int = 0;
BEGIN
IF depth = maxdepth THEN
  ret_paths := Format('%s
%s',ret_paths,path);
  RETURN ret_paths;
END IF;
fwdlinks := GetFwdLinks(start.Dst,exclude,sttype, maxlimit);
horizon := maxlimit - array_length(fwdlinks,1);IF horizon < 0 THEN
  horizon = 0;
  maxdepth = depth + 1;END IF;
FOREACH lnk IN ARRAY fwdlinks LOOP 
   IF NOT lnk.Dst = ANY(exclude) THEN
      exclude = array_append(exclude,lnk.Dst);
      IF lnk IS NULL OR count >= maxlimit THEN         ret_paths := Format('%s
%s',ret_paths,path);
         RETURN ret_paths;      ELSE
         count = count + 1;         tot_path := Format('%s;%s',path,lnk::Text);
         appendix := SumFwdPaths(lnk,tot_path,sttype,depth+1,maxdepth,exclude,horizon);
         IF appendix IS NOT NULL THEN
            ret_paths := Format('%s
%s',ret_paths,appendix);
            count = count + regexp_count(appendix,';');         ELSE            ret_paths := Format('%s
%s',ret_paths,tot_path);         END IF;      END IF;   END IF;END LOOP;RETURN ret_paths; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION AllPathsAsLinks(start NodePtr,orientation text,maxdepth INT, maxlimit INT)
RETURNS Text AS $fn$
DECLARE
   hop Text;
   path Text;
   summary_path Text[];
   exclude NodePtr[] = ARRAY[start]::NodePtr[];
   ret_paths Text;
   startlnk Link;BEGIN
startlnk := GetSingletonAsLink(start);
path := Format('%s',startlnk::Text);
ret_paths := SumAllPaths(startlnk,path,orientation,1,maxdepth,exclude, maxlimit);RETURN ret_paths; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION SumAllPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,exclude NodePtr[],maxlimit int)
RETURNS Text AS $fn$
DECLARE 
    fwdlinks Link[];
    stlinks  Link[];
    empty Link[] = ARRAY[]::Link[];
    lnk Link;
    fwd Link;
    ret_paths Text;
    appendix Text;
    tot_path Text;
    counter int = 0;BEGIN
IF depth = maxdepth THEN
  ret_paths := Format('%s
%s',ret_paths,path);
  RETURN ret_paths;
END IF;
CASE 
   WHEN orientation = 'bwd' THEN
     stlinks := GetFwdLinks(start.Dst,exclude,-3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,-2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,-1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
   WHEN orientation = 'fwd' THEN
     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
   ELSE
     stlinks := GetFwdLinks(start.Dst,exclude,-3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,-2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,-1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,0,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetFwdLinks(start.Dst,exclude,3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
END CASE;
FOREACH lnk IN ARRAY fwdlinks LOOP 
   IF counter > maxlimit THEN
      RETURN ret_paths;   END IF;   IF NOT lnk.Dst = ANY(exclude) THEN
      exclude = array_append(exclude,lnk.Dst);
      IF lnk IS NULL THEN
         ret_paths := Format('%s
%s',ret_paths,path);
         RETURN ret_paths;      ELSE
         tot_path := Format('%s;%s',path,lnk::Text);
         appendix := SumAllPaths(lnk,tot_path,orientation,depth+1,maxdepth,exclude,maxlimit);
         IF appendix IS NOT NULL THEN
            ret_paths := Format('%s
%s',ret_paths,appendix);
         ELSE
            ret_paths := Format('%s
%s',ret_paths,tot_path);         END IF;
         counter = counter + 1;
      END IF;
   END IF;
END LOOP;
RETURN ret_paths; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION empty_path(path text)
RETURNS boolean AS $fn$
BEGIN 
   IF strpos(path,';') THEN 
      RETURN true;
   END IF;
RETURN false;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION match_context(thisctxptr int,user_set text[])
RETURNS boolean AS $fn$
DECLARE
   ctxstr text;
   db_set text[];
   notes text[];
   client text[];
   pattern text;
   and_list text[];
   and_count int;
   and_result int = 0;
   end_result int = 0;
   or_list text[] = ARRAY[]::text[];
   item text;
   item_db text;
   item_us text;
   partial text;
   diff int;
   ref text;
   c text;
BEGIN 
IF array_length(user_set,1) IS NULL THEN
   RETURN true;
END IF;
IF user_set[0] = '' THEN
   RETURN true;
END IF;
SELECT Context INTO ctxstr FROM ContextDirectory WHERE ctxPtr=thisctxptr;db_set = regexp_split_to_array(ctxstr,',');
IF array_length(db_set,1) IS NULL AND array_length(user_set,1) IS NOT NULL THEN
   RETURN false;
END IF;
IF array_length(db_set,1) IS NULL AND array_length(user_set,1) IS NULL THEN
   RETURN true;
END IF;
FOREACH item_db IN ARRAY db_set LOOP
   FOREACH item_us IN ARRAY user_set LOOP
      IF item_db = item_us THEN
         RETURN true;
      END IF;
   END LOOP;
END LOOP;
FOREACH item IN ARRAY db_set LOOP
   notes = array_append(notes,lower(unaccent(item)));
END LOOP;
FOREACH item IN ARRAY user_set LOOP
   client = array_append(client,lower(unaccent(item)));
END LOOP;
FOREACH item IN ARRAY notes LOOP
   and_list = regexp_split_to_array(item, '\.');
   and_count = array_length(and_list,1);
   IF and_count > 1 THEN
      and_result = 0;
      FOREACH ref IN ARRAY and_list LOOP
         FOREACH c IN ARRAY client LOOP
            IF ref = c THEN 
               and_result = and_result + 1;
            END IF;
         END LOOP;
      END LOOP;
      IF and_result = and_count THEN
         end_result = end_result + 1;
      END IF;
   ELSE
      or_list = array_append(or_list,item);
   END IF;
END LOOP;
IF end_result > 0 THEN
   RETURN true;
END IF;
FOREACH ref IN ARRAY or_list LOOP
   FOREACH c IN ARRAY client LOOP
      pattern := Format('%s[^.]*',c);
      partial := substring(ref,pattern);
      diff := length(partial) - length(c);
      IF partial IS NOT NULL AND length(c) >= 4 AND diff < 3 THEN 
         return true;
      END IF;
   END LOOP;
END LOOP;
RETURN false;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION match_arrows(arr int,user_set int[])
RETURNS boolean AS $fn$
BEGIN 
   IF array_length(user_set,1) IS NULL THEN 
      RETURN true;   END IF;   IF arr = ANY(user_set) THEN 
      RETURN true;
   END IF;
RETURN false;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION ArrowInList(arrow int,links Link[])
RETURNS boolean AS $fn$
DECLARE 
   lnk Link;
BEGIN
IF links IS NULL THEN
   RETURN false;END IF;FOREACH lnk IN ARRAY links LOOP
  IF lnk.Arr = arrow THEN
     RETURN true;
  END IF;
END LOOP;RETURN false;END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION ArrowInContextList(arrow int,links Link[],context text[])
RETURNS boolean AS $fn$
DECLARE 
   lnk Link;
BEGIN
IF links IS NULL THEN
   RETURN false;END IF;FOREACH lnk IN ARRAY links LOOP
  IF lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
     RETURN true;
  END IF;
END LOOP;RETURN false;END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION UnCmp(value text,unacc boolean)
RETURNS text AS $fn$
DECLARE 
   retval nodeptr[] = ARRAY[]::nodeptr[];
BEGIN
  IF unacc THEN
    RETURN lower(unaccent(value)); 
  ELSE
    RETURN lower(value); 
  END IF;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION SumAllNCPaths(start Link,path TEXT,orientation text,depth int, maxdepth INT,chapter text,rm_acc boolean,context text[],exclude NodePtr[],maxlimit int)
RETURNS Text AS $fn$
DECLARE 
    fwdlinks Link[];
    stlinks  Link[];
    empty Link[] = ARRAY[]::Link[];
    lnk Link;
    fwd Link;
    ret_paths Text;
    appendix Text;
    tot_path Text;
    count    int = 0;
    horizon  int = 0;
BEGIN
IF depth = maxdepth THEN
  ret_paths := Format('%s
%s',ret_paths,path);
  RETURN ret_paths;
END IF;
CASE 
   WHEN orientation = 'bwd' THEN
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
   WHEN orientation = 'fwd' THEN
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
   ELSE
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,-1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,0,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,1,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,2,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
     stlinks := GetNCFwdLinks(start.Dst,chapter,rm_acc,context,exclude,3,maxlimit);
     fwdlinks := array_cat(fwdlinks,stlinks);
END CASE;
horizon := maxlimit - array_length(fwdlinks,1);IF horizon < 0 THEN
  horizon = 0;
  maxdepth = depth + 1;END IF;
FOREACH lnk IN ARRAY fwdlinks LOOP 
   IF NOT lnk.Dst = ANY(exclude) THEN
      exclude = array_append(exclude,lnk.Dst);
      IF lnk IS NULL OR count > maxlimit THEN
         ret_paths := Format('%s
%s',ret_paths,path);
      ELSE
         count = count + 1;         IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN
            CONTINUE;
         END IF;
         tot_path := Format('%s;%s',path,lnk::Text);
         appendix := SumAllNCPaths(lnk,tot_path,orientation,depth+1,maxdepth,chapter,rm_acc,context,exclude,horizon);
         IF appendix IS NOT NULL THEN
            ret_paths := Format('%s
%s',ret_paths,appendix);
            count = count + regexp_count(appendix,';');         ELSE
            ret_paths := Format('%s
%s',ret_paths,tot_path);         END IF;
      END IF;
   END IF;
END LOOP;
RETURN ret_paths; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION AllNCPathsAsLinks(start NodePtr[],chapter text,rm_acc boolean,context text[],orientation text,maxdepth INT,maxlimit int)
RETURNS Text AS $fn$
DECLARE
   root Text;
   path Text;
   node NodePtr;
   summary_path Text[];
   exclude NodePtr[] = start;
   ret_paths Text;
   startlnk Link;
BEGIN
FOREACH node IN ARRAY start LOOP
   startlnk := GetSingletonAsLink(node);
   path := Format('%s',startlnk::Text);
   root := SumAllNCPaths(startlnk,path,orientation,1,maxdepth,chapter,rm_acc,context,exclude,maxlimit);
   ret_paths := Format('%s
%s',ret_paths,root);
END LOOP;RETURN ret_paths;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION ConstraintPathsAsLinks(start NodePtr[],chapter text,rm_acc boolean,context text[],arrows int[],sttypes int[],maxdepth INT,maxlimit int)
RETURNS Text AS $fn$
DECLARE
   root Text;
   path Text;
   node NodePtr;
   summary_path Text[];
   exclude NodePtr[] = start;
   ret_paths Text;
   startlnk Link;
BEGIN
IF sttypes IS NULL OR array_length(sttypes,1) IS NULL THEN
   sttypes = ARRAY[-3,-2,-1,0,1,2,3];
END IF;
FOREACH node IN ARRAY start LOOP
   startlnk := GetSingletonAsLink(node);
   path := Format('%s',startlnk::Text);
   root := SumConstraintPaths(startlnk,path,1,maxdepth,chapter,rm_acc,context,arrows,sttypes,exclude,maxlimit);
   ret_paths := Format('%s
%s',ret_paths,root);
END LOOP;RETURN ret_paths;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION SumConstraintPaths(start Link,path TEXT,depth int,maxdepth INT,chapter text,rm_acc boolean,context text[],arrows int[],sttypes int[],exclude NodePtr[],maxlimit int)
RETURNS Text AS $fn$
DECLARE 
    fwdlinks Link[];
    stlinks  Link[];
    empty Link[] = ARRAY[]::Link[];
    lnk Link;
    fwd Link;
    ret_paths Text;
    appendix Text;
    tot_path Text;
    count    int = 0;
    horizon  int = 0;
    sttype   int;
BEGIN
IF depth = maxdepth THEN
  ret_paths := Format('%s
%s',ret_paths,path);
  RETURN ret_paths;
END IF;
FOREACH sttype IN ARRAY sttypes LOOP
   stlinks := GetConstrainedFwdLinks(start.Dst,chapter,rm_acc,context,exclude,sttype,arrows,maxlimit);
   fwdlinks := array_cat(fwdlinks,stlinks);
END LOOP;
horizon := maxlimit - array_length(fwdlinks,1);IF horizon < 0 THEN
  horizon = 0;
  maxdepth = depth + 1;END IF;
FOREACH lnk IN ARRAY fwdlinks LOOP 
   IF NOT lnk.Dst = ANY(exclude) THEN
      exclude = array_append(exclude,lnk.Dst);
      IF lnk IS NULL OR count > maxlimit THEN
         ret_paths := Format('%s
%s',ret_paths,path);
      ELSE
         count = count + 1;         IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN
            CONTINUE;
         END IF;
         tot_path := Format('%s;%s',path,lnk::Text);
         appendix := SumConstraintPaths(lnk,tot_path,depth+1,maxdepth,chapter,rm_acc,context,arrows,sttypes,exclude,horizon);
         IF appendix IS NOT NULL THEN
            ret_paths := Format('%s
%s',ret_paths,appendix);
            count = count + regexp_count(appendix,';');         ELSE
            ret_paths := Format('%s
%s',ret_paths,tot_path);         END IF;
      END IF;
   END IF;
END LOOP;
RETURN ret_paths; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetConstrainedFwdLinks(start NodePtr,chapter text,rm_acc boolean,context text[],exclude NodePtr[],sttype int,arrows int[],maxlimit int)
RETURNS Link[] AS $fn$
DECLARE 
    neighbours Link[];
    fwdlinks Link[];
    lnk Link;
BEGIN
   fwdlinks = GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);
   IF fwdlinks IS NULL THEN
       RETURN '{}';
   END IF;
    neighbours := ARRAY[]::Link[];
    FOREACH lnk IN ARRAY fwdlinks
    LOOP
      IF lnk.Arr = 0 THEN
         CONTINUE;      END IF;
      IF arrows IS NOT NULL AND array_length(arrows,1) > 0 AND NOT lnk.Arr=ANY(arrows) THEN
         CONTINUE;
      END IF;
      IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN
         CONTINUE;
      END IF;
      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN
         neighbours := array_append(neighbours, lnk);
      END IF; 
    END LOOP;
    RETURN neighbours; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetNCFwdLinks(start NodePtr,chapter text,rm_acc boolean,context text[],exclude NodePtr[],sttype int,maxlimit int)
RETURNS Link[] AS $fn$
DECLARE 
    neighbours Link[];
    fwdlinks Link[];
    lnk Link;
BEGIN
    fwdlinks = GetNCNeighboursByType(start,chapter,rm_acc,sttype,maxlimit);
    IF fwdlinks IS NULL THEN
        RETURN '{}';
    END IF;
    neighbours := ARRAY[]::Link[];
    FOREACH lnk IN ARRAY fwdlinks
    LOOP
      IF lnk.Arr = 0 THEN
         CONTINUE;      END IF;
      IF context is not NULL AND NOT match_context(lnk.Ctx,context::text[]) THEN
         CONTINUE;
      END IF;
      IF exclude is not NULL AND NOT lnk.dst=ANY(exclude) THEN
         neighbours := array_append(neighbours, lnk);
      END IF; 
    END LOOP;
    RETURN neighbours; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetNCNeighboursByType(start NodePtr, chapter text,rm_acc boolean,sttype int,maxlimit int)
RETURNS Link[] AS $fn$
DECLARE 
    fwdlinks Link[] := Array[] :: Link[];
    lnk Link := (0,1.0,0,(0,0));
BEGIN
   CASE sttype 
WHEN -3 THEN
     SELECT Im3 INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;
WHEN -2 THEN
     SELECT Im2 INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;
WHEN -1 THEN
     SELECT Im1 INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;
WHEN 0 THEN
     SELECT In0 INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;
WHEN 1 THEN
     SELECT Il1 INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;
WHEN 2 THEN
     SELECT Ic2 INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;
WHEN 3 THEN
     SELECT Ie3 INTO fwdlinks FROM Node WHERE NOT L=0 AND Nptr=start AND UnCmp(Chap,rm_acc) LIKE lower(chapter) LIMIT maxlimit;
ELSE RAISE EXCEPTION 'No such sttype %', sttype;
END CASE;
    RETURN fwdlinks; 
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION GetAppointments(arrow int,sttype int,min int,chaptxt text,context text[],with_accents bool)
RETURNS Appointment[] AS $fn$
DECLARE 
    app       Appointment;
    appointed Appointment[];
    this      RECORD;    thischap  text;    arrscalar text;    thisarray Link[];    count     int;    lnk       Link;BEGIN
   CASE sttype 
WHEN -3 THEN
   IF with_accents THEN
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Im3 as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   ELSE
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Im3 as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   END IF;
WHEN -2 THEN
   IF with_accents THEN
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Im2 as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   ELSE
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Im2 as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   END IF;
WHEN -1 THEN
   IF with_accents THEN
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Im1 as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   ELSE
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Im1 as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   END IF;
WHEN 0 THEN
   IF with_accents THEN
      FOR this IN SELECT NPtr as thptr,Chap as thchap,In0 as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   ELSE
      FOR this IN SELECT NPtr as thptr,Chap as thchap,In0 as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   END IF;
WHEN 1 THEN
   IF with_accents THEN
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Il1 as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   ELSE
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Il1 as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   END IF;
WHEN 2 THEN
   IF with_accents THEN
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Ic2 as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   ELSE
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Ic2 as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   END IF;
WHEN 3 THEN
   IF with_accents THEN
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Ie3 as chn FROM Node WHERE lower(unaccent(chap)) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   ELSE
      FOR this IN SELECT NPtr as thptr,Chap as thchap,Ie3 as chn FROM Node WHERE lower(chap) LIKE lower(chaptxt)
      LOOP
         count := 0;
         app.NFrom = null;         app.NTo = this.thptr::NodePtr;
         app.Chap = this.thchap;
         app.Arr = arrow;         app.STType = sttype;         app.Ctx = lnk.Ctx;

         IF this.chn::Link[] IS NOT NULL THEN
           FOREACH lnk IN ARRAY this.chn::Link[]
           LOOP
	       IF arrow > 0 AND lnk.Arr = arrow AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              ELSIF arrow < 0 AND match_context(lnk.Ctx,context) THEN
  	          count = count + 1;
                 app.Arr = lnk.Arr; 	          app.NFrom = array_append(app.NFrom,lnk.Dst);
              END IF;
           END LOOP;
         END IF;
         IF count >= min THEN
	    appointed = array_append(appointed,app);
         END IF;
      END LOOP;
   END IF;
END CASE;
    RETURN appointed;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION DeleteChapter(chapter text)
RETURNS boolean AS $fn$
DECLARE
   marked    NodePtr[];
   autoset   NodePtr[];
   nnptr     NodePtr;
   lnk       Link;
   links     Link[];
   ed_list   Link[];
   oleft     text;
   oright    text;
   chaparray text[];
   chaplist  text;
   ed_chap   text;
   chp       text;
BEGIN 
chp := Format('%%%s%%',chapter);
SELECT array_agg(NPtr) into autoset FROM Node WHERE Chap LIKE chp;
IF autoset IS NULL THEN
   RETURN false;
END IF;
oleft := Format('%%%s,%%',chapter);
oright := Format('%%,%s%%',chapter);
SELECT array_agg(NPtr) into marked FROM Node WHERE Chap LIKE oleft OR Chap LIKE oright;
IF marked IS NULL THEN
   DELETE FROM Node WHERE Chap = chapter;
   RETURN true;
END IF;
FOREACH nnptr IN ARRAY marked LOOP
   SELECT Chap into chaplist FROM Node WHERE NPtr = nnptr;
   chaparray = string_to_array(chaplist,',');
IF chaparray IS NOT NULL AND array_length(chaparray,1) > 1 THEN   FOREACH chp IN ARRAY chaparray LOOP
      IF NOT chp = chapter THEN         IF length(ed_chap) > 0 THEN
            ed_chap = Format('%s,%s',ed_chap,chp);
         ELSE            ed_chap = chp;         END IF;      END IF;   END LOOP;   UPDATE Node SET Chap = ed_chap WHERE NPtr = nnptr;
   marked = array_remove(marked,nnptr);END IF;
SELECT Im3 into links FROM Node WHERE NPtr = nnptr;
   IF links IS NOT NULL THEN
      ed_list = ARRAY[]::Link[];
      FOREACH lnk in ARRAY links LOOP
         IF NOT lnk.Dst = ANY(marked) THEN
            ed_list = array_append(ed_list,lnk);
         END IF;
      END LOOP;
      UPDATE Node SET Im3 = ed_list WHERE NPtr = nnptr;
   END IF;
SELECT Im2 into links FROM Node WHERE NPtr = nnptr;
   IF links IS NOT NULL THEN
      ed_list = ARRAY[]::Link[];
      FOREACH lnk in ARRAY links LOOP
         IF NOT lnk.Dst = ANY(marked) THEN
            ed_list = array_append(ed_list,lnk);
         END IF;
      END LOOP;
      UPDATE Node SET Im2 = ed_list WHERE NPtr = nnptr;
   END IF;
SELECT Im1 into links FROM Node WHERE NPtr = nnptr;
   IF links IS NOT NULL THEN
      ed_list = ARRAY[]::Link[];
      FOREACH lnk in ARRAY links LOOP
         IF NOT lnk.Dst = ANY(marked) THEN
            ed_list = array_append(ed_list,lnk);
         END IF;
      END LOOP;
      UPDATE Node SET Im1 = ed_list WHERE NPtr = nnptr;
   END IF;
SELECT In0 into links FROM Node WHERE NPtr = nnptr;
   IF links IS NOT NULL THEN
      ed_list = ARRAY[]::Link[];
      FOREACH lnk in ARRAY links LOOP
         IF NOT lnk.Dst = ANY(marked) THEN
            ed_list = array_append(ed_list,lnk);
         END IF;
      END LOOP;
      UPDATE Node SET In0 = ed_list WHERE NPtr = nnptr;
   END IF;
SELECT Il1 into links FROM Node WHERE NPtr = nnptr;
   IF links IS NOT NULL THEN
      ed_list = ARRAY[]::Link[];
      FOREACH lnk in ARRAY links LOOP
         IF NOT lnk.Dst = ANY(marked) THEN
            ed_list = array_append(ed_list,lnk);
         END IF;
      END LOOP;
      UPDATE Node SET Il1 = ed_list WHERE NPtr = nnptr;
   END IF;
SELECT Ic2 into links FROM Node WHERE NPtr = nnptr;
   IF links IS NOT NULL THEN
      ed_list = ARRAY[]::Link[];
      FOREACH lnk in ARRAY links LOOP
         IF NOT lnk.Dst = ANY(marked) THEN
            ed_list = array_append(ed_list,lnk);
         END IF;
      END LOOP;
      UPDATE Node SET Ic2 = ed_list WHERE NPtr = nnptr;
   END IF;
SELECT Ie3 into links FROM Node WHERE NPtr = nnptr;
   IF links IS NOT NULL THEN
      ed_list = ARRAY[]::Link[];
      FOREACH lnk in ARRAY links LOOP
         IF NOT lnk.Dst = ANY(marked) THEN
            ed_list = array_append(ed_list,lnk);
         END IF;
      END LOOP;
      UPDATE Node SET Ie3 = ed_list WHERE NPtr = nnptr;
   END IF;
END LOOP;
DELETE FROM Node WHERE Nptr = ANY(marked);
DELETE FROM Node WHERE Chap = chapter;
RETURN true;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION LastSawSection(this text)
RETURNS bool AS $fn$
DECLARE 
  prev      timestamp = NOW();
  prevdelta int;
  deltat    int;
  avdeltat  real;
  nowt      int;
  f         int = 0;BEGIN
  SELECT last,EXTRACT(EPOCH FROM NOW()-last),delta,freq INTO prev,deltat,prevdelta,f FROM LastSeen WHERE section=this;
  IF NOT FOUND THEN
     INSERT INTO LastSeen (section,first,last,delta,freq,nptr) VALUES (this,NOW(),NOW(),0,1,'(-1,-1)');
  ELSE
     avdeltat = 0.5 * deltat::real + 0.5 * prevdelta::real;
     f = f + 1;
     IF deltat > 60 THEN
       UPDATE LastSeen SET last=NOW(),delta=avdeltat,freq=f WHERE section = this;
     ELSE
        return false;
     END IF;
  END IF;
  RETURN true;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION LastSawNPtr(this NodePtr,name text)
RETURNS bool AS $fn$
DECLARE 
  prev      timestamp = NOW();
  prevdelta int;
  avdeltat  real;
  deltat    int;
  nowt      int;
  ep        int = 0;  f         int = 0;BEGIN
  SELECT last,EXTRACT(EPOCH FROM NOW()-last),delta,freq INTO prev,deltat,prevdelta,f FROM LastSeen WHERE nptr=this;
  IF NOT FOUND THEN
     INSERT INTO LastSeen (section,nptr,first,last,freq,delta) VALUES (name,this,NOW(),NOW(),1,0);
  ELSE
     avdeltat = 0.5 * deltat::real + 0.5 * prevdelta::real;
     f = f + 1;
     IF deltat > 60 THEN
        UPDATE LastSeen SET last=NOW(),delta=avdeltat,freq=f WHERE nptr = this;
     ELSE
        return false;
     END IF;
  END IF;
  RETURN true;
END ;
$fn$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sst_unaccent(this text)
RETURNS text AS $fn$
DECLARE 
  s text;
BEGIN
  s = unaccent(this);
  RETURN s;
END ;
$fn$ LANGUAGE plpgsql IMMUTABLE;

