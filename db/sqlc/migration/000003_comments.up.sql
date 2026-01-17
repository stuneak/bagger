CREATE TABLE comments (
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES users(id),
  source       TEXT NOT NULL, -- reddit | twitter etc 
  external_id  TEXT, -- post / comment id
  content      TEXT, -- post / comment content
  created_at   TIMESTAMP NOT NULL
);

CREATE INDEX idx_comments_user_time
  ON comments (user_id, created_at DESC);
