-- name: CreateTickerMention :one
INSERT INTO ticker_mentions (
  ticker_id,
  user_id,
  comment_id,
  mentioned_at
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserMentionsComplete :many
SELECT
  tn.symbol,
  COALESCE(mention_price.price::text, '0') AS mention_price,
  COALESCE(current_price.price::text, '0') AS current_price,
  COALESCE(current_price.recorded_at, now()) AS current_price_date,
  tm.mentioned_at,
  COALESCE((
    SELECT EXP(SUM(LN(ts.ratio::double precision)))
    FROM ticker_splits ts
    WHERE ts.ticker_id = tm.ticker_id
      AND ts.effective_date >= tm.mentioned_at
      AND ts.effective_date <= COALESCE(current_price.recorded_at, now())
  ), 1.0)::double precision AS split_ratio
FROM (
  SELECT DISTINCT ON (ticker_id) ticker_id, mentioned_at
  FROM ticker_mentions
  WHERE user_id = (SELECT id FROM users WHERE username = $1)
    AND mentioned_at >= $2
  ORDER BY ticker_id, mentioned_at ASC
) tm
JOIN ticker_names tn ON tn.id = tm.ticker_id
LEFT JOIN LATERAL (
  SELECT price, recorded_at
  FROM ticker_prices
  WHERE ticker_id = tm.ticker_id AND recorded_at <= tm.mentioned_at
  ORDER BY recorded_at DESC
  LIMIT 1
) mention_price ON true
LEFT JOIN LATERAL (
  SELECT price, recorded_at
  FROM ticker_prices
  WHERE ticker_id = tm.ticker_id
  ORDER BY recorded_at DESC
  LIMIT 1
) current_price ON true
ORDER BY tn.symbol;

-- name: GetAllMentionsComplete :many
SELECT
  tn.symbol,
  u.username,
  COALESCE(mention_price.price::text, '0') AS mention_price,
  COALESCE(current_price.price::text, '0') AS current_price,
  COALESCE(current_price.recorded_at, now()) AS current_price_date,
  tm.mentioned_at,
  COALESCE((
    SELECT EXP(SUM(LN(ts.ratio::double precision)))
    FROM ticker_splits ts
    WHERE ts.ticker_id = tm.ticker_id
      AND ts.effective_date >= tm.mentioned_at
      AND ts.effective_date <= COALESCE(current_price.recorded_at, now())
  ), 1.0)::double precision AS split_ratio
FROM ticker_mentions tm
JOIN users u ON u.id = tm.user_id
JOIN ticker_names tn ON tn.id = tm.ticker_id
LEFT JOIN LATERAL (
  SELECT price, recorded_at
  FROM ticker_prices
  WHERE ticker_id = tm.ticker_id AND recorded_at <= tm.mentioned_at
  ORDER BY recorded_at DESC
  LIMIT 1
) mention_price ON true
LEFT JOIN LATERAL (
  SELECT price, recorded_at
  FROM ticker_prices
  WHERE ticker_id = tm.ticker_id
  ORDER BY recorded_at DESC
  LIMIT 1
) current_price ON true
WHERE tm.mentioned_at >= $1
ORDER BY tm.mentioned_at ASC;
