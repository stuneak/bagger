# Background Jobs

All jobs run in the `America/New_York` timezone via [gocron](https://github.com/go-co-op/gocron).

## Job Overview

| Job                 | Interval | First Run | Target Table      |
| ------------------- | -------- | --------- | ----------------- |
| nasdaq-tickers-sync | 24h      | on start  | `ticker_names`    |
| ticker-prices       | 6h       | +5 min    | `ticker_prices`   |
| ticker-splits       | 24h      | +5 min    | `ticker_splits`   |
| reddit-scrape-\*    | 3h cycle | staggered | `ticker_mentions` |

---

## 1. nasdaq-tickers-sync

Fetches the full NASDAQ ticker list and upserts into `ticker_names`.

- **Source:** `cron/external_api/nasdaq.go` → `FetchTickers`
- **Runs:** On startup + every 24h
- **Upserts** by symbol — safe to re-run without duplicates
- **Must complete first** so that price/split jobs have ticker IDs to reference
  -- **Rule** ignore symbols with ^ or / signs in their symbols.

---

## 2. ticker-prices

Fetches the current price and volume for every ticker in the database. Override the price, if it's in the same day.

- **Source:** `cron/external_api/yahoo.go` → `FetchCurrentPriceAndVolume`
- **Runs:** 5 min after startup + every 6h
- Iterates all rows from `ticker_names`
- **Inserts** a new row per fetch (append-only, never overwrites)
- `recorded_at` = timestamp returned from the API
- Deduplication via unique constraint `(ticker_id, recorded_at)`

---

## 3. ticker-splits

Fetches stock split history for every ticker.

- **Source:** `cron/external_api/yahoo.go` → `FetchSplits`
- **Runs:** 5 min after startup + every 24h
- Iterates all rows from `ticker_names`
- `effective_date` = `events.splits[<key>].date` from the API response
- Deduplication via unique constraint `(ticker_id, effective_date)`

---

## 4. reddit-scrape-{subreddit}

Scrapes posts and comments from subreddits to extract ticker mentions.

- **Source:** `scrapeSubreddit(subreddit)`
- **Stores:** `ticker_mentions`

### Round-Robin Schedule

| Order | First Run Delay | Repeats        |
| ----- | --------------- | -------------- |
| #1    | 15 minutes      | Every 3h cycle |
| #2    | 1 hour          | Every 3h cycle |
| #3    | 2 hours         | Every 3h cycle |

**Cycle example:**

- T+15m → Subreddit #1
- T+1h → Subreddit #2
- T+2h → Subreddit #3
- T+3h15m → Subreddit #1 (repeat)

---

## General Rules

- All jobs are **idempotent** — rely on database unique constraints to prevent duplicates
- Historical market data is append-only (never overwritten)
- Errors are logged but do not crash the scheduler; the job retries on the next cycle
