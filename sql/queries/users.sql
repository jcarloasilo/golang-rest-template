-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (email, name, hashed_password) VALUES ($1, $2, $3) RETURNING *;

-- name: VerifyUser :exec
UPDATE users
SET verified_at = sqlc.arg(verified_at)::TIMESTAMPTZ
WHERE id = sqlc.arg(user_id)::UUID;