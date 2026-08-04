-- name: ListBookmarks :many
SELECT bookmark, query
FROM bookmarks;

-- name: InsertBookmark :exec
INSERT INTO bookmarks (bookmark, query)
VALUES ($1, $2);
