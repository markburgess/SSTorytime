-- Graph PL/pgSQL wrappers (functions from 000002_functions).
-- Composite / array results as text for existing Go parsers.
-- nodeptr[] / int[] inputs often passed as text casts so sqlc's analyzer stays happy.

-- name: FwdConeAsNodes :many
SELECT (u).chan::int AS chan, (u).cptr::int AS cptr
FROM unnest(fwdconeasnodes(ROW($1::int, $2::int)::nodeptr, $3::int, $4::int, $5::int)) AS u;

-- name: FwdConeAsLinks :many
SELECT x::text AS link
FROM unnest(fwdconeaslinks(ROW($1::int, $2::int)::nodeptr, $3::int, $4::int, $5::int)) AS x;

-- name: FwdPathsAsLinks :one
SELECT fwdpathsaslinks(ROW($1::int, $2::int)::nodeptr, $3::int, $4::int, $5::int)::text AS paths;

-- name: AllPathsAsLinks :one
SELECT allpathsaslinks(ROW($1::int, $2::int)::nodeptr, $3::text, $4::int, $5::int)::text AS paths;

-- name: AllNCPathsAsLinks :one
-- $1 = text literal of nodeptr[] e.g. {"(1,0)","(2,1)"}
SELECT allncpathsaslinks(
  $1::text::nodeptr[],
  $2::text,
  $3::boolean,
  COALESCE($4::text[], ARRAY[]::text[]),
  $5::text,
  $6::int,
  $7::int
)::text AS paths;

-- name: ConstraintPathsAsLinks :one
SELECT constraintpathsaslinks(
  $1::text::nodeptr[],
  $2::text,
  $3::boolean,
  COALESCE($4::text[], ARRAY[]::text[]),
  COALESCE($5::int[], ARRAY[]::int[]),
  COALESCE($6::int[], ARRAY[]::int[]),
  $7::int,
  $8::int
)::text AS paths;

-- name: GetConstrainedFwdLinks :one
SELECT getconstrainedfwdlinks(
  ROW($1::int, $2::int)::nodeptr,
  $3::text,
  $4::boolean,
  COALESCE($5::text[], ARRAY[]::text[]),
  $6::text::nodeptr[],
  $7::int,
  COALESCE($8::int[], ARRAY[]::int[]),
  $9::int
)::text AS links;

-- name: GetAppointments :many
SELECT x::text AS appointment
FROM unnest(getappointments(
  $1::int,
  $2::int,
  $3::int,
  $4::text,
  COALESCE($5::text[], ARRAY[]::text[]),
  $6::boolean
)) AS x;
