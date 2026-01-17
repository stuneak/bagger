CREATE TABLE ticker_mentions (
  id                BIGSERIAL PRIMARY KEY,
  ticker_id         BIGINT NOT NULL REFERENCES tickers(id),
  user_id           BIGINT NOT NULL REFERENCES users(id),
  comment_id        BIGINT REFERENCES comments(id),
  mentioned_at      TIMESTAMP NOT NULL,
  price_at_mention  NUMERIC(18,6) NOT NULL
);

CREATE INDEX idx_mentions_ticker_time
  ON ticker_mentions (ticker_id, mentioned_at DESC);

CREATE INDEX idx_mentions_user
  ON ticker_mentions (user_id);
