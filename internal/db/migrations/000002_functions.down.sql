-- Drop graph functions installed in 000002 (names lowercased in PG)
DO $$ DECLARE r record;
BEGIN
  FOR r IN
    SELECT p.oid::regprocedure AS sig
    FROM pg_proc p
    JOIN pg_namespace n ON n.oid = p.pronamespace
    WHERE n.nspname = 'public'
      AND p.proname IN (
        'idempinsertnode','insertnode','idempappendnode','idempinsertcontext',
        'ncc_match','getsingletonaslinkarray','getsingletonaslink','getneighboursbytype',
        'getfwdnodes','getfwdlinks','fwdconeasnodes','fwdconeaslinks','fwdpathsaslinks',
        'sumfwdpaths','allpathsaslinks','sumallpaths','empty_path','match_context',
        'match_arrows','arrowinlist','arrowincontextlist','uncmp','sumallncpaths',
        'allncpathsaslinks','constraintpathsaslinks','sumconstraintpaths',
        'getconstrainedfwdlinks','getncfwdlinks','getncclinks','getncneighboursbytype',
        'getappointments','deletechapter','lastsawsection','lastsawnptr'
      )
  LOOP
    EXECUTE 'DROP FUNCTION IF EXISTS ' || r.sig || ' CASCADE';
  END LOOP;
END $$;
