-- name: CallLastSawSection :exec
SELECT LastSawSection($1);

-- name: CallLastSawNPtr :exec
SELECT LastSawNPtr(ROW($1::int, $2::int)::nodeptr, $3);

-- name: ListLastSeen :many
SELECT section,
       (nptr).chan AS chan,
       (nptr).cptr AS cptr,
       EXTRACT(EPOCH FROM first)::float8 AS first_epoch,
       EXTRACT(EPOCH FROM last)::float8 AS last_epoch,
       delta,
       freq,
       EXTRACT(EPOCH FROM (NOW() - last))::float8 AS ndelta
FROM lastseen
ORDER BY section;

-- name: GetLastSeenByNPtr :one
SELECT section,
       EXTRACT(EPOCH FROM first)::float8 AS first_epoch,
       EXTRACT(EPOCH FROM last)::float8 AS last_epoch,
       delta,
       freq,
       EXTRACT(EPOCH FROM (NOW() - last))::float8 AS ndelta
FROM lastseen
WHERE nptr = ROW($1::int, $2::int)::nodeptr;

-- name: ListRecentLastSeenNPtrs :many
SELECT (nptr).chan AS chan, (nptr).cptr AS cptr
FROM lastseen
WHERE last > NOW() - make_interval(hours => $1);

-- name: ListAllLastSeenNPtrs :many
SELECT (nptr).chan AS chan, (nptr).cptr AS cptr
FROM lastseen;
