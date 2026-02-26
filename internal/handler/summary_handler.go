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
