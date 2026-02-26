# Plan: Add Alpha Vantage as a News Source

## Context

The fetcher currently only pulls news from FinnHub. Adding Alpha Vantage as a second source increases article coverage and diversity. The existing `NewsClient` interface in `pkg/news/client.go` already supports multiple implementations — we just need a new one and a multi-client loop in the fetcher.

## Files to Create

### 1. `pkg/news/alphavantage.go` — Alpha Vantage `NewsClient` implementation

Uses Go stdlib only (`net/http`, `encoding/json`, `crypto/sha256`). No new dependencies.

- **Struct**: `AlphaVantageClient` with `apiKey` and `*http.Client` (30s timeout)
- **Constructor**: `NewAlphaVantageClient(apiKey string)`
- **`Name()`**: returns `"AlphaVantage"`
- **`Fetch(limit int)`**:
  - `GET https://www.alphavantage.co/query?function=NEWS_SENTIMENT&limit=N&sort=LATEST&apikey=KEY`
  - Decode JSON into unexported response structs
  - Map each feed item to `news.Article`:

    | Alpha Vantage field | → | `Article` field |
    |---|---|---|
    | `title` | → | `Headline` |
    | `summary` | → | `Detail` |
    | `url` | → | `URL` |
    | `source` | → | `Publisher` |
    | `time_published` (`"20260226T075324"`) | → | `PublishedAt` (parse with layout `"20060102T150405"`) |
    | `ticker_sentiment[].ticker` | → | `Symbols` |
    | (none — hash the URL) | → | `ExternalID` (first 16 chars of SHA-256 hex of URL) |
    | `c.Name()` | → | `Source` |

- **Helper**: `generateExternalID(url string) string` — `sha256(url)[:16]` hex. Deterministic, collision-resistant. The real dedup is the DB's `ON CONFLICT (url)`.

### 2. `pkg/news/alphavantage_test.go` — unit tests

- `TestGenerateExternalID` — same URL → same ID, different URLs → different IDs
- `TestParseTimePublished` — validates `"20060102T150405"` layout works
- `TestFetch` — use `httptest.NewServer` to return a canned JSON response, verify articles are mapped correctly

## Files to Modify

### 3. `cmd/fetcher/main.go` — multi-client fetcher loop

Replace the single FinnHub client with an opt-in client list:

```go
var clients []news.NewsClient
if key := os.Getenv("FINNHUB_API_KEY"); key != "" {
    clients = append(clients, news.NewFinnHubClient(key))
}
if key := os.Getenv("ALPHA_VANTAGE_API_KEY"); key != "" {
    clients = append(clients, news.NewAlphaVantageClient(key))
}

if len(clients) == 0 {
    slog.Error("no news source API keys configured")
    return
}
```

Then loop over clients, running the existing fetch-and-save logic for each. Key behaviors:
- If one source fails, log and `continue` to the next (don't abort)
- Log `"source"` field on all slog calls so output shows which source each message belongs to
- Per-source stats (saved/duplicated/errors)

### 4. `CLAUDE.md` and `README.md` — documentation

- Add `ALPHA_VANTAGE_API_KEY` to the environment variables table
- Mention both news sources in the architecture section

## Implementation Order

1. `pkg/news/alphavantage.go` — self-contained, no existing code touched
2. `pkg/news/alphavantage_test.go` — validate before integrating
3. `cmd/fetcher/main.go` — integrate multi-client loop
4. `CLAUDE.md` / `README.md` — update docs
5. `go build ./...` — verify compilation

## Verification

```bash
# 1. Build
go build ./...

# 2. Run all tests
go test ./...

# 3. Run the fetcher with only Alpha Vantage configured
ALPHA_VANTAGE_API_KEY=your_key go run ./cmd/fetcher

# 4. Run with both sources
FINNHUB_API_KEY=your_key ALPHA_VANTAGE_API_KEY=your_key go run ./cmd/fetcher

# 5. Verify articles appear in the feed
curl http://localhost:8080/feed
```
