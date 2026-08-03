-- name: InsertNode :exec
INSERT INTO node (class, cptr, l, s, chap, seq)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetNodeByPtr :one
SELECT class, cptr, l, s, chap, seq
FROM node
WHERE class = $1 AND cptr = $2;

-- name: GetNodeByText :one
SELECT class, cptr, l, s, chap, seq
FROM node
WHERE lower(s) = lower($1)
LIMIT 1;

-- name: MaxCPtrForClass :one
SELECT COALESCE(max(cptr), -1)::int AS max_cptr
FROM node
WHERE class = $1;

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

-- name: SearchNodesTsquery :many
SELECT class, cptr, l, s, chap, seq
FROM node
WHERE search @@ to_tsquery('english', $1)
ORDER BY s ASC
LIMIT $2;

-- name: SearchNodesTsqueryUnaccent :many
SELECT class, cptr, l, s, chap, seq
FROM node
WHERE unsearch @@ to_tsquery('english', $1)
ORDER BY s ASC
LIMIT $2;

-- name: SearchNodesExact :many
SELECT class, cptr, l, s, chap, seq
FROM node
WHERE lower(s) = lower($1)
ORDER BY s ASC
LIMIT $2;

-- name: SearchNodesLike :many
SELECT class, cptr, l, s, chap, seq
FROM node
WHERE lower(s) LIKE lower($1)
ORDER BY s ASC
LIMIT $2;

-- name: DeleteNodesByChapter :exec
DELETE FROM node
WHERE chap = $1;

-- name: UpdateNodeChapter :exec
UPDATE node
SET chap = $3
WHERE class = $1 AND cptr = $2;

-- name: UpdateNodeSeq :exec
UPDATE node
SET seq = $3
WHERE class = $1 AND cptr = $2;
