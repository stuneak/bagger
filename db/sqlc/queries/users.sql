-- name: CreateUser :one
INSERT INTO users (username)
VALUES ($1)
RETURNING id, username, created_at;

-- name: GetUserByUsername :one
SELECT id, username, created_at
FROM users
WHERE username = $1;
