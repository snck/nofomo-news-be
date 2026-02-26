package main

import (
	"log"
	"log/slog"
	"os"
	"zennews/db"
	"zennews/internal/model"
	"zennews/internal/repository"
	"zennews/pkg/llm"

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

	summaryRepo := repository.NewSummaryRepository(db.DB)
	openAIClient := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))

	fromID, err := summaryRepo.GetLastToArticleID()
	if err != nil {
		log.Fatalf("error getting last summary article id: %v", err)
	}

	articles, err := summaryRepo.GetArticlesForSummary(fromID)
	if err != nil {
		log.Fatalf("error fetching articles for summary: %v", err)
	}

	if len(articles) == 0 {
		slog.Info("no new articles to summarize, exiting")
		return
	}

	slog.Info("summarizing articles", "count", len(articles), "from_id", fromID)

	inputs := make([]llm.SummaryInput, len(articles))
	for i, a := range articles {
		inputs[i] = llm.SummaryInput{
			Headline: a.Headline,
			Detail:   a.Detail,
		}
	}

	result, err := openAIClient.Summarize(inputs)
	if err != nil {
		log.Fatalf("error generating summary: %v", err)
	}

	summary := &model.NewsSummary{
		Paragraph:     result.Paragraph,
		Bullets:       result.Bullets,
		ArticleCount:  len(articles),
		FromArticleID: articles[0].ID,
		ToArticleID:   articles[len(articles)-1].ID,
		ModelUsed:     result.ModelUsed,
	}

	err = summaryRepo.SaveSummary(summary)
	if err != nil {
		log.Fatalf("error saving summary: %v", err)
	}

	slog.Info("summary saved successfully", "summary_id", summary.ID, "article_count", summary.ArticleCount)
}
