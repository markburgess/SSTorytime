-- name: ListLastSeen :many
SELECT section, class, cptr,
       EXTRACT(EPOCH FROM first)::float8 AS first_epoch,
       EXTRACT(EPOCH FROM last)::float8 AS last_epoch,
       delta, freq,
       EXTRACT(EPOCH FROM (now() - last))::float8 AS ndelta
FROM last_seen
ORDER BY section;

-- name: GetLastSeenByNPtr :one
SELECT section, class, cptr,
       EXTRACT(EPOCH FROM first)::float8 AS first_epoch,
       EXTRACT(EPOCH FROM last)::float8 AS last_epoch,
       delta, freq,
       EXTRACT(EPOCH FROM (now() - last))::float8 AS ndelta
FROM last_seen
WHERE class = $1 AND cptr = $2;

-- name: ListRecentNPtrs :many
SELECT class, cptr
FROM last_seen
WHERE last > now() - make_interval(hours => $1);

-- name: InsertLastSeenSection :exec
INSERT INTO last_seen (section, class, cptr, first, last, delta, freq)
VALUES ($1, -1, -1, now(), now(), 0, 1);
