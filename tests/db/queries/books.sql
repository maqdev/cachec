-- name: GetBook :one
SELECT *
FROM books
WHERE id = $1 LIMIT 1;

-- name: GetAuthorBooks :many
SELECT *
FROM books
WHERE author_id = $1
ORDER BY title;

-- name: GetBooksByIDs :many
SELECT *
FROM books
WHERE id = ANY (@ids::bigint[])
ORDER BY title;

-- name: CreateBook :one
INSERT INTO books (author_id, title, publish_date)
VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateBook :exec
UPDATE books
set author_id    = $2,
    title        = $3,
    publish_date = $4
WHERE id = $1;

-- name: DeleteBook :exec
DELETE
FROM books
WHERE id = $1;
