# API Package

The `api` package implements the HTTP server and request handlers for the Sopeko backend. It uses the [Gin](https://github.com/gin-gonic/gin) web framework.

## Server (`server.go`)

### Struct

```go
type Server struct {
    store  *db.Queries
    router *gin.Engine
}
```

- `store` — database query layer (sqlc-generated)
- `router` — Gin HTTP router

### Constructor

`NewServer(store *db.Queries, ginMode string) *Server`

Initializes the server with:
- CORS middleware (allows all origins, GET/POST/OPTIONS methods)
- Visitor tracking middleware (logs IP + endpoint asynchronously)
- All route registrations

### Methods

| Method | Description |
|--------|-------------|
| `Start(address string) error` | Starts the HTTP server on the given address |
| `visitorTrackingMiddleware()` | Records each request's IP and endpoint to the database in a background goroutine |

## Routes

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/health` | inline | Returns `{"status": "ok"}` |
| GET | `/api/mentions/:username` | `getUserMentions` | Get ticker mentions for a user |
| GET | `/api/excluded-usernames` | `getExcludedUsernames` | List of excluded usernames |
| GET | `/api/top-performers` | `getTopPerformingUsers` | Top 50 users by cumulative % gain |
| GET | `/api/top-picks` | `getTopPerformingPicks` | Top 50 ticker picks by % gain |
| GET | `/api/worst-picks` | `getWorstPerformingPicks` | Worst 50 ticker picks by % loss |

## Handlers (`handler.go`)

### `getUserMentions`

**GET** `/api/mentions/:username?period=<period>`

Returns all ticker mentions for a given Reddit username with current price performance.

- Filters out excluded usernames (bots/mods)
- Adjusts historical mention prices for stock splits
- Calculates percent change between mention price and current price

**Query params:**
- `period` — `daily`, `weekly`, `monthly`, or omit for all-time

**Response:** `[]MentionResponse`

```json
{
  "symbol": "AAPL",
  "mention_price": "150.00",
  "current_price": "175.50",
  "current_price_date": "2025-01-20T00:00:00Z",
  "percent_change": "+17.00%",
  "split_ratio": 1.0,
  "mentioned_at": "2024-06-15T12:00:00Z"
}
```

### `getExcludedUsernames`

**GET** `/api/excluded-usernames`

Returns the list of excluded Reddit usernames (moderators, bots, special accounts).

### `getTopPerformingPicks` / `getWorstPerformingPicks`

**GET** `/api/top-picks?period=<period>`
**GET** `/api/worst-picks?period=<period>`

Returns the top/worst 50 individual ticker picks sorted by percent change. Excludes mentions from excluded usernames.

**Response:** `[]PickPerformanceResponse`

```json
{
  "symbol": "TSLA",
  "mention_price": "200.00",
  "current_price": "350.00",
  "current_price_date": "2025-01-20T00:00:00Z",
  "percent_change": 75.0,
  "split_ratio": 1.0,
  "mentioned_at": "2024-03-01T00:00:00Z"
}
```

### `getTopPerformingUsers`

**GET** `/api/top-performers?period=<period>`

Returns the top 50 users ranked by total cumulative percent gain across all their picks.

**Response:** `[]TopUserResponse`

```json
{
  "username": "trader123",
  "total_percent_gain": 245.5,
  "picks": [
    {
      "symbol": "NVDA",
      "pick_price": "120.00",
      "current_price": "450.00",
      "percent_gain": 275.0,
      "split_ratio": 1.0
    }
  ]
}
```

## Helper Functions

| Function | Description |
|----------|-------------|
| `parsePeriodCutoff(period string) time.Time` | Converts period string to a cutoff timestamp (daily/weekly/monthly/all-time) |
| `calculatePercentChange(old, new string) string` | Returns formatted percent change string (e.g. `+12.50%`) |
| `calculatePercentChangeFloat(old, new string) float64` | Returns raw percent change as float |
| `adjustPriceForSplits(price string, splitRatio float64) string` | Adjusts a historical price by the cumulative split ratio |
