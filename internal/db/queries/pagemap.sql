-- name: InsertPageMap :exec
INSERT INTO page_map (chap, alias, ctx, line, path)
VALUES ($1, $2, $3, $4, $5);

-- name: ListPageMapByContext :many
SELECT chap, alias, ctx, line, path
FROM page_map
WHERE match_context(ctx, $1::text[])
ORDER BY chap, line;

-- name: ListPageMapDistinctChapCtx :many
SELECT DISTINCT chap, ctx
FROM page_map
WHERE match_context(ctx, $1::text[])
ORDER BY chap;
