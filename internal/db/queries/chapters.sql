-- name: ListChaptersLike :many
SELECT DISTINCT Chap FROM Node WHERE lower(Chap) LIKE lower($1) ORDER BY Chap;
