-- name: ListChaptersLike :many
SELECT DISTINCT chap
FROM node
WHERE lower(chap) LIKE lower($1)
ORDER BY chap;

-- name: ListChaptersLikeUnaccent :many
SELECT DISTINCT chap
FROM node
WHERE lower(unaccent(chap)) LIKE lower($1)
ORDER BY chap;

-- name: MaxCPtrForChan :one
SELECT COALESCE(max((nptr).cptr), -1)::int AS max_cptr
FROM node
WHERE (nptr).chan = $1::int;

-- name: TruncateAllData :exec
TRUNCATE TABLE
  node, pagemap, bookmarks, lastseen,
  arrowdirectory, arrowinverses, contextdirectory
CASCADE;

-- name: GetNodeByNPtr :one
SELECT l,
       s,
       chap,
       im3::text AS im3,
       im2::text AS im2,
       im1::text AS im1,
       in0::text AS in0,
       il1::text AS il1,
       ic2::text AS ic2,
       ie3::text AS ie3
FROM node
WHERE nptr = ROW($1::int, $2::int)::nodeptr
  AND NOT l = 0;

-- name: SearchNodePtrs :many
-- $1 any_chapter, $2 chap_unaccent, $3 chap_pattern
-- $4 name_mode: any|exact|like|fts|ufts, $5 name_arg
-- $6 exclude_paths, $7 seq_only
-- $8 context text[], $9 arrows int[], $10 sttypes int[], $11 limit
SELECT (nptr).chan::int AS chan, (nptr).cptr::int AS cptr
FROM node
WHERE
  CASE
    WHEN $1::boolean THEN true
    WHEN $2::boolean THEN lower(unaccent(chap)) LIKE lower($3)
    ELSE lower(chap) LIKE lower($3)
  END
  AND CASE $4::text
    WHEN 'any' THEN true
    WHEN 'exact' THEN lower(s) = lower($5)
    WHEN 'like' THEN lower(s) LIKE lower($5)
    WHEN 'fts' THEN search @@ to_tsquery('english', $5)
    WHEN 'ufts' THEN unsearch @@ to_tsquery('english', $5)
    ELSE true
  END
  AND (NOT $6::boolean OR s NOT LIKE '/%')
  AND (NOT $7::boolean OR seq = true)
  AND ncc_match(
    nptr,
    COALESCE($8::text[], ARRAY[]::text[]),
    COALESCE($9::int[], ARRAY[]::int[]),
    COALESCE($10::int[], ARRAY[]::int[]),
    im3, im2, im1, in0, il1, ic2, ie3
  )
ORDER BY s ASC, (cardinality(ie3) + cardinality(im3) + cardinality(il1)) DESC
LIMIT $11;

-- name: InsertNodeFn :exec
SELECT insertnode($1::int, $2::int, $3::int, $4, $5, $6::boolean);

-- name: IdempAppendNode :one
-- Body fills ret_cptr/ret_channel from (Chan, CPtr) respectively.
SELECT r.ret_cptr::int AS chan, r.ret_channel::int AS cptr
FROM idempappendnode($1::int, $2::int, $3, $4) AS r;

-- name: InsertNodeRow :exec
INSERT INTO node (nptr, l, s, chap, seq, im3, im2, im1, in0, il1, ic2, ie3)
VALUES (
  ROW($1::int, $2::int)::nodeptr,
  $3::int,
  $4,
  $5,
  $6::boolean,
  COALESCE($7::text::link[], '{}'::link[]),
  COALESCE($8::text::link[], '{}'::link[]),
  COALESCE($9::text::link[], '{}'::link[]),
  COALESCE($10::text::link[], '{}'::link[]),
  COALESCE($11::text::link[], '{}'::link[]),
  COALESCE($12::text::link[], '{}'::link[]),
  COALESCE($13::text::link[], '{}'::link[])
);

-- Append link into one ST channel (dynamic column avoided via separate queries).

