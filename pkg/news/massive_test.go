package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestMassiveFetch(t *testing.T) {
	payload := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"id":            "576d99da",
				"title":         "Acme Corp Reports Q4 Earnings",
				"description":   "Acme Corp beat expectations with strong Q4 results.",
				"article_url":   "https://example.com/acme-q4",
				"published_utc": "2026-02-26T11:02:00Z",
				"tickers":       []string{"ACME", "SPY"},
				"publisher": map[string]interface{}{
					"name": "GlobeNewswire Inc.",
				},
			},
		},
		"status": "OK",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	client := &MassiveClient{
		apiKey:     "test-key",
		httpClient: srv.Client(),
	}
	client.httpClient.Transport = &rewriteTransport{base: srv.URL, inner: http.DefaultTransport}

	articles, err := client.Fetch(1)

	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(articles))

	a := articles[0]
	assert.Equal(t, "576d99da", a.ExternalID)
	assert.Equal(t, "Acme Corp Reports Q4 Earnings", a.Headline)
	assert.Equal(t, "Acme Corp beat expectations with strong Q4 results.", a.Detail)
	assert.Equal(t, "https://example.com/acme-q4", a.URL)
	assert.Equal(t, "GlobeNewswire Inc.", a.Publisher)
	assert.Equal(t, "Massive", a.Source)
	assert.Equal(t, []string{"ACME", "SPY"}, a.Symbols)
	assert.NotEqual(t, time.Time{}, a.PublishedAt)
	assert.Equal(t, 2026, a.PublishedAt.Year())
	assert.Equal(t, time.February, a.PublishedAt.Month())
	assert.Equal(t, 26, a.PublishedAt.Day())
}

func TestMassiveFetchEmptyTickers(t *testing.T) {
	payload := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"id":            "abc123",
				"title":         "Market Update",
				"description":   "General market overview.",
				"article_url":   "https://example.com/market",
				"published_utc": "2026-02-26T10:00:00Z",
				"tickers":       []string{},
				"publisher": map[string]interface{}{
					"name": "Reuters",
				},
			},
		},
		"status": "OK",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	client := &MassiveClient{
		apiKey:     "test-key",
		httpClient: srv.Client(),
	}
	client.httpClient.Transport = &rewriteTransport{base: srv.URL, inner: http.DefaultTransport}

	articles, err := client.Fetch(1)

	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(articles))
	assert.Equal(t, 0, len(articles[0].Symbols))
}
