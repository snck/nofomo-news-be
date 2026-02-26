package handler

type ArticleResponse struct {
	ID             int64            `json:"id"`
	Headline       string           `json:"headline"`
	Detail         string           `json:"detail"`
	Publisher      string           `json:"publisher"`
	PublishedAt    string           `json:"published_at"`
	URL            string           `json:"url"`
	SentimentScore int              `json:"sentiment_score"`
	Category       CategoryResponse `json:"category"`
	Symbols        []string         `json:"symbols"`
}

type CategoryResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type FeedResponse struct {
	Articles []ArticleResponse `json:"articles"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

type SingleArticleResponse struct {
	ID             int64            `json:"id"`
	Headline       string           `json:"headline"`
	Detail         string           `json:"detail"`
	Publisher      string           `json:"publisher"`
	PublishedAt    string           `json:"published_at"`
	URL            string           `json:"url"`
	SentimentScore int              `json:"sentiment_score"`
	Category       CategoryResponse `json:"category"`
	Symbols        []string         `json:"symbols"`
	Original       OriginalResponse `json:"original"`
}

type OriginalResponse struct {
	Headline string `json:"headline"`
	Detail   string `json:"detail"`
}

type OriginalArticleResponse struct {
	ID          int64    `json:"id"`
	Headline    string   `json:"headline"`
	Detail      string   `json:"detail"`
	URL         string   `json:"url"`
	Source      string   `json:"source"`
	Publisher   string   `json:"publisher"`
	PublishedAt string   `json:"published_at"`
	Symbols     []string `json:"symbols"`
}

type OriginalFeedResponse struct {
	Articles []OriginalArticleResponse `json:"articles"`
	Total    int                       `json:"total"`
	Limit    int                       `json:"limit"`
	Offset   int                       `json:"offset"`
}