-- name: AppendLinkIm3 :exec
UPDATE node
SET im3 = array_append(im3, ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link)
WHERE (nptr).cptr = $1::int
  AND (nptr).chan = $2::int
  AND (im3 IS NULL OR NOT ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link = ANY (im3));

-- name: AppendLinkIm2 :exec
UPDATE node
SET im2 = array_append(im2, ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link)
WHERE (nptr).cptr = $1::int
  AND (nptr).chan = $2::int
  AND (im2 IS NULL OR NOT ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link = ANY (im2));

-- name: AppendLinkIm1 :exec
UPDATE node
SET im1 = array_append(im1, ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link)
WHERE (nptr).cptr = $1::int
  AND (nptr).chan = $2::int
  AND (im1 IS NULL OR NOT ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link = ANY (im1));

-- name: AppendLinkIn0 :exec
UPDATE node
SET in0 = array_append(in0, ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link)
WHERE (nptr).cptr = $1::int
  AND (nptr).chan = $2::int
  AND (in0 IS NULL OR NOT ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link = ANY (in0));

-- name: AppendLinkIl1 :exec
UPDATE node
SET il1 = array_append(il1, ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link)
WHERE (nptr).cptr = $1::int
  AND (nptr).chan = $2::int
  AND (il1 IS NULL OR NOT ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link = ANY (il1));

-- name: AppendLinkIc2 :exec
UPDATE node
SET ic2 = array_append(ic2, ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link)
WHERE (nptr).cptr = $1::int
  AND (nptr).chan = $2::int
  AND (ic2 IS NULL OR NOT ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link = ANY (ic2));

-- name: AppendLinkIe3 :exec
UPDATE node
SET ie3 = array_append(ie3, ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link)
WHERE (nptr).cptr = $1::int
  AND (nptr).chan = $2::int
  AND (ie3 IS NULL OR NOT ROW($3::int, $4::real, $5::int, ROW($6::int, $7::int)::nodeptr)::link = ANY (ie3));

-- name: SetLinkArrayIm3 :exec
UPDATE node SET im3 = $3::text::link[] WHERE (nptr).cptr = $1::int AND (nptr).chan = $2::int;

-- name: SetLinkArrayIm2 :exec
UPDATE node SET im2 = $3::text::link[] WHERE (nptr).cptr = $1::int AND (nptr).chan = $2::int;

-- name: SetLinkArrayIm1 :exec
UPDATE node SET im1 = $3::text::link[] WHERE (nptr).cptr = $1::int AND (nptr).chan = $2::int;

-- name: SetLinkArrayIn0 :exec
UPDATE node SET in0 = $3::text::link[] WHERE (nptr).cptr = $1::int AND (nptr).chan = $2::int;

-- name: SetLinkArrayIl1 :exec
UPDATE node SET il1 = $3::text::link[] WHERE (nptr).cptr = $1::int AND (nptr).chan = $2::int;

-- name: SetLinkArrayIc2 :exec
UPDATE node SET ic2 = $3::text::link[] WHERE (nptr).cptr = $1::int AND (nptr).chan = $2::int;

-- name: SetLinkArrayIe3 :exec
UPDATE node SET ie3 = $3::text::link[] WHERE (nptr).cptr = $1::int AND (nptr).chan = $2::int;

-- Singleton sources/sinks: flags enable each ST channel (positive ST = out).

