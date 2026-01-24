# Ticker Mentions Queries

SQLC queries for the `ticker_mentions` table. These handle creation and retrieval of ticker mentions with enriched price and split data.

---

## CreateTickerMention

Inserts a new ticker mention record.

| Parameter | Type      | Description                        |
|-----------|-----------|------------------------------------|
| $1        | BIGINT    | ticker_id (FK -> ticker_names)     |
| $2        | BIGINT    | user_id (FK -> users)              |
| $3        | BIGINT    | comment_id (FK -> comments)        |
| $4        | TIMESTAMP | mentioned_at                       |

**Returns:** The inserted row.

---

## GetUserMentionsComplete

Returns the **first mention** of each ticker by a given user (within a time range), enriched with:
- The ticker symbol
- The price at the time of mention (closest preceding price)
- The most recent price
- A cumulative stock split adjustment ratio for the period between mention and current price

| Parameter | Type      | Description                                  |
|-----------|-----------|----------------------------------------------|
| $1        | TEXT      | username (looked up in `users` table)        |
| $2        | TIMESTAMP | earliest `mentioned_at` to include           |

**Logic:**
1. Selects `DISTINCT ON (ticker_id)` ordered by `mentioned_at ASC` to get each ticker's first mention.
2. Joins `ticker_names` for the symbol.
3. Uses `LATERAL` subqueries on `ticker_prices` to find:
   - `mention_price`: most recent price recorded on or before `mentioned_at`.
   - `current_price`: the latest price recorded for that ticker.
4. Computes `split_ratio` as the product of all `ticker_splits.ratio` values with `effective_date` between the mention and the current price date.

**Returns:** Rows ordered by `symbol`, each containing:

| Column             | Type             | Description                                    |
|--------------------|------------------|------------------------------------------------|
| symbol             | TEXT             | Ticker symbol                                  |
| mention_price      | TEXT             | Price at time of mention (or '0')              |
| current_price      | TEXT             | Latest recorded price (or '0')                 |
| current_price_date | TIMESTAMPTZ      | When the current price was recorded            |
| mentioned_at       | TIMESTAMP        | When the user first mentioned this ticker      |
| split_ratio        | DOUBLE PRECISION | Cumulative split adjustment (default 1.0)      |

---

## GetAllMentionsComplete

Returns **all** ticker mentions across all users (within a time range), enriched with the same price and split data as `GetUserMentionsComplete`.

| Parameter | Type      | Description                          |
|-----------|-----------|--------------------------------------|
| $1        | TIMESTAMP | earliest `mentioned_at` to include   |

**Differences from GetUserMentionsComplete:**
- No `DISTINCT ON` -- returns every mention, not just the first per ticker.
- Joins `users` to include `username` in output.
- Ordered by `mentioned_at ASC` instead of `symbol`.

**Returns:** Rows ordered by `mentioned_at ASC`, each containing:

| Column             | Type             | Description                                    |
|--------------------|------------------|------------------------------------------------|
| symbol             | TEXT             | Ticker symbol                                  |
| username           | TEXT             | User who made the mention                      |
| mention_price      | TEXT             | Price at time of mention (or '0')              |
| current_price      | TEXT             | Latest recorded price (or '0')                 |
| current_price_date | TIMESTAMPTZ      | When the current price was recorded            |
| mentioned_at       | TIMESTAMP        | When the mention occurred                      |
| split_ratio        | DOUBLE PRECISION | Cumulative split adjustment (default 1.0)      |
