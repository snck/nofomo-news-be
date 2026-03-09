package news

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const marketauxPerPage = 3

type MarketauxClient struct {
	apiKey     string
	maxPages   int
	httpClient *http.Client
}

func NewMarketauxClient(apiKey string, maxPages int) *MarketauxClient {
	return &MarketauxClient{
		apiKey:     apiKey,
		maxPages:   maxPages,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *MarketauxClient) Name() string {
	return "Marketaux"
}

func (c *MarketauxClient) Fetch(limit int) ([]Article, error) {
	var articles []Article

	for page := 1; page <= c.maxPages; page++ {
		url := fmt.Sprintf(
			"https://api.marketaux.com/v1/news/all?language=en&limit=%d&page=%d&api_token=%s",
			marketauxPerPage, page, c.apiKey,
		)

		resp, err := c.httpClient.Get(url)
		if err != nil {
			return articles, fmt.Errorf("marketaux fetch page %d: %w", page, err)
		}

		var raw marketauxResponse
		err = json.NewDecoder(resp.Body).Decode(&raw)
		resp.Body.Close()
		if err != nil {
			return articles, fmt.Errorf("marketaux decode page %d: %w", page, err)
		}

		if len(raw.Data) == 0 {
			break
		}

		for _, item := range raw.Data {
			articles = append(articles, c.toArticle(item))
		}

		slog.Info("marketaux page fetched", "page", page, "count", len(raw.Data))

		if len(articles) >= limit {
			break
		}
	}

	return articles, nil
}

func (c *MarketauxClient) toArticle(item marketauxArticle) Article {
	publishedAt, err := time.Parse("2006-01-02T15:04:05.000000Z", item.PublishedAt)
	if err != nil {
		publishedAt = time.Time{}
	}

	detail := item.Description
	if detail == "" {
		detail = item.Snippet
	}

	var symbols []string
	for _, entity := range item.Entities {
		if entity.Type == "equity" {
			symbols = append(symbols, entity.Symbol)
		}
	}

	return Article{
		ExternalID:  item.UUID,
		Headline:    item.Title,
		Detail:      detail,
		URL:         item.URL,
		Publisher:   item.Source,
		PublishedAt: publishedAt,
		Symbols:     symbols,
		Source:      c.Name(),
	}
}

type marketauxResponse struct {
	Data []marketauxArticle `json:"data"`
}

type marketauxArticle struct {
	UUID        string            `json:"uuid"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Snippet     string            `json:"snippet"`
	URL         string            `json:"url"`
	Source      string            `json:"source"`
	PublishedAt string            `json:"published_at"`
	Entities    []marketauxEntity `json:"entities"`
}

type marketauxEntity struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
}
