-- name: CreateVisitor :exec
INSERT INTO visitors (ip_address, endpoint, visited_at)
VALUES ($1, $2, $3);

-- name: GetVisitorsByIP :many
SELECT id, ip_address, endpoint, visited_at
FROM visitors
WHERE ip_address = $1
ORDER BY visited_at DESC;

-- name: GetVisitorsByEndpoint :many
SELECT id, ip_address, endpoint, visited_at
FROM visitors
WHERE endpoint = $1
ORDER BY visited_at DESC;

-- name: GetAllVisitors :many
SELECT id, ip_address, endpoint, visited_at
FROM visitors
ORDER BY visited_at DESC
LIMIT $1;

-- name: GetVisitorsLastDay :many
SELECT id, ip_address, endpoint, visited_at
FROM visitors
WHERE visited_at >= NOW() - INTERVAL '1 day'
ORDER BY visited_at DESC;

-- name: GetVisitorsLastWeek :many
SELECT id, ip_address, endpoint, visited_at
FROM visitors
WHERE visited_at >= NOW() - INTERVAL '1 week'
ORDER BY visited_at DESC;

-- name: GetVisitorsLastMonth :many
SELECT id, ip_address, endpoint, visited_at
FROM visitors
WHERE visited_at >= NOW() - INTERVAL '1 month'
ORDER BY visited_at DESC;

-- name: GetVisitorCountLastDay :one
SELECT COUNT(*) as count
FROM visitors
WHERE visited_at >= NOW() - INTERVAL '1 day';

-- name: GetVisitorCountLastWeek :one
SELECT COUNT(*) as count
FROM visitors
WHERE visited_at >= NOW() - INTERVAL '1 week';

-- name: GetVisitorCountLastMonth :one
SELECT COUNT(*) as count
FROM visitors
WHERE visited_at >= NOW() - INTERVAL '1 month';

-- name: GetVisitorCountAll :one
SELECT COUNT(*) as count
FROM visitors;