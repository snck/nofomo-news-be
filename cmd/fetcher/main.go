package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"
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

	err = db.ConnectRedis()
	if err != nil {
		log.Fatalf("error connecting to Redis: %v", err)
	}
	defer db.CloseRedis()

	var clients []news.NewsClient
	if key := os.Getenv("FINNHUB_API_KEY"); key != "" {
		clients = append(clients, news.NewFinnHubClient(key))
	}
	if key := os.Getenv("ALPHA_VANTAGE_API_KEY"); key != "" {
		clients = append(clients, news.NewAlphaVantageClient(key))
	}

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

			err = db.PushToQueue(db.TransformQueueKey, strconv.FormatInt(article.ID, 10))
			if err != nil {
				slog.Error("error pushing to Redis queue", "source", source, "error", err, "article_id", article.ID)
				errors++
			}
		}

		slog.Info("fetch complete", "source", source, "saved", saved, "duplicated", duplicated, "errors", errors)
	}
}
