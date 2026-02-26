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
	summaries []model.NewsSummary
	total     int
	err       error
}

func newTestSummaryRouter(store SummaryStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSummaryHandler(store)
	r.GET("/summaries", h.GetSummaries)
	return r
}

func (f *fakeSummarytore) GetSummaries(limit int, offset int) ([]model.NewsSummary, error) {
	return f.summaries, f.err
}

func (f *fakeSummarytore) GetSummaryTotal() (int, error) {
	return f.total, f.err
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
