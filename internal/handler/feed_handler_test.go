package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"zennews/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

type fakeStore struct {
	feed       []model.FeedArticle
	feedTotal  int
	symbolMap  map[int64][]string
	article    *model.SingleArticle
	symbols    []string
	categories []model.Category
	err        error
}

func (f *fakeStore) GetFeed(limit int, offset int) ([]model.FeedArticle, error) {
	return f.feed, f.err
}

func (f *fakeStore) GetFeedTotal() (int, error) {
	return f.feedTotal, f.err
}

func (f *fakeStore) GetSymbolsByOriginalIDs(ids []int64) (map[int64][]string, error) {
	return f.symbolMap, f.err
}

func (f *fakeStore) GetTransformedByID(id int64) (*model.SingleArticle, error) {
	return f.article, f.err
}

func (f *fakeStore) GetSymbolsByOriginalID(id int64) ([]string, error) {
	return f.symbols, f.err
}

func (f *fakeStore) GetAllCategories() ([]model.Category, error) {
	return f.categories, f.err
}

func newTestRouter(store ArticleStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewArticleHandler(store)
	r.GET("/feed", h.GetFeed)
	r.GET("/feed/:id", h.GetArticle)
	r.GET("/categories", h.GetCategories)
	r.GET("/health", h.GetHealth)
	return r
}

func TestGetFeed_ReturnArticles(t *testing.T) {
	store := &fakeStore{
		feed: []model.FeedArticle{
			{ID: 1, Headline: "Test headline", OriginalID: 10},
		},
		feedTotal: 1,
		symbolMap: map[int64][]string{10: {"AAPL"}},
	}

	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/feed?limit=10&ofset=0", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res FeedResponse
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, 1, res.Total)
	assert.Equal(t, len(res.Articles), 1)
	assert.Equal(t, "Test headline", res.Articles[0].Headline)
	assert.Equal(t, []string{"AAPL"}, res.Articles[0].Symbols)
}

func TestGetFeed_DBError(t *testing.T) {
	store := &fakeStore{err: errors.New("DB down")}
	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/feed?limit=10&offset=0", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetFeed_DefaultLimit(t *testing.T) {
	store := &fakeStore{
		feed:      []model.FeedArticle{},
		feedTotal: 0,
	}
	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/feed", nil)
	r.ServeHTTP(w, req)

	var res FeedResponse
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, 10, res.Limit)
	assert.Equal(t, 0, res.Offset)
}

func TestGetArticle_Found(t *testing.T) {
	store := &fakeStore{
		article: &model.SingleArticle{
			FeedArticle: model.FeedArticle{
				ID:       1,
				Headline: "Transformed headline",
			},
			OriginalHeadline: "Original headline",
		},
		symbols: []string{"AAPL"},
	}

	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/feed/1", nil)
	r.ServeHTTP(w, req)

	var res SingleArticleResponse
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "Transformed headline", res.Headline)
	assert.Equal(t, "Original headline", res.Original.Headline)
	assert.Equal(t, []string{"AAPL"}, res.Symbols)
}

func TestGetArticle_NotFound(t *testing.T) {
	store := &fakeStore{}
	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/feed/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetArticle_InvalidID(t *testing.T) {
	store := &fakeStore{}
	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/feed/aaa", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetHealth_Healthy(t *testing.T) {
	store := &fakeStore{feedTotal: 0}
	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "healthy", res["status"])
}

func TestGetHealth_Unhealthy(t *testing.T) {
	store := &fakeStore{err: errors.New("DB down")}
	r := newTestRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "unhealthy", res["status"])
}
