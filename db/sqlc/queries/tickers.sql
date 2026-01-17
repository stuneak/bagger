-- name: CreateTicker :one
INSERT INTO tickers (symbol, company_name, exchange)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTickerBySymbol :one
SELECT *
FROM tickers
WHERE symbol = $1;
