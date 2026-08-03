-- name: ListArrows :many
SELECT arr_ptr, sta_index, long_name, short_name
FROM arrow_directory
ORDER BY arr_ptr;

-- name: InsertArrow :exec
INSERT INTO arrow_directory (arr_ptr, sta_index, long_name, short_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (arr_ptr) DO NOTHING;

-- name: ListArrowInverses :many
SELECT plus, minus
FROM arrow_inverses
ORDER BY plus;

-- name: InsertArrowInverse :exec
INSERT INTO arrow_inverses (plus, minus)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
