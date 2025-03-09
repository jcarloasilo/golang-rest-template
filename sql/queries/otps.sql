-- name: CreateOTP :exec
INSERT INTO otps (code, type, user_id, expires_at, created_at)
VALUES ($1, $2, sqlc.arg(user_id)::UUID, $3, $4);

-- name: InvalidateExistingOTP :exec
DELETE FROM otps
WHERE user_id = sqlc.arg(user_id)::UUID AND type = $1;

-- name: GetLatestOTP :one
SELECT * FROM otps
WHERE user_id = sqlc.arg(user_id)::UUID AND type = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementOTPAttempts :exec
UPDATE otps
SET attempts = attempts + 1
WHERE id = sqlc.arg(id)::UUID;

-- name: ExpireOTP :exec
UPDATE otps
SET expires_at = CURRENT_TIMESTAMP - INTERVAL '1 second'
WHERE id = sqlc.arg(id)::UUID;

-- name: DeleteOTP :exec
DELETE FROM otps
WHERE id = sqlc.arg(id)::UUID;
