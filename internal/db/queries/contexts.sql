-- name: ListContexts :many
SELECT context, ctxptr
FROM contextdirectory
ORDER BY ctxptr;

-- name: GetContextByPtr :one
SELECT context, ctxptr
FROM contextdirectory
WHERE ctxptr = $1;

-- name: GetContextByText :one
SELECT context, ctxptr
FROM contextdirectory
WHERE context = $1;

-- name: GetContextByTextUnaccent :one
SELECT context, ctxptr
FROM contextdirectory
WHERE unaccent(context) = unaccent($1);

-- name: IdempInsertContext :one
SELECT ideminsertcontext(sqlc.arg(constr)::text, sqlc.arg(conptr)::int)::int AS ctxptr;

