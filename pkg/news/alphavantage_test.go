package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestGenerateExternalID(t *testing.T) {
	url := "https://example.com/article/123"

	id1 := generateExternalID(url)
	id2 := generateExternalID(url)

	assert.Equal(t, id1, id2)
	assert.Equal(t, 16, len(id1))

	other := generateExternalID("https://example.com/article/456")
	assert.NotEqual(t, id1, other)
}

func TestParseTimePublished(t *testing.T) {
	input := "20260226T075324"
	got, err := time.Parse("20060102T150405", input)

	assert.Equal(t, nil, err)
	assert.Equal(t, 2026, got.Year())
	assert.Equal(t, time.February, got.Month())
	assert.Equal(t, 26, got.Day())
	assert.Equal(t, 7, got.Hour())
	assert.Equal(t, 53, got.Minute())
	assert.Equal(t, 24, got.Second())
}

func TestFetch(t *testing.T) {
	payload := map[string]interface{}{
		"feed": []map[string]interface{}{
			{
				"title":          "Fed Holds Rates Steady",
				"summary":        "The Federal Reserve kept interest rates unchanged.",
				"url":            "https://example.com/fed-rates",
				"source":         "Reuters",
				"time_published": "20260226T120000",
				"ticker_sentiment": []map[string]interface{}{
					{"ticker": "SPY"},
					{"ticker": "TLT"},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	client := &AlphaVantageClient{
		apiKey:     "test-key",
		httpClient: srv.Client(),
	}
	client.httpClient.Transport = &rewriteTransport{base: srv.URL, inner: http.DefaultTransport}

	articles, err := client.Fetch(1)

	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(articles))

	a := articles[0]
	assert.Equal(t, "Fed Holds Rates Steady", a.Headline)
	assert.Equal(t, "The Federal Reserve kept interest rates unchanged.", a.Detail)
	assert.Equal(t, "https://example.com/fed-rates", a.URL)
	assert.Equal(t, "Reuters", a.Publisher)
	assert.Equal(t, "AlphaVantage", a.Source)
	assert.Equal(t, []string{"SPY", "TLT"}, a.Symbols)
	assert.Equal(t, generateExternalID("https://example.com/fed-rates"), a.ExternalID)
	assert.NotEqual(t, time.Time{}, a.PublishedAt)
}

// rewriteTransport redirects all requests to a fixed base URL (test server).
type rewriteTransport struct {
	base  string
	inner http.RoundTripper
}

func (rt *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	parsed, _ := http.NewRequest("GET", rt.base, nil)
	req2.URL.Host = parsed.URL.Host
	req2.URL.Scheme = parsed.URL.Scheme
	return rt.inner.RoundTrip(req2)
}
