package handler

import (
	"log/slog"
	"net/http"
	"time"
	"zennews/internal/model"

	"github.com/gin-gonic/gin"
)

type SummaryStore interface {
	GetSummaries(limit, offset int) ([]model.NewsSummary, error)
	GetSummaryTotal() (int, error)
	GetLatestSummary() (*model.NewsSummary, error)
	GetLatestStories() ([]model.NewsStory, error)
	GetStoriesBySummaryID(summaryID int64) ([]model.NewsStory, error)
}

type SummaryHandler struct {
	repository SummaryStore
}

func NewSummaryHandler(repository SummaryStore) *SummaryHandler {
	return &SummaryHandler{repository: repository}
}

type SummaryResponse struct {
	ID            int64    `json:"id"`
	Paragraph     string   `json:"paragraph"`
	Bullets       []string `json:"bullets"`
	ArticleCount  int      `json:"article_count"`
	FromArticleID int64    `json:"from_article_id"`
	ToArticleID   int64    `json:"to_article_id"`
	ModelUsed     string   `json:"model_used"`
	CreatedAt     string   `json:"created_at"`
}

type SummariesResponse struct {
	Latest  *SummaryResponse  `json:"latest"`
	History []SummaryResponse `json:"history"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

func toSummaryResponse(s model.NewsSummary) SummaryResponse {
	return SummaryResponse{
		ID:            s.ID,
		Paragraph:     s.Paragraph,
		Bullets:       s.Bullets,
		ArticleCount:  s.ArticleCount,
		FromArticleID: s.FromArticleID,
		ToArticleID:   s.ToArticleID,
		ModelUsed:     s.ModelUsed,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
	}
}

func (h *SummaryHandler) GetSummaries(c *gin.Context) {
	limit := getQueryInt("limit", 10, c)
	offset := getQueryInt("offset", 0, c)

	summaries, err := h.repository.GetSummaries(limit, offset)
	if err != nil {
		slog.Error("error fetching summaries", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	total, err := h.repository.GetSummaryTotal()
	if err != nil {
		slog.Error("error fetching summary total", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	res := SummariesResponse{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		History: []SummaryResponse{},
	}

	if len(summaries) > 0 {
		latest := toSummaryResponse(summaries[0])
		res.Latest = &latest
		for _, s := range summaries[1:] {
			res.History = append(res.History, toSummaryResponse(s))
		}
	}

	c.JSON(http.StatusOK, res)
}

type StoryResponse struct {
	ID         int64    `json:"id"`
	SummaryID  int64    `json:"summary_id"`
	Rank       int      `json:"rank"`
	Headline   string   `json:"headline"`
	Summary    string   `json:"summary"`
	Angles     []string `json:"angles"`
	Tickers    []string `json:"tickers"`
	Publishers []string `json:"publishers"`
	TimeRange  string   `json:"time_range"`
}

func toStoryResponse(s model.NewsStory) StoryResponse {
	return StoryResponse{
		ID:         s.ID,
		SummaryID:  s.SummaryID,
		Rank:       s.Rank,
		Headline:   s.Headline,
		Summary:    s.Summary,
		Angles:     s.Angles,
		Tickers:    s.Tickers,
		Publishers: s.Publishers,
		TimeRange:  s.TimeRange,
	}
}

func (h *SummaryHandler) GetLatestStories(c *gin.Context) {
	stories, err := h.repository.GetLatestStories()
	if err != nil {
		slog.Error("error fetching latest stories", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if len(stories) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No stories available"})
		return
	}

	res := make([]StoryResponse, len(stories))
	for i, s := range stories {
		res[i] = toStoryResponse(s)
	}

	c.JSON(http.StatusOK, res)
}

func (h *SummaryHandler) GetStories(c *gin.Context) {
	limit := getQueryInt("limit", 10, c)
	offset := getQueryInt("offset", 0, c)

	summaries, err := h.repository.GetSummaries(limit, offset)
	if err != nil {
		slog.Error("error fetching summaries", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	total, err := h.repository.GetSummaryTotal()
	if err != nil {
		slog.Error("error fetching summary total", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	type SummaryWithStories struct {
		SummaryResponse
		Stories []StoryResponse `json:"stories"`
	}

	items := make([]SummaryWithStories, 0, len(summaries))
	for _, s := range summaries {
		stories, err := h.repository.GetStoriesBySummaryID(s.ID)
		if err != nil {
			slog.Error("error fetching stories for summary", "summary_id", s.ID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		storyResponses := make([]StoryResponse, len(stories))
		for i, st := range stories {
			storyResponses[i] = toStoryResponse(st)
		}

		items = append(items, SummaryWithStories{
			SummaryResponse: toSummaryResponse(s),
			Stories:         storyResponses,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *SummaryHandler) GetLatestSummary(c *gin.Context) {
	summary, err := h.repository.GetLatestSummary()
	if err != nil {
		slog.Error("error fetching latest summary", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if summary == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No summary available"})
		return
	}

	c.JSON(http.StatusOK, toSummaryResponse(*summary))
}
