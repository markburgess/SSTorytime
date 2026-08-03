-- SSTorytime initial schema (flattened for sqlc/pgx)
-- ST link types are a Postgres enum (value), not seven column names.

CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TYPE st_type AS ENUM (
    'm_express',  -- -3
    'm_contains', -- -2
    'm_leads',    -- -1
    'near',       --  0
    'p_leads',    -- +1
    'p_contains', -- +2
    'p_express'   -- +3
);

-- Map Go ST ints (-3..+3) to enum labels
CREATE OR REPLACE FUNCTION st_type_from_int(i int)
RETURNS st_type
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT CASE i
    WHEN -3 THEN 'm_express'::st_type
    WHEN -2 THEN 'm_contains'::st_type
    WHEN -1 THEN 'm_leads'::st_type
    WHEN  0 THEN 'near'::st_type
    WHEN  1 THEN 'p_leads'::st_type
    WHEN  2 THEN 'p_contains'::st_type
    WHEN  3 THEN 'p_express'::st_type
    ELSE NULL
  END;
$$;

CREATE OR REPLACE FUNCTION st_type_to_int(t st_type)
RETURNS int
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT CASE t
    WHEN 'm_express'  THEN -3
    WHEN 'm_contains' THEN -2
    WHEN 'm_leads'    THEN -1
    WHEN 'near'       THEN  0
    WHEN 'p_leads'    THEN  1
    WHEN 'p_contains' THEN  2
    WHEN 'p_express'  THEN  3
  END;
$$;

CREATE OR REPLACE FUNCTION sst_unaccent(this text)
RETURNS text
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT unaccent(this);
$$;

CREATE TABLE node (
    class   int  NOT NULL,
    cptr    int  NOT NULL,
    l       int  NOT NULL DEFAULT 0,
    s       text NOT NULL,
    search  tsvector GENERATED ALWAYS AS (to_tsvector('english', s)) STORED,
    unsearch tsvector GENERATED ALWAYS AS (to_tsvector('english', sst_unaccent(s))) STORED,
    chap    text NOT NULL DEFAULT '',
    seq     boolean NOT NULL DEFAULT false,
    PRIMARY KEY (class, cptr)
);

CREATE INDEX node_search_gin ON node USING GIN (search);
CREATE INDEX node_unsearch_gin ON node USING GIN (unsearch);
CREATE INDEX node_chap_lower ON node (lower(chap));

CREATE TABLE edge (
    src_class int NOT NULL,
    src_cptr  int NOT NULL,
    dst_class int NOT NULL,
    dst_cptr  int NOT NULL,
    arr       int NOT NULL DEFAULT 0,
    wgt       real NOT NULL DEFAULT 1.0,
    ctx       int NOT NULL DEFAULT 0,
    st        st_type NOT NULL,
    PRIMARY KEY (src_class, src_cptr, dst_class, dst_cptr, arr, ctx, st),
    FOREIGN KEY (src_class, src_cptr) REFERENCES node (class, cptr) ON DELETE CASCADE,
    FOREIGN KEY (dst_class, dst_cptr) REFERENCES node (class, cptr) ON DELETE CASCADE
);

CREATE INDEX edge_src ON edge (src_class, src_cptr, st);
CREATE INDEX edge_dst ON edge (dst_class, dst_cptr, st);
CREATE INDEX edge_st ON edge (st);

CREATE TABLE arrow_directory (
    arr_ptr   int PRIMARY KEY,
    sta_index int NOT NULL DEFAULT 0,
    long_name text NOT NULL,
    short_name text NOT NULL
);

CREATE UNIQUE INDEX arrow_long_lower ON arrow_directory (lower(long_name));
CREATE UNIQUE INDEX arrow_short_lower ON arrow_directory (lower(short_name));

CREATE TABLE arrow_inverses (
    plus  int NOT NULL,
    minus int NOT NULL,
    PRIMARY KEY (plus, minus)
);

CREATE TABLE context_directory (
    ctx_ptr int PRIMARY KEY,
    context text NOT NULL
);

CREATE UNIQUE INDEX context_text ON context_directory (context);

CREATE TABLE page_map (
    chap  text NOT NULL,
    alias text NOT NULL DEFAULT '',
    ctx   int  NOT NULL DEFAULT 0,
    line  int  NOT NULL DEFAULT 0,
    -- path stored as JSON array of {arr,wgt,ctx,dst_class,dst_cptr} for sqlc friendliness
    path  jsonb NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX page_map_chap ON page_map (chap);

CREATE TABLE last_seen (
    section text,
    class   int,
    cptr    int,
    first   timestamptz NOT NULL DEFAULT now(),
    last    timestamptz NOT NULL DEFAULT now(),
    delta   real NOT NULL DEFAULT 0,
    freq    int  NOT NULL DEFAULT 0
);

CREATE INDEX last_seen_nptr ON last_seen (class, cptr);
CREATE INDEX last_seen_section ON last_seen (section);

CREATE TABLE bookmarks (
    bookmark text NOT NULL,
    query    text NOT NULL
);

-- Neighbours by ST type (replaces GetNeighboursByType over link arrays)
CREATE OR REPLACE FUNCTION get_neighbours_by_type(
    start_class int,
    start_cptr int,
    sttype int,
    maxlimit int
)
RETURNS TABLE (
    arr int,
    wgt real,
    ctx int,
    dst_class int,
    dst_cptr int
)
LANGUAGE sql
STABLE
AS $$
  SELECT e.arr, e.wgt, e.ctx, e.dst_class, e.dst_cptr
  FROM edge e
  WHERE e.src_class = start_class
    AND e.src_cptr = start_cptr
    AND e.st = st_type_from_int(sttype)
    AND e.arr <> 0
  LIMIT maxlimit;
$$;

-- Context match: empty user set matches all; otherwise context string must be in set
CREATE OR REPLACE FUNCTION match_context(thisctxptr int, user_set text[])
RETURNS boolean
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
  ctxstr text;
BEGIN
  IF user_set IS NULL OR array_length(user_set, 1) IS NULL THEN
    RETURN true;
  END IF;
  IF thisctxptr IS NULL OR thisctxptr < 0 THEN
    RETURN true;
  END IF;
  SELECT c.context INTO ctxstr FROM context_directory c WHERE c.ctx_ptr = thisctxptr;
  IF ctxstr IS NULL OR ctxstr = '' THEN
    RETURN true;
  END IF;
  RETURN ctxstr = ANY (user_set);
END;
$$;
