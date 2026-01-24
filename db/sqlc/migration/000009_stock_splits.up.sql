CREATE TABLE stock_splits (
  id             BIGSERIAL PRIMARY KEY,
  ticker_id      BIGINT NOT NULL REFERENCES tickers(id),
  ratio          NUMERIC(10,4) NOT NULL,
  effective_date DATE NOT NULL,
  created_at     TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE (ticker_id, effective_date)
);

CREATE INDEX idx_stock_splits_ticker_date
  ON stock_splits (ticker_id, effective_date);
