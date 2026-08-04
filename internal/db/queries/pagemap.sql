-- name: ListPageMap :many
SELECT chap,
       ctx,
       line,
       path::text AS path
FROM pagemap
WHERE match_context(ctx, COALESCE($1::text[], ARRAY[]::text[])) = true
  AND lower(chap) LIKE lower($2)
ORDER BY chap, line
OFFSET $3
LIMIT $4;

-- name: ListPageMapChapters :many
SELECT DISTINCT chap, ctx
FROM pagemap
WHERE match_context(ctx, COALESCE($1::text[], ARRAY[]::text[]))
  AND (
    $2::boolean
    OR (
      CASE
        WHEN $3::boolean THEN lower(unaccent(chap)) LIKE lower($4)
        ELSE lower(chap) LIKE lower($4)
      END
    )
  )
ORDER BY chap;

-- name: InsertPageMap :exec
INSERT INTO pagemap (chap, alias, ctx, line, path)
VALUES ($1, $2, $3::int, $4::int, $5::text::link[]);
