-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE user_id = $1;

-- name: GetUserFromRefreshToken :one
SELECT users.*
FROM refresh_tokens
INNER JOIN users ON refresh_tokens.user_id = users.id
WHERE refresh_tokens.token = $1
  AND refresh_tokens.revoked_at IS NULL
  AND refresh_tokens.expires_at > NOW();

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET token = NOW()
WHERE token = $1
RETURNING *;