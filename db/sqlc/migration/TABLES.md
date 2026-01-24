# Database Tables

## users

Registered platform users.

| Column     | Type      | Constraints              |
|------------|-----------|--------------------------|
| id         | BIGSERIAL | PRIMARY KEY              |
| username   | TEXT      | NOT NULL, UNIQUE         |
| created_at | TIMESTAMP | NOT NULL, DEFAULT now()  |

Indexes: `idx_users_username` on `(username)`

---

## ticker_names

Stock ticker symbols and their associated company info.

| Column       | Type      | Constraints              |
|--------------|-----------|--------------------------|
| id           | BIGSERIAL | PRIMARY KEY              |
| symbol       | TEXT      | NOT NULL, UNIQUE         |
| company_name | TEXT      | NOT NULL                 |
| exchange     | TEXT      | NOT NULL (e.g. "NASDAQ") |
| currency     | TEXT      | NOT NULL, DEFAULT 'USD'  |
| created_at   | TIMESTAMP | NOT NULL, DEFAULT now()  |

---

## comments

User posts/comments collected from external sources (Reddit, Twitter, etc.).

| Column      | Type      | Constraints                          |
|-------------|-----------|--------------------------------------|
| id          | BIGSERIAL | PRIMARY KEY                          |
| user_id     | BIGINT    | NOT NULL, FK -> users(id)            |
| source      | TEXT      | NOT NULL (reddit, twitter, etc.)     |
| external_id | TEXT      | NOT NULL (original post/comment id)  |
| content     | TEXT      | NOT NULL (post/comment body)         |
| created_at  | TIMESTAMP | NOT NULL                             |

Unique: `(user_id, external_id)`
Indexes: `idx_comments_user_time` on `(user_id, created_at DESC)`

---

## ticker_prices

Historical price and volume data for tickers.

| Column      | Type          | Constraints                    |
|-------------|---------------|--------------------------------|
| id          | BIGSERIAL     | PRIMARY KEY                    |
| ticker_id   | BIGINT        | NOT NULL, FK -> ticker_names(id) |
| price       | NUMERIC(18,2) | NOT NULL                       |
| recorded_at | TIMESTAMPTZ   | NOT NULL                       |
| volume      | BIGINT        | NOT NULL, DEFAULT 0            |

Unique: `(ticker_id, recorded_at)`
Indexes:
- `idx_ticker_prices_latest` on `(ticker_id, recorded_at DESC)`
- `idx_ticker_prices_highest` on `(ticker_id, recorded_at, price DESC)`

---

## ticker_mentions

Links a comment to a ticker it mentions, tracking which user mentioned which ticker and when.

| Column       | Type      | Constraints                      |
|--------------|-----------|----------------------------------|
| id           | BIGSERIAL | PRIMARY KEY                      |
| ticker_id    | BIGINT    | NOT NULL, FK -> ticker_names(id) |
| user_id      | BIGINT    | NOT NULL, FK -> users(id)        |
| comment_id   | BIGINT    | NOT NULL, FK -> comments(id)     |
| mentioned_at | TIMESTAMP | NOT NULL                         |

Indexes:
- `idx_mentions_ticker_time` on `(ticker_id, mentioned_at DESC)`
- `idx_mentions_user` on `(user_id)`
- `idx_mentions_user_ticker` on `(user_id, ticker_id, mentioned_at ASC)`

---

## visitors

Tracks API endpoint visits by IP address.

| Column     | Type         | Constraints             |
|------------|--------------|-------------------------|
| id         | BIGSERIAL    | PRIMARY KEY             |
| ip_address | VARCHAR(255) | NOT NULL                |
| endpoint   | VARCHAR(255) | NOT NULL                |
| visited_at | TIMESTAMP    | NOT NULL, DEFAULT NOW() |

Indexes:
- `idx_visitors_ip` on `(ip_address)`
- `idx_visitors_endpoint` on `(endpoint)`
- `idx_visitors_visited_at` on `(visited_at DESC)`

---

## ticker_splits

Records stock split events for tickers.

| Column         | Type          | Constraints                      |
|----------------|---------------|----------------------------------|
| id             | BIGSERIAL     | PRIMARY KEY                      |
| ticker_id      | BIGINT        | NOT NULL, FK -> ticker_names(id) |
| ratio          | NUMERIC(10,4) | NOT NULL                         |
| effective_date | DATE          | NOT NULL                         |

Unique: `(ticker_id, effective_date)`
Indexes: `idx_ticker_splits_ticker_date` on `(ticker_id, effective_date)`
