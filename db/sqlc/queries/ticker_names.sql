-- name: CreateTicker :one
INSERT INTO ticker_names (symbol, company_name, exchange)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTickerBySymbol :one
SELECT *
FROM ticker_names
WHERE symbol = $1;

-- name: UpsertTicker :exec
INSERT INTO ticker_names (symbol, company_name, exchange)
VALUES ($1, $2, $3)
ON CONFLICT (symbol) DO NOTHING;

-- name: ListAllTickers :many
SELECT * FROM ticker_names ORDER BY symbol;
