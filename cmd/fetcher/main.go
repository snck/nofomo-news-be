package main

import (
	"log"
	"log/slog"
	"os"
	"zennews/db"
	"zennews/internal/model"
	"zennews/internal/repository"
	"zennews/pkg/news"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	err := db.Connect()
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}
	defer db.Close()

	var clients []news.NewsClient
	if key := os.Getenv("FINNHUB_API_KEY"); key != "" {
		clients = append(clients, news.NewFinnHubClient(key))
	}
	if key := os.Getenv("ALPHA_VANTAGE_API_KEY"); key != "" {
		clients = append(clients, news.NewAlphaVantageClient(key))
	}
	if key := os.Getenv("MASSIVE_API_KEY"); key != "" {
		clients = append(clients, news.NewMassiveClient(key))
	}
	if key := os.Getenv("MARKETAUX_API_KEY"); key != "" {
		clients = append(clients, news.NewMarketauxClient(key, 4))
	}

	// RSS feeds (no API key required)
	clients = append(clients,
		news.NewRSSClient("https://search.cnbc.com/rs/search/combinedcms/view.xml?partnerId=wrss01&id=100727362", "CNBC"),
		news.NewRSSClient("https://search.cnbc.com/rs/search/combinedcms/view.xml?partnerId=wrss01&id=15837362", "CNBC"),
	)

	if len(clients) == 0 {
		slog.Error("no news source API keys configured")
		return
	}

	repo := repository.NewArticleRepository(db.DB)

	for _, client := range clients {
		source := client.Name()

		fetchedArticles, err := client.Fetch(50)
		if err != nil {
			slog.Error("error fetching articles", "source", source, "error", err)
			continue
		}

		var saved, duplicated, errors int

		for _, a := range fetchedArticles {
			article := model.OriginalArticle{
				Headline:    a.Headline,
				Detail:      a.Detail,
				URL:         a.URL,
				Source:      a.Source,
				Publisher:   a.Publisher,
				PublishedAt: a.PublishedAt,
				ExternalID:  a.ExternalID,
			}

			success, err := repo.SaveOriginalWithSymbols(&article, a.Symbols)
			if err != nil {
				slog.Error("error saving article", "source", source, "error", err)
				errors++
				continue
			}

			if !success {
				slog.Info("duplicate article skipped", "source", source, "url", a.URL)
				duplicated++
				continue
			}

			saved++
		}

		slog.Info("fetch complete", "source", source, "saved", saved, "duplicated", duplicated, "errors", errors)
	}
}
