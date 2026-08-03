-- name: ListContexts :many
SELECT ctx_ptr, context
FROM context_directory
ORDER BY ctx_ptr;

-- name: GetContextByPtr :one
SELECT ctx_ptr, context
FROM context_directory
WHERE ctx_ptr = $1;

-- name: GetContextByText :one
SELECT ctx_ptr, context
FROM context_directory
WHERE context = $1;

-- name: InsertContext :exec
INSERT INTO context_directory (ctx_ptr, context)
VALUES ($1, $2)
ON CONFLICT (ctx_ptr) DO NOTHING;

-- name: MaxCtxPtr :one
SELECT COALESCE(max(ctx_ptr), -1)::int AS max_ptr
FROM context_directory;
