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

	client := news.NewFinnHubClient(os.Getenv("FINNHUB_API_KEY"))

	fetchedArticles, err := client.Fetch(10)
	if err != nil {
		slog.Error("error fetching from FinnHub", "error", err)
		return
	}

	repository := repository.NewArticleRepository(db.DB)

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

		success, err := repository.SaveOriginalWithSymbols(&article, a.Symbols)

		if err != nil {
			slog.Error("error saving article", "error", err)
			errors++
			continue
		}

		if !success {
			slog.Info("duplicate article skipped", "url", a.URL)
			duplicated++
			continue
		}

		saved++

		err = db.PushToQueue(db.TransformQueueKey, strconv.FormatInt(article.ID, 10))
		if err != nil {
			slog.Error("error pushing to Redis queue", "error", err, "article_id", article.ID)
			errors++
		}

	}

	slog.Info("fetch complete", "saved", saved, "duplicated", duplicated, "errors", errors)
}
