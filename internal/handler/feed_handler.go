package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"zennews/internal/model"

	"github.com/gin-gonic/gin"
)

type ArticleStore interface {
	GetFeed(limit, offset int) ([]model.FeedArticle, error)
	GetFeedTotal() (int, error)
	GetSymbolsByOriginalIDs(ids []int64) (map[int64][]string, error)
	GetTransformedByID(id int64) (*model.SingleArticle, error)
	GetSymbolsByOriginalID(id int64) ([]string, error)
	GetAllCategories() ([]model.Category, error)
	GetOriginalFeed(limit, offset int) ([]model.OriginalArticle, error)
	GetOriginalFeedTotal() (int, error)
}

type ArticleHandler struct {
	repository ArticleStore
}

func NewArticleHandler(repository ArticleStore) *ArticleHandler {
	return &ArticleHandler{repository: repository}
}

func (h *ArticleHandler) GetFeed(c *gin.Context) {

	limit := getQueryLimit(c)
	offset := getQueryOffset(c)

	articles, err := h.repository.GetFeed(limit, offset)
	if err != nil {
		slog.Error("error fetching feed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	total, err := h.repository.GetFeedTotal()
	if err != nil {
		slog.Error("error fetching feed total", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var originalIDs []int64
	for _, a := range articles {
		originalIDs = append(originalIDs, a.OriginalID)
	}

	symbolMap, err := h.repository.GetSymbolsByOriginalIDs(originalIDs)
	if err != nil {
		slog.Error("error fetching symbols", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var articleRes []ArticleResponse
	for _, a := range articles {

		symbols := symbolMap[a.OriginalID]

		category := CategoryResponse{
			ID:   a.CategoryID,
			Name: a.CategoryName,
		}

		article := ArticleResponse{
			ID:             a.ID,
			Headline:       a.Headline,
			Detail:         a.Detail,
			Publisher:      a.Publisher,
			PublishedAt:    a.PublishedAt.Format(time.RFC3339),
			URL:            a.URL,
			SentimentScore: a.SentimentScore,
			Category:       category,
			Symbols:        symbols,
		}

		articleRes = append(articleRes, article)
	}

	res := FeedResponse{
		Articles: articleRes,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	c.JSON(http.StatusOK, res)
}

func (h *ArticleHandler) GetArticle(c *gin.Context) {
	id := c.Param("id")

	articleId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		slog.Error("invalid article id", "id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article id"})
		return
	}

	article, err := h.repository.GetTransformedByID(articleId)
	if err != nil {
		slog.Error("error fetching article", "error", err, "article_id", articleId)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if article == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	symbols, err := h.repository.GetSymbolsByOriginalID(article.OriginalID)
	if err != nil {
		slog.Error("error fetching symbols", "error", err, "original_id", article.OriginalID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	category := CategoryResponse{
		ID:   article.CategoryID,
		Name: article.CategoryName,
	}

	original := OriginalResponse{
		Headline: article.OriginalHeadline,
		Detail:   article.OriginalDetail,
	}

	res := SingleArticleResponse{
		ID:             article.ID,
		Headline:       article.Headline,
		Detail:         article.Detail,
		Publisher:      article.Publisher,
		PublishedAt:    article.PublishedAt.Format(time.RFC3339),
		URL:            article.URL,
		SentimentScore: article.SentimentScore,
		Category:       category,
		Symbols:        symbols,
		Original:       original,
	}

	c.JSON(http.StatusOK, res)
}

func (h *ArticleHandler) GetCategories(c *gin.Context) {
	categories, err := h.repository.GetAllCategories()
	if err != nil {
		slog.Error("error fetching categories", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var res []CategoryResponse
	for _, c := range categories {
		category := CategoryResponse{
			ID:   c.ID,
			Name: c.Name,
		}
		res = append(res, category)
	}

	c.JSON(http.StatusOK, res)
}

func (h *ArticleHandler) GetHealth(c *gin.Context) {
	_, err := h.repository.GetFeedTotal()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
	})
}

func (h *ArticleHandler) GetOriginalFeed(c *gin.Context) {
	limit := getQueryLimit(c)
	offset := getQueryOffset(c)

	total, err := h.repository.GetOriginalFeedTotal()
	if err != nil {
		slog.Error("error fetching original feed total", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	articles, err := h.repository.GetOriginalFeed(limit, offset)
	if err != nil {
		slog.Error("error fetching original feed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var ids []int64
	for _, a := range articles {
		ids = append(ids, a.ID)
	}

	symbolMap, err := h.repository.GetSymbolsByOriginalIDs(ids)
	if err != nil {
		slog.Error("error fetching symbols", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var articleRes []OriginalArticleResponse
	for _, a := range articles {
		articleRes = append(articleRes, OriginalArticleResponse{
			ID:          a.ID,
			Headline:    a.Headline,
			Detail:      a.Detail,
			URL:         a.URL,
			Source:      a.Source,
			Publisher:   a.Publisher,
			PublishedAt: a.PublishedAt.Format(time.RFC3339),
			Symbols:     symbolMap[a.ID],
		})
	}

	c.JSON(http.StatusOK, OriginalFeedResponse{
		Articles: articleRes,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

func getQueryInt(name string, defaultValue int, c *gin.Context) int {
	paramLimit := c.Query(name)

	if paramLimit == "" {
		return defaultValue
	}

	parsedValue, err := strconv.Atoi(paramLimit)
	if err != nil {
		slog.Warn("invalid query parameter, using default", "param", name, "value", paramLimit, "error", err)
		return defaultValue
	}

	return parsedValue
}

func getQueryLimit(c *gin.Context) int {
	const (
		defaultLimit = 10
		maxLimit     = 100
	)

	limit := getQueryInt("limit", defaultLimit, c)
	if limit < 1 {
		slog.Warn("invalid query parameter, using default", "param", "limit", "value", limit, "default", defaultLimit)
		return defaultLimit
	}

	if limit > maxLimit {
		slog.Warn("query parameter exceeds max, clamping", "param", "limit", "value", limit, "max", maxLimit)
		return maxLimit
	}

	return limit
}

func getQueryOffset(c *gin.Context) int {
	offset := getQueryInt("offset", 0, c)
	if offset < 0 {
		slog.Warn("invalid query parameter, using default", "param", "offset", "value", offset, "default", 0)
		return 0
	}
	return offset
}
