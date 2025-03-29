-- name: StoreRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetUserIDByRefreshToken :one
SELECT user_id FROM refresh_tokens WHERE token = $1;

-- name: RevokeRefreshToken :exec
DELETE FROM refresh_tokens WHERE token = $1;