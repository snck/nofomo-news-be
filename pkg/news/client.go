package news

import "time"

type Article struct {
	ExternalID  string
	Headline    string
	Detail      string
	URL         string
	Source      string
	PublishedAt time.Time
	Symbols     []string
	Publisher   string
}

type NewsClient interface {
	Fetch(limit int) ([]Article, error)
	Name() string
}
