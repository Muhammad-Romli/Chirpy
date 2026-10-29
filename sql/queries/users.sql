-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, created_at, updated_at, email;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET 
    email = sqlc.arg(email),
    hashed_password = sqlc.arg(hashed_password),
    updated_at = sqlc.arg(updated_time) -- Using parameter instead NOW() to synchronize
WHERE id = sqlc.arg(id)
RETURNING *;


-- <------------------------------ PREMIUM ------------------------------------------->

-- name: PremiumPurchase :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1
RETURNING *;