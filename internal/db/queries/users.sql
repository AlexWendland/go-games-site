-- name: GetUserByUserID :one
SELECT * FROM users WHERE user_id = ?;

-- name: CreateUser :one
INSERT INTO users (user_id, display_name, created_at, is_active)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: SetUserActive :exec
UPDATE users SET is_active = ? WHERE id = ?;
