package news

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

type RSSClient struct {
	url        string
	sourceName string
	parser     *gofeed.Parser
}

func NewRSSClient(url, sourceName string) *RSSClient {
	return &RSSClient{
		url:        url,
		sourceName: sourceName,
		parser:     gofeed.NewParser(),
	}
}

func (c *RSSClient) Fetch(limit int) ([]Article, error) {
	feed, err := c.parser.ParseURL(c.url)
	if err != nil {
		slog.Error("failed to parse RSS feed", "url", c.url, "error", err)
		return nil, err
	}

	var articles []Article
	for _, item := range feed.Items {
		if len(articles) >= limit {
			break
		}

		detail := item.Description
		if detail == "" {
			detail = item.Content
		}
		detail = stripHTML(detail)

		externalID := item.GUID
		if externalID == "" {
			externalID = item.Link
		}

		a := Article{
			ExternalID: externalID,
			Headline:   item.Title,
			Detail:     detail,
			URL:        item.Link,
			Source:      c.Name(),
			Publisher:   c.sourceName,
			Symbols:    []string{},
		}

		if item.PublishedParsed != nil {
			a.PublishedAt = *item.PublishedParsed
		}

		articles = append(articles, a)
	}

	return articles, nil
}

func (c *RSSClient) Name() string {
	return c.sourceName
}

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
