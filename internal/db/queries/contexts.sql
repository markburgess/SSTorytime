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
