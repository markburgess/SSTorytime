DROP TABLE IF EXISTS LastSeen CASCADE;
DROP TABLE IF EXISTS ArrowDirectory CASCADE;
DROP TABLE IF EXISTS ArrowInverses CASCADE;
DROP TABLE IF EXISTS Node CASCADE;
DROP TABLE IF EXISTS PageMap CASCADE;
DROP TABLE IF EXISTS Bookmarks CASCADE;
DROP TABLE IF EXISTS ContextDirectory CASCADE;
DROP FUNCTION IF EXISTS sst_unaccent(text);
DROP TYPE IF EXISTS Appointment CASCADE;
DROP TYPE IF EXISTS Link CASCADE;
DROP TYPE IF EXISTS NodePtr CASCADE;
-- PL/pgSQL helpers dropped on type/table cascade where dependent; leftover funcs:
DO $$ DECLARE r record;
BEGIN
  FOR r IN
    SELECT oid::regprocedure AS p
    FROM pg_proc
    WHERE pronamespace = 'public'::regnamespace
      AND proname IN (
        'idempinsertnode','insertnode','idempappendnode','idempinsertcontext',
        'ncc_match','getsingletonaslinkarray','getsingletonaslink','getneighboursbytype',
        'getfwdnodes','getfwdlinks','fwdconeasnodes','fwdconeaslinks','fwdpathsaslinks',
        'sumfwdpaths','allpathsaslinks','sumallpaths','empty_path','match_context',
        'match_arrows','arrowinlist','arrowincontextlist','uncmp','sumallncpaths',
        'allncpathsaslinks','constraintpathsaslinks','sumconstraintpaths',
        'getconstrainedfwdlinks','getncfwdlinks','getncclinks','getncneighboursbytype',
        'getappointments','deletechapter','lastsawsection','lastsawnptr','sst_unaccent'
      )
  LOOP
    EXECUTE 'DROP FUNCTION IF EXISTS ' || r.p || ' CASCADE';
  END LOOP;
END $$;
