-- Original SSTorytime schema (exact upstream shapes & names).
-- PL/pgSQL graph functions are installed by DefineStoredFunctions on Open
-- (CREATE OR REPLACE), same as upstream — keeps function bodies bit-identical.

CREATE EXTENSION IF NOT EXISTS unaccent;

DO $$ BEGIN
  CREATE TYPE NodePtr AS (Chan int, CPtr int);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE Link AS (Arr int, Wgt real, Ctx int, Dst NodePtr);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
  CREATE TYPE Appointment AS (
    Arr int, STType int, Chap text, Ctx int, NTo NodePtr, NFrom NodePtr[]
  );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE OR REPLACE FUNCTION sst_unaccent(this text)
RETURNS text
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT unaccent(this);
$$;

CREATE TABLE IF NOT EXISTS ContextDirectory (
    Context text,
    CtxPtr  int primary key
);

CREATE TABLE IF NOT EXISTS Bookmarks (
    Query text,
    Bookmark text
);

CREATE UNLOGGED TABLE IF NOT EXISTS PageMap (
    Chap  text,
    Alias text,
    Ctx   int,
    Line  int,
    Path  Link[]
);

CREATE UNLOGGED TABLE IF NOT EXISTS Node (
    NPtr     NodePtr,
    L        int,
    S        text,
    Search   TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', S)) STORED,
    UnSearch TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', sst_unaccent(S))) STORED,
    Chap     text,
    Seq      boolean,
    Im3      Link[],
    Im2      Link[],
    Im1      Link[],
    In0      Link[],
    Il1      Link[],
    Ic2      Link[],
    Ie3      Link[]
);

CREATE UNLOGGED TABLE IF NOT EXISTS ArrowInverses (
    Plus  int,
    Minus int,
    PRIMARY KEY (Plus, Minus)
);

CREATE UNLOGGED TABLE IF NOT EXISTS ArrowDirectory (
    STAindex int,
    Long     text,
    Short    text,
    ArrPtr   int primary key
);

CREATE TABLE IF NOT EXISTS LastSeen (
    Section text,
    NPtr    NodePtr,
    First   timestamp,
    Last    timestamp,
    Delta   real,
    Freq    int
);