-- name: ListSingletonSources :many
SELECT (nptr).chan::int AS chan, (nptr).cptr::int AS cptr
FROM node
WHERE lower(chap) LIKE lower($1)
  AND (
    ($2::boolean AND array_length(il1::text[], 1) IS NOT NULL AND array_length(im1::text[], 1) IS NULL
      AND match_context((il1)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
    OR ($3::boolean AND array_length(ic2::text[], 1) IS NOT NULL AND array_length(im2::text[], 1) IS NULL
      AND match_context((ic2)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
    OR ($4::boolean AND array_length(ie3::text[], 1) IS NOT NULL AND array_length(im3::text[], 1) IS NULL
      AND match_context((ie3)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
    OR ($5::boolean AND array_length(in0::text[], 1) IS NOT NULL
      AND match_context((in0)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
  );

-- name: ListSingletonSinks :many
SELECT (nptr).chan::int AS chan, (nptr).cptr::int AS cptr
FROM node
WHERE lower(chap) LIKE lower($1)
  AND (
    ($2::boolean AND array_length(im1::text[], 1) IS NOT NULL AND array_length(il1::text[], 1) IS NULL
      AND match_context((im1)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
    OR ($3::boolean AND array_length(im2::text[], 1) IS NOT NULL AND array_length(ic2::text[], 1) IS NULL
      AND match_context((im2)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
    OR ($4::boolean AND array_length(im3::text[], 1) IS NOT NULL AND array_length(ie3::text[], 1) IS NULL
      AND match_context((im3)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
    OR ($5::boolean AND array_length(in0::text[], 1) IS NOT NULL
      AND match_context((in0)[0].ctx, COALESCE($6::text[], ARRAY[]::text[])))
  );

-- Adjacency: always return all link channels as text; Go picks which ST types matter.

-- name: ListAdjacentNodes :many
SELECT (nptr).chan::int AS chan,
       (nptr).cptr::int AS cptr,
       im3::text AS im3,
       im2::text AS im2,
       im1::text AS im1,
       in0::text AS in0,
       il1::text AS il1,
       ic2::text AS ic2,
       ie3::text AS ie3
FROM node
WHERE lower(chap) LIKE lower($1)
  AND (
    ($2::boolean AND array_length(im3::text[], 1) IS NOT NULL
      AND match_context((im3)[0].ctx, COALESCE($9::text[], ARRAY[]::text[])))
    OR ($3::boolean AND array_length(im2::text[], 1) IS NOT NULL
      AND match_context((im2)[0].ctx, COALESCE($9::text[], ARRAY[]::text[])))
    OR ($4::boolean AND array_length(im1::text[], 1) IS NOT NULL
      AND match_context((im1)[0].ctx, COALESCE($9::text[], ARRAY[]::text[])))
    OR ($5::boolean AND array_length(in0::text[], 1) IS NOT NULL
      AND match_context((in0)[0].ctx, COALESCE($9::text[], ARRAY[]::text[])))
    OR ($6::boolean AND array_length(il1::text[], 1) IS NOT NULL
      AND match_context((il1)[0].ctx, COALESCE($9::text[], ARRAY[]::text[])))
    OR ($7::boolean AND array_length(ic2::text[], 1) IS NOT NULL
      AND match_context((ic2)[0].ctx, COALESCE($9::text[], ARRAY[]::text[])))
    OR ($8::boolean AND array_length(ie3::text[], 1) IS NOT NULL
      AND match_context((ie3)[0].ctx, COALESCE($9::text[], ARRAY[]::text[])))
  );

-- name: CreateNodeIndexes :exec
CREATE INDEX IF NOT EXISTS sst_gin ON node USING gin (to_tsvector('english', search));

-- name: CreateNodeIndexes2 :exec
CREATE INDEX IF NOT EXISTS sst_ungin ON node USING gin (to_tsvector('english', unsearch));

-- name: CreateNodeIndexes3 :exec
CREATE INDEX IF NOT EXISTS sst_s ON node USING gin (s);

-- name: CreateNodeIndexes4 :exec
CREATE INDEX IF NOT EXISTS sst_n ON node USING gin (nptr);

-- name: CreateContextIndex :exec
CREATE INDEX IF NOT EXISTS sst_cnt ON contextdirectory USING gin (context);

-- name: AlterNodeLogged :exec
ALTER TABLE node SET LOGGED;

-- name: AlterPageMapLogged :exec
ALTER TABLE pagemap SET LOGGED;
