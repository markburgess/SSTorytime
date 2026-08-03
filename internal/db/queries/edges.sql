-- name: InsertEdge :exec
INSERT INTO edge (src_class, src_cptr, dst_class, dst_cptr, arr, wgt, ctx, st)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT DO NOTHING;

-- name: ListEdgesFrom :many
SELECT src_class, src_cptr, dst_class, dst_cptr, arr, wgt, ctx, st
FROM edge
WHERE src_class = $1 AND src_cptr = $2;

-- name: ListEdgesFromByST :many
SELECT src_class, src_cptr, dst_class, dst_cptr, arr, wgt, ctx, st
FROM edge
WHERE src_class = $1 AND src_cptr = $2 AND st = $3
LIMIT $4;

-- name: ListEdgesTo :many
SELECT src_class, src_cptr, dst_class, dst_cptr, arr, wgt, ctx, st
FROM edge
WHERE dst_class = $1 AND dst_cptr = $2;

-- name: DeleteEdgesForNode :exec
DELETE FROM edge
WHERE (src_class = $1 AND src_cptr = $2)
   OR (dst_class = $1 AND dst_cptr = $2);

-- name: GetNeighboursByType :many
SELECT e.arr, e.wgt, e.ctx, e.dst_class, e.dst_cptr
FROM edge e
WHERE e.src_class = $1
  AND e.src_cptr = $2
  AND e.st = st_type_from_int($3)
  AND e.arr <> 0
LIMIT $4;
