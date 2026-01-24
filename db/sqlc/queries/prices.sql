-- name: InsertTickerPrice :one
INSERT INTO ticker_prices (ticker_id, price, volume, recorded_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (ticker_id, recorded_at) DO UPDATE SET price = EXCLUDED.price, volume = EXCLUDED.volume
RETURNING *;

-- name: GetLatestTickerPrice :one
SELECT price, recorded_at
FROM ticker_prices
WHERE ticker_id = $1
ORDER BY recorded_at DESC
LIMIT 1;

-- name: GetTickerPriceByDate :one
SELECT id, price, recorded_at
FROM ticker_prices
WHERE ticker_id = $1 AND DATE(recorded_at) = DATE($2);

-- name: TickerPriceExistsForDate :one
SELECT EXISTS(
  SELECT 1 FROM ticker_prices
  WHERE ticker_id = $1 AND DATE(recorded_at) = DATE($2)
) AS exists;

-- name: GetHighestPriceAfterDate :one
SELECT price, recorded_at
FROM ticker_prices
WHERE ticker_id = $1 AND recorded_at >= $2
ORDER BY price DESC
LIMIT 1;

