-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, author_id)
VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;


-- name: GetChirpsFilterAuthorId :many
SELECT * FROM chirps
WHERE author_id == sqlc.arg(author_id)
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirp :one
DELETE FROM chirps 
WHERE id = $1
RETURNING *;