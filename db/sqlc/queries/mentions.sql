-- name: CreateTickerMention :one
INSERT INTO ticker_mentions (
  ticker_id,
  user_id,
  comment_id,
  mentioned_at,
  price_id
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFirstMentionPerTickerByUsername :many
WITH first_mentions AS (
  SELECT DISTINCT ON (tm.ticker_id)
    t.symbol,
    t.id AS ticker_id,
    tp.price AS mention_price,
    tm.mentioned_at
  FROM ticker_mentions tm
  JOIN users u ON tm.user_id = u.id
  JOIN tickers t ON tm.ticker_id = t.id
  JOIN ticker_prices tp ON tm.price_id = tp.id
  WHERE u.username = $1
  ORDER BY tm.ticker_id, tm.mentioned_at ASC
)
SELECT
  fm.symbol,
  fm.ticker_id,
  fm.mention_price,
  fm.mentioned_at,
  lp.price AS current_price,
  lp.recorded_at AS current_price_date,
  COALESCE(ROUND(exp(sum(ln(NULLIF(ss.ratio::double precision, 0))))::numeric, 4), 1.0)::double precision AS split_ratio
FROM first_mentions fm
CROSS JOIN LATERAL (
  SELECT price, recorded_at
  FROM ticker_prices
  WHERE ticker_id = fm.ticker_id
  ORDER BY recorded_at DESC
  LIMIT 1
) lp
LEFT JOIN stock_splits ss ON ss.ticker_id = fm.ticker_id
  AND ss.effective_date > fm.mentioned_at::date
  AND ss.effective_date <= $2
GROUP BY fm.symbol, fm.ticker_id, fm.mention_price, fm.mentioned_at, lp.price, lp.recorded_at;

-- name: GetAllPicksWithPricesAndSplitsSince :many
WITH picks_with_prices AS (
  SELECT
    u.username,
    t.symbol,
    t.id AS ticker_id,
    tp.price AS mention_price,
    tm.mentioned_at,
    lp.price AS current_price,
    lp.recorded_at AS current_price_date
  FROM ticker_mentions tm
  JOIN users u ON tm.user_id = u.id
  JOIN tickers t ON tm.ticker_id = t.id
  JOIN ticker_prices tp ON tm.price_id = tp.id
  CROSS JOIN LATERAL (
    SELECT price, recorded_at
    FROM ticker_prices
    WHERE ticker_id = t.id
    ORDER BY recorded_at DESC
    LIMIT 1
  ) lp
  WHERE tm.mentioned_at >= $1
),
picks_with_splits AS (
  SELECT
    pp.username,
    pp.symbol,
    pp.ticker_id,
    pp.mention_price,
    pp.mentioned_at,
    pp.current_price,
    pp.current_price_date,
    COALESCE(ROUND(exp(sum(ln(NULLIF(ss.ratio::double precision, 0))))::numeric, 4), 1.0)::double precision AS split_ratio
  FROM picks_with_prices pp
  LEFT JOIN stock_splits ss ON ss.ticker_id = pp.ticker_id
    AND ss.effective_date >= pp.mentioned_at::date
    AND ss.effective_date <= pp.current_price_date::date
  GROUP BY pp.username, pp.symbol, pp.ticker_id, pp.mention_price, pp.mentioned_at, pp.current_price, pp.current_price_date
)
SELECT
  username,
  symbol,
  ticker_id,
  mention_price,
  mentioned_at,
  current_price,
  current_price_date,
  (CASE
    WHEN mention_price * split_ratio BETWEEN current_price * 0.8 AND current_price * 1.2
    THEN split_ratio
    WHEN mention_price BETWEEN current_price * 0.8 AND current_price * 1.2
    THEN 1.0
    ELSE split_ratio
  END)::double precision AS calculated_split_ratio
FROM picks_with_splits;

-- name: GetUniqueUserPicksWithLatestPricesSince :many
WITH first_picks AS (
  SELECT DISTINCT ON (tm.user_id, tm.ticker_id)
    u.username,
    t.symbol,
    tm.ticker_id,
    tp.price AS mention_price,
    tm.mentioned_at
  FROM ticker_mentions tm
  JOIN users u ON tm.user_id = u.id
  JOIN tickers t ON tm.ticker_id = t.id
  JOIN ticker_prices tp ON tm.price_id = tp.id
  WHERE tm.mentioned_at >= $1
  ORDER BY tm.user_id, tm.ticker_id, tm.mentioned_at ASC
),
picks_with_prices AS (
  SELECT
    fp.username,
    fp.symbol,
    fp.ticker_id,
    fp.mention_price,
    fp.mentioned_at,
    lp.price AS current_price,
    lp.recorded_at AS current_price_date
  FROM first_picks fp
  CROSS JOIN LATERAL (
    SELECT price, recorded_at
    FROM ticker_prices
    WHERE ticker_id = fp.ticker_id
    ORDER BY recorded_at DESC
    LIMIT 1
  ) lp
),
picks_with_splits AS (
  SELECT
    pp.username,
    pp.symbol,
    pp.ticker_id,
    pp.mention_price,
    pp.mentioned_at,
    pp.current_price,
    pp.current_price_date,
    COALESCE(ROUND(exp(sum(ln(ss.ratio::double precision)))::numeric, 4), 1.0)::double precision AS split_ratio
  FROM picks_with_prices pp
  LEFT JOIN stock_splits ss ON ss.ticker_id = pp.ticker_id
    AND ss.effective_date >= pp.mentioned_at::date
    AND ss.effective_date <= pp.current_price_date::date
  GROUP BY pp.username, pp.symbol, pp.ticker_id, pp.mention_price, pp.mentioned_at, pp.current_price, pp.current_price_date
)
SELECT
  username,
  symbol,
  ticker_id,
  mention_price,
  mentioned_at,
  current_price,
  current_price_date,
  (CASE
    WHEN mention_price * split_ratio BETWEEN current_price * 0.8 AND current_price * 1.2
    THEN split_ratio
    WHEN mention_price BETWEEN current_price * 0.8 AND current_price * 1.2
    THEN 1.0
    ELSE split_ratio
  END)::double precision AS calculated_split_ratio
FROM picks_with_splits;
