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
WHERE (nptr).chan = $1;

-- name: TruncateAllData :exec
TRUNCATE TABLE
  node, pagemap, bookmarks, lastseen,
  arrowdirectory, arrowinverses, contextdirectory
CASCADE;
