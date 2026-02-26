package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MassiveClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewMassiveClient(apiKey string) *MassiveClient {
	return &MassiveClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *MassiveClient) Name() string {
	return "Massive"
}

func (c *MassiveClient) Fetch(limit int) ([]Article, error) {
	url := fmt.Sprintf(
		"https://api.massive.com/v2/reference/news?limit=%d&order=desc&sort=published_utc&apiKey=%s",
		limit, c.apiKey,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("massive fetch: %w", err)
	}
	defer resp.Body.Close()

	var raw massiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("massive decode: %w", err)
	}

	articles := make([]Article, 0, len(raw.Results))
	for _, item := range raw.Results {
		publishedAt, err := time.Parse(time.RFC3339, item.PublishedUTC)
		if err != nil {
			publishedAt = time.Time{}
		}

		articles = append(articles, Article{
			ExternalID:  item.ID,
			Headline:    item.Title,
			Detail:      item.Description,
			URL:         item.ArticleURL,
			Publisher:   item.Publisher.Name,
			PublishedAt: publishedAt,
			Symbols:     item.Tickers,
			Source:      c.Name(),
		})
	}

	return articles, nil
}

type massiveResponse struct {
	Results []massiveResult `json:"results"`
}

type massiveResult struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	ArticleURL   string           `json:"article_url"`
	PublishedUTC string           `json:"published_utc"`
	Tickers      []string         `json:"tickers"`
	Publisher    massivePublisher `json:"publisher"`
}

type massivePublisher struct {
	Name string `json:"name"`
}
