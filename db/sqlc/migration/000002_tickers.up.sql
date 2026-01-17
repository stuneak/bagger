
CREATE TABLE tickers (
  id            BIGSERIAL PRIMARY KEY,
  symbol        TEXT NOT NULL UNIQUE,
  company_name  TEXT,
  exchange      TEXT, -- nasdaq | 
  currency      TEXT NOT NULL DEFAULT 'USD',
  created_at    TIMESTAMP NOT NULL DEFAULT now()
);
