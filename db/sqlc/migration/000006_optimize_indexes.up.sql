-- Index to optimize GetHighestPriceAfterDate query
-- Covers (ticker_id, recorded_at) filter with price for sorting
CREATE INDEX idx_ticker_prices_highest
  ON ticker_prices (ticker_id, recorded_at, price DESC);

-- Index on ticker_mentions for unique user-ticker lookups
CREATE INDEX idx_mentions_user_ticker
  ON ticker_mentions (user_id, ticker_id, mentioned_at ASC);

-- Index on users username for faster lookups
CREATE INDEX idx_users_username ON users (username);
