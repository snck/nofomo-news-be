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
