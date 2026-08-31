-- name: CreateUser :one
INSERT INTO users (user_id, display_name, created_at, is_active)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: CreateSession :one
INSERT INTO sessions (user_id, session_token, created_at, expires_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetActiveUserBySessionToken :one
SELECT users.*
FROM sessions JOIN users ON sessions.user_id = users.id WHERE session_token = ? AND expires_at > ? AND users.is_active = 1;

-- name: UpdateDisplayNameByUserID :one
UPDATE users set display_name = ? WHERE user_id = ? AND is_active = 1 RETURNING *;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions WHERE session_token = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < ?;

-- name: GetUserByUserID :one
SELECT * FROM users WHERE user_id = ?;

-- name: GetActiveUserByUserID :one
SELECT * FROM users WHERE user_id = ? AND is_active = 1;
