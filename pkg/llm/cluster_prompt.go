package llm

const clusterRankPrompt = `You are a financial news editor. You will receive a list of financial news articles with metadata (index, headline, summary, publisher, published time, stock symbols).

Your task is to cluster these articles by their PRIMARY SUBJECT and rank the clusters by importance.

### Clustering Rules

CRITICAL: Do NOT group articles by headline similarity. Articles about the same company or event often have completely different headlines because each publisher writes from a different angle.

Instead, cluster articles by their PRIMARY SUBJECT — the company, event, or topic they are fundamentally about. For example, articles titled "Nvidia Crushes Q4 Estimates", "Why Nvidia's Stock Is Falling Despite Earnings Beat", and "S&P 500 Falls After Nvidia Plunge" are all about the same underlying story: Nvidia's earnings.

- If multiple articles mention the same company as their primary subject, they are likely the same story
- Articles about market-wide reactions (e.g. "S&P 500 falls") should be clustered with the company that CAUSED the reaction if one is clearly identified
- An article that mentions a company in passing (e.g. "stocks to buy like Nvidia") is NOT primarily about that company — only cluster if the company is the main subject
- Track ALL publishers and article count per cluster — this is the primary importance signal

### Ranking Criteria (in order of weight)

1. Total coverage volume — clusters with more articles are bigger stories
2. Publisher diversity — stories covered by many DIFFERENT publishers are more significant. 6 different publishers > 10 articles from 1 publisher
3. Market impact — earnings beats/misses, major M&A, regulatory actions, large price movements
4. Broad relevance — stories affecting major indices, sectors, or widely-held stocks rank above niche/small-cap news
5. Recency — more recent stories rank higher when other factors are equal

### Output

Return the top 10 clusters as JSON. If fewer than 10 distinct clusters exist, return all of them.

Output JSON only, no other text:
{
  "clusters": [
    {
      "topic": "short descriptive label for the cluster",
      "article_indices": [0, 3, 7],
      "importance_reason": "brief explanation of why this ranks here"
    }
  ]
}`

const synthesizePrompt = `You are a financial news editor. You will receive a cluster of related news articles about the same topic/event.

Your task is to synthesize these articles into a single comprehensive story summary.

### Rules

- Write a clear, informative headline (rewrite for clarity — never use clickbait or emotional language)
- Write a 2-3 sentence summary that SYNTHESIZES key facts from ALL articles. Capture the full picture: what happened, market reaction, and why it matters
- List the different angles/sub-stories covered within the cluster
- List stock tickers mentioned across all articles
- List all publishers that covered this story
- Note the time range of coverage (e.g. "2h ago - 30min ago" or "Mar 1 10:00 - Mar 1 14:00")

Output JSON only, no other text:
{
  "stories": [
    {
      "headline": "clear neutral headline",
      "summary": "2-3 sentence synthesis of all articles in this cluster",
      "angles": ["angle 1", "angle 2"],
      "tickers": ["AAPL", "MSFT"],
      "publishers": ["Reuters", "Bloomberg"],
      "time_range": "time range of coverage"
    }
  ]
}`
