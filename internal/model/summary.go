package model

import "time"

type NewsSummary struct {
	ID            int64
	Paragraph     string
	Bullets       []string
	ArticleCount  int
	FromArticleID int64
	ToArticleID   int64
	ModelUsed     string
	CreatedAt     time.Time
}

type NewsStory struct {
	ID        int64
	SummaryID int64
	Rank      int
	Headline  string
	Summary   string
	TimeRange string
	Angles    []string
	Tickers   []string
	Publishers []string
}
