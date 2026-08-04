-- name: ListArrowDirectory :many
SELECT staindex, long, short, arrptr
FROM arrowdirectory
ORDER BY arrptr;

-- name: ListArrowInverses :many
SELECT plus, minus
FROM arrowinverses
ORDER BY plus;

-- name: InsertArrowDirectory :exec
INSERT INTO arrowdirectory (staindex, long, short, arrptr)
SELECT $1, $2, $3, $4
WHERE NOT EXISTS (
  SELECT 1 FROM arrowdirectory
  WHERE lower(long) = lower($2) OR lower(short) = lower($3) OR arrptr = $4
);

-- name: InsertArrowInverse :exec
INSERT INTO arrowinverses (plus, minus)
SELECT $1, $2
WHERE NOT EXISTS (
  SELECT 1 FROM arrowinverses WHERE plus = $1 OR minus = $2
);

-- name: TruncateArrowDirectory :exec
TRUNCATE TABLE arrowdirectory;

-- name: TruncateArrowInverses :exec
TRUNCATE TABLE arrowinverses;
