# ZenNews

ZenNews is a financial news aggregation pipeline that fetches articles from [FinnHub](https://finnhub.io), rewrites them into neutral, fact-focused language using an LLM, and serves the results through a REST API.

## How it works

```
FinnHub API → fetcher → PostgreSQL + Redis queue
                                   ↓
                             transformer → PostgreSQL
                                                    ↑
                                               API (port 8080)
```

1. **fetcher** — pulls the latest market news from FinnHub, saves articles to PostgreSQL, and pushes their IDs to a Redis queue
2. **transformer** — reads from the queue, rewrites each article using an LLM (OpenAI or Anthropic), and saves the result back to PostgreSQL
3. **api** — serves the transformed articles over HTTP

## Prerequisites

- Go 1.21+
- PostgreSQL
- Redis
- A [FinnHub API key](https://finnhub.io)
- An OpenAI or Anthropic API key

## Setup

**1. Create the database and run the migration**

```bash
createdb zen_news
psql zen_news < migration/001_init.sql
```

The migration creates all tables, indexes, and seeds the category list.

**2. Create a `.env` file** in the project root and fill in your values:

```env
DATABASE_URL=postgres://user:password@localhost:5432/zen_news?sslmode=require
REDIS_URL=localhost:6379
FINNHUB_API_KEY=your_finnhub_key
OPENAI_API_KEY=your_openai_key
ANTHROPIC_API_KEY=your_anthropic_key
```

## Running the services

Each service is a separate binary. Run them in separate terminals:

```bash
# Start the API server (port 8080)
go run ./cmd/api

# Fetch latest articles from FinnHub
go run ./cmd/fetcher

# Start the transformer worker
go run ./cmd/transformer
```

The fetcher is a one-shot command — run it on a schedule (e.g. cron) to keep articles fresh. The transformer runs continuously, blocking on the Redis queue until new article IDs arrive.

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/feed` | Paginated list of transformed articles |
| `GET` | `/feed/:id` | Single article with original vs. transformed comparison |
| `GET` | `/categories` | All available categories |
| `GET` | `/health` | Service health check |

### Query parameters for `/feed`

| Parameter | Default | Description |
|-----------|---------|-------------|
| `limit` | `10` | Number of articles to return |
| `offset` | `0` | Number of articles to skip |

## Building

```bash
go build ./cmd/api
go build ./cmd/fetcher
go build ./cmd/transformer
```

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Vet
go vet ./...
```

## Deployment
```
gcloud builds submit --config=cloudbuild-api.yaml
gcloud builds submit --config=cloudbuild-fetcher.yaml
gcloud builds submit --config=cloudbuild-transformer.yaml
```