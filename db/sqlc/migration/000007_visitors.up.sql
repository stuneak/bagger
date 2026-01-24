CREATE TABLE visitors (
  id         BIGSERIAL PRIMARY KEY,
  ip_address VARCHAR(255) NOT NULL,
  endpoint   VARCHAR(255) NOT NULL,
  visited_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_visitors_ip ON visitors (ip_address);
CREATE INDEX idx_visitors_endpoint ON visitors (endpoint);
CREATE INDEX idx_visitors_visited_at ON visitors (visited_at DESC);
