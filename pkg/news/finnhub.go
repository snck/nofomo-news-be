package news

import (
	"context"
	"strconv"
	"strings"
	"time"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

type FinnHubClient struct {
	client *finnhub.DefaultApiService
}

func NewFinnHubClient(apiKey string) *FinnHubClient {
	cfg := finnhub.NewConfiguration()
	cfg.AddDefaultHeader("X-Finnhub-Token", apiKey)
	client := finnhub.NewAPIClient(cfg).DefaultApi
	return &FinnHubClient{client: client}
}

func (c *FinnHubClient) Fetch(limit int) ([]Article, error) {
	res, _, err := c.client.MarketNews(context.Background()).Category("general").Execute()
	if err != nil {
		return nil, err
	}

	var articles []Article

	for _, news := range res {
		a := Article{
			Source: c.Name(),
		}

		if news.Id != nil {
			a.ExternalID = strconv.FormatInt(*news.Id, 10)
		}

		if news.Headline != nil {
			a.Headline = *news.Headline
		}

		if news.Summary != nil {
			a.Detail = *news.Summary
		}

		if news.Url != nil {
			a.URL = *news.Url
		}

		if news.Datetime != nil {
			a.PublishedAt = time.Unix(*news.Datetime, 0)
		}

		if news.Source != nil {
			a.Publisher = *news.Source
		}

		if news.Related != nil && *news.Related != "" {
			a.Symbols = strings.Split(*news.Related, ",")
		} else {
			a.Symbols = []string{}
		}

		articles = append(articles, a)
	}

	return articles, nil
}

func (c *FinnHubClient) Name() string {
	return "FinnHub"
}
