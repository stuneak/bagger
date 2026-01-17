-- name: InsertTickerPrice :exec
INSERT INTO ticker_prices (ticker_id, price, recorded_at)
VALUES ($1, $2, $3)
ON CONFLICT (ticker_id, recorded_at) DO NOTHING;

-- name: GetLatestTickerPrice :one
SELECT price, recorded_at
FROM ticker_prices
WHERE ticker_id = $1
ORDER BY recorded_at DESC
LIMIT 1;
