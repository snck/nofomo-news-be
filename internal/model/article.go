package model

import "time"

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	OthersCategory   = "Others"
)

type OriginalArticle struct {
	ID          int64
	Headline    string
	Detail      string
	URL         string
	Source      string
	Publisher   string
	PublishedAt time.Time
	FetchedAt   time.Time
	ExternalID  string
	Status      string
}

type TransformedArticle struct {
	ID             int64
	Headline       string
	Detail         string
	OriginalID     int64
	CategoryID     int64
	SentimentScore int
	PromptVersion  string
	ModelUsed      string
	TransformedAt  time.Time
}

type Category struct {
	ID   int64
	Name string
}

type ArticleSymbol struct {
	ID        int64
	ArticleId int64
	Symbol    string
	CreatedAt time.Time
}

type ProcessingError struct {
	ID           int64
	ArticleId    int64
	ErrorMessage string
	ErrorType    string
	AttemptCount int
	CreatedAt    time.Time
}

type ApiUsage struct {
	ID           int64
	ApiName      string
	UsageDate    time.Time
	RequestCount int
	TokenCount   int
}

type FeedArticle struct {
	ID             int64
	Headline       string
	Detail         string
	SentimentScore int
	Publisher      string
	PublishedAt    time.Time
	URL            string
	CategoryID     int64
	CategoryName   string
	OriginalID     int64
}

type SingleArticle struct {
	FeedArticle
	OriginalHeadline string
	OriginalDetail   string
}
