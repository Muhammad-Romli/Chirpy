-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(refresh_token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE user_id = $1;