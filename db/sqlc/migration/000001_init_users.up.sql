CREATE TABLE users (
  id           BIGSERIAL PRIMARY KEY,
  username     TEXT NOT NULL UNIQUE,
  created_at   TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_username ON users (username);
