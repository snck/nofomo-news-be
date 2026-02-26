# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Database

Migration file: `migration/001_init.sql`

```bash
createdb zen_news
psql zen_news < migration/001_init.sql
```

Creates tables `original_article`, `transformed_article`, `category`, `article_symbol`, `processing_error`, along with indexes and the seeded category list (Earnings, Market Movement, Economy, Crypto, Mergers & Acquisitions, Policy & Regulation, Company News, Analysis, Others).

## Commands

```bash
# Build individual services
go build ./cmd/api
go build ./cmd/fetcher
go build ./cmd/transformer

# Run services directly
go run ./cmd/api
go run ./cmd/fetcher
go run ./cmd/transformer

# Test
go test ./...
go test ./internal/...
go test ./pkg/...

# Format and vet
go fmt ./...
go vet ./...

# Dependencies
go mod tidy
```

## Environment Variables

| Variable | Default | Required by |
|---|---|---|
| `DATABASE_URL` | — (required) | All services |
| `REDIS_URL` | `localhost:6379` | Fetcher, Transformer |
| `FINNHUB_API_KEY` | — | Fetcher |
| `ALPHA_VANTAGE_API_KEY` | — | Fetcher |
| `OPENAI_API_KEY` | — | Transformer (if using OpenAI) |
| `ANTHROPIC_API_KEY` | — | Transformer (if using Anthropic) |

## Architecture

ZenNews is a financial news aggregation and transformation pipeline with three independent services:

```
FinnHub API ↘
              Fetcher → PostgreSQL + Redis queue
Alpha Vantage ↗
                                    ↓
                              Transformer → PostgreSQL (transformed articles)
                                                      ↑
                                                  API server (port 8080)
```

### Services (`cmd/`)

- **`api`** — Gin REST API on port 8080. CORS allows `http://localhost:3000`. Read-only endpoints.
- **`fetcher`** — Pulls articles from FinnHub and/or Alpha Vantage (whichever API keys are set), saves to `original_article`, pushes article IDs to Redis queue `zennews:queue:transform`.
- **`transformer`** — Blocking-pops from Redis queue (`BRPop` with 0 timeout), calls LLM, saves to `transformed_article`. Retries up to 3 times; failed articles go to `zennews:queue:failed`.

### API Endpoints

- `GET /feed` — Paginated transformed articles (`limit`, `offset` params)
- `GET /feed/:id` — Single article with original vs. transformed comparison
- `GET /categories` — All categories
- `GET /summaries` — Paginated news summaries, latest first (`limit`, `offset` params)
- `GET /summaries/latest` — Latest news summary only
- `GET /health` — DB connectivity check

### LLM Integration (`pkg/llm/`)

Two interchangeable implementations behind the `LLMClient` interface:
- **OpenAI** (`openai.go`): `gpt-4o-mini`
- **Anthropic** (`anthropic.go`): `claude-haiku-4-5`

Both use the same system prompt (financial news editor persona) and return the same `TransformResult`: neutral headline, neutral summary, category, and sentiment score (1–10).

### Data Layer

- **`db/postgres.go`** — Connection pool (max 25 conns). Repository methods live in `internal/repository/article_repository.go`.
- **`db/redis.go`** — Queue client. Uses blocking right-pop; zero timeout means infinite wait.

### Domain Models (`internal/model/article.go`)

Key types: `OriginalArticle`, `TransformedArticle`, `Category`, `ArticleSymbol`, `FeedArticle` (combined view for the feed endpoint), `SingleArticle` (combined view with original comparison).

Article status lifecycle: `pending` → `processing` → `completed` / `failed`.

### News Integration (`pkg/news/`)

`NewsClient` interface with two implementations: FinnHub (`finnhub.go`) and Alpha Vantage (`alphavantage.go`). The fetcher activates whichever clients have API keys configured. FinnHub fetches from the `general` market news category; Alpha Vantage uses the `NEWS_SENTIMENT` endpoint. Each article may include related stock symbols, stored in `article_symbol`.
