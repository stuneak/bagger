-- name: CreateComment :one
INSERT INTO comments (user_id, source, external_id, content, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
