CREATE TABLE ticker_prices (
  ticker_id     BIGINT NOT NULL REFERENCES tickers(id),
  price         NUMERIC(18,6) NOT NULL,
  recorded_at   TIMESTAMP NOT NULL,

  PRIMARY KEY (ticker_id, recorded_at)
);

CREATE INDEX idx_prices_latest
  ON ticker_prices (ticker_id, recorded_at DESC);
