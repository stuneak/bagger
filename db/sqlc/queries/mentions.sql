-- name: CreateTickerMention :one
INSERT INTO ticker_mentions (
  ticker_id,
  user_id,
  comment_id,
  mentioned_at,
  price_at_mention
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
