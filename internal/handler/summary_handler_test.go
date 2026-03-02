package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"zennews/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

type fakeSummarytore struct {
	summaries    []model.NewsSummary
	total        int
	err          error
	latest       *model.NewsSummary
	latestStories []model.NewsStory
	storiesMap    map[int64][]model.NewsStory
	storiesErr   error
}

func newTestSummaryRouter(store SummaryStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSummaryHandler(store)
	r.GET("/summaries/latest", h.GetLatestSummary)
	r.GET("/summaries", h.GetSummaries)
	r.GET("/stories/latest", h.GetLatestStories)
	r.GET("/stories", h.GetStories)
	return r
}

func (f *fakeSummarytore) GetSummaries(limit int, offset int) ([]model.NewsSummary, error) {
	return f.summaries, f.err
}

func (f *fakeSummarytore) GetSummaryTotal() (int, error) {
	return f.total, f.err
}

func (f *fakeSummarytore) GetLatestSummary() (*model.NewsSummary, error) {
	return f.latest, f.err
}

func (f *fakeSummarytore) GetLatestStories() ([]model.NewsStory, error) {
	if f.storiesErr != nil {
		return nil, f.storiesErr
	}
	return f.latestStories, f.err
}

func (f *fakeSummarytore) GetStoriesBySummaryID(summaryID int64) ([]model.NewsStory, error) {
	if f.storiesErr != nil {
		return nil, f.storiesErr
	}
	if f.storiesMap != nil {
		return f.storiesMap[summaryID], nil
	}
	return nil, f.err
}

func TestGetSummaries_DBError(t *testing.T) {
	store := &fakeSummarytore{err: errors.New("DB down")}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/summaries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSummaries_Empty(t *testing.T) {
	store := &fakeSummarytore{
		summaries: []model.NewsSummary{},
		total:     0,
	}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/summaries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res SummariesResponse
	json.Unmarshal(w.Body.Bytes(), &res)

	assert.Equal(t, nil, res.Latest)
	assert.Equal(t, 0, len(res.History))
	assert.Equal(t, 0, res.Total)
}

func TestGetLatestSummary_Found(t *testing.T) {
	now := time.Now()
	store := &fakeSummarytore{
		latest: &model.NewsSummary{
			ID:           3,
			Paragraph:    "Latest summary",
			Bullets:      []string{"Event A", "Event B"},
			ArticleCount: 5,
			CreatedAt:    now,
		},
	}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/summaries/latest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res SummaryResponse
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, "Latest summary", res.Paragraph)
	assert.Equal(t, 2, len(res.Bullets))
}

func TestGetLatestSummary_NotFound(t *testing.T) {
	store := &fakeSummarytore{latest: nil}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/summaries/latest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetLatestSummary_DBError(t *testing.T) {
	store := &fakeSummarytore{err: errors.New("db down")}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/summaries/latest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetLatestStories_Found(t *testing.T) {
	store := &fakeSummarytore{
		latestStories: []model.NewsStory{
			{ID: 1, SummaryID: 5, Rank: 1, Headline: "Story A", Summary: "Summary A", Angles: []string{"angle1"}, Tickers: []string{"AAPL"}, Publishers: []string{"Reuters"}, TimeRange: "1h ago"},
			{ID: 2, SummaryID: 5, Rank: 2, Headline: "Story B", Summary: "Summary B", Angles: []string{"angle2"}, Tickers: []string{"MSFT"}, Publishers: []string{"Bloomberg"}, TimeRange: "2h ago"},
		},
	}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/stories/latest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res []StoryResponse
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, "Story A", res[0].Headline)
	assert.Equal(t, 1, res[0].Rank)
}

func TestGetLatestStories_NotFound(t *testing.T) {
	store := &fakeSummarytore{latestStories: nil}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/stories/latest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetLatestStories_DBError(t *testing.T) {
	store := &fakeSummarytore{err: errors.New("db down")}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/stories/latest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetStories_WithResults(t *testing.T) {
	now := time.Now()
	store := &fakeSummarytore{
		summaries: []model.NewsSummary{
			{ID: 3, Paragraph: "", Bullets: []string{}, ArticleCount: 5, CreatedAt: now},
		},
		total: 1,
		storiesMap: map[int64][]model.NewsStory{
			3: {
				{ID: 1, SummaryID: 3, Rank: 1, Headline: "Story A", Summary: "Summary A", Angles: []string{}, Tickers: []string{"AAPL"}, Publishers: []string{"Reuters"}},
			},
		},
	}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/stories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res struct {
		Items []struct {
			ID      int64           `json:"id"`
			Stories []StoryResponse `json:"stories"`
		} `json:"items"`
		Total int `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.Equal(t, 1, len(res.Items))
	assert.Equal(t, 1, len(res.Items[0].Stories))
	assert.Equal(t, "Story A", res.Items[0].Stories[0].Headline)
	assert.Equal(t, 1, res.Total)
}

func TestGetSummaries_WithResults(t *testing.T) {
	now := time.Now()
	store := &fakeSummarytore{
		summaries: []model.NewsSummary{
			{ID: 3,
				Paragraph:    "Latest summary",
				Bullets:      []string{"Event A", "Event B"},
				ArticleCount: 3,
				CreatedAt:    now,
			},
			{ID: 1,
				Paragraph:    "Older summary",
				Bullets:      []string{"Event A", "Event B"},
				ArticleCount: 3,
				CreatedAt:    now.Add(-24 * time.Hour),
			},
			{ID: 1,
				Paragraph:    "Oldest summary",
				Bullets:      []string{"Event A", "Event B"},
				ArticleCount: 3,
				CreatedAt:    now.Add(-48 * time.Hour),
			},
		},
		total: 3,
	}

	r := newTestSummaryRouter(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/summaries", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res SummariesResponse
	json.Unmarshal(w.Body.Bytes(), &res)

	assert.NotEqual(t, nil, res.Latest)
	assert.Equal(t, "Latest summary", res.Latest.Paragraph)
	assert.Equal(t, 2, len(res.Latest.Bullets))

	assert.Equal(t, 2, len(res.History))
	assert.Equal(t, "Older summary", res.History[0].Paragraph)
	assert.Equal(t, 3, res.Total)
}
