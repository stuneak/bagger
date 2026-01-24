CREATE TABLE ticker_splits (
  id             BIGSERIAL PRIMARY KEY,
  ticker_id      BIGINT NOT NULL REFERENCES ticker_names(id),
  ratio          NUMERIC(10,4) NOT NULL,
  effective_date DATE NOT NULL,
  UNIQUE (ticker_id, effective_date)
);

CREATE INDEX idx_ticker_splits_ticker_date
  ON ticker_splits (ticker_id, effective_date);
