package news

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AlphaVantageClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewAlphaVantageClient(apiKey string) *AlphaVantageClient {
	return &AlphaVantageClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *AlphaVantageClient) Name() string {
	return "AlphaVantage"
}

func (c *AlphaVantageClient) Fetch(limit int) ([]Article, error) {
	url := fmt.Sprintf(
		"https://www.alphavantage.co/query?function=NEWS_SENTIMENT&limit=%d&sort=LATEST&apikey=%s",
		limit, c.apiKey,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("alphavantage fetch: %w", err)
	}
	defer resp.Body.Close()

	var raw avResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("alphavantage decode: %w", err)
	}

	articles := make([]Article, 0, len(raw.Feed))
	for _, item := range raw.Feed {
		publishedAt, err := time.Parse("20060102T150405", item.TimePublished)
		if err != nil {
			publishedAt = time.Time{}
		}

		symbols := make([]string, 0, len(item.TickerSentiment))
		for _, ts := range item.TickerSentiment {
			if ts.Ticker != "" {
				symbols = append(symbols, ts.Ticker)
			}
		}

		articles = append(articles, Article{
			ExternalID:  generateExternalID(item.URL),
			Headline:    item.Title,
			Detail:      item.Summary,
			URL:         item.URL,
			Publisher:   item.Source,
			PublishedAt: publishedAt,
			Symbols:     symbols,
			Source:      c.Name(),
		})
	}

	return articles, nil
}

func generateExternalID(url string) string {
	sum := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", sum)[:16]
}

type avResponse struct {
	Feed []avFeedItem `json:"feed"`
}

type avFeedItem struct {
	Title           string            `json:"title"`
	Summary         string            `json:"summary"`
	URL             string            `json:"url"`
	Source          string            `json:"source"`
	TimePublished   string            `json:"time_published"`
	TickerSentiment []avTickerSentiment `json:"ticker_sentiment"`
}

type avTickerSentiment struct {
	Ticker string `json:"ticker"`
}
