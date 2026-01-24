-- name: GetSplitsBetweenDates :many
SELECT ratio, effective_date
FROM ticker_splits
WHERE ticker_id = $1
  AND effective_date >= $2::date
  AND effective_date <= $3
ORDER BY effective_date;

-- name: InsertTickerSplit :exec
INSERT INTO ticker_splits (ticker_id, ratio, effective_date)
VALUES ($1, $2, $3)
ON CONFLICT (ticker_id, effective_date) DO NOTHING;

-- name: GetSplitsByTicker :many
SELECT id, ticker_id, ratio, effective_date
FROM ticker_splits
WHERE ticker_id = $1
ORDER BY effective_date DESC;

-- name: GetAllSplits :many
SELECT ticker_id, ratio, effective_date
FROM ticker_splits
ORDER BY ticker_id, effective_date;
