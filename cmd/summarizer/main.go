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

	articleRepo := repository.NewArticleRepository(db.DB)
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

	// Batch-load symbols for all articles
	articleIDs := make([]int64, len(articles))
	for i, a := range articles {
		articleIDs[i] = a.ID
	}
	symbolsMap, err := articleRepo.GetSymbolsByOriginalIDs(articleIDs)
	if err != nil {
		log.Fatalf("error fetching symbols: %v", err)
	}

	inputs := make([]llm.SummaryInput, len(articles))
	for i, a := range articles {
		inputs[i] = llm.SummaryInput{
			ID:          a.ID,
			Headline:    a.Headline,
			Detail:      a.Detail,
			Publisher:   a.Publisher,
			PublishedAt: a.PublishedAt,
			Symbols:     symbolsMap[a.ID],
		}
	}

	result, err := openAIClient.ClusterAndSummarize(inputs)
	if err != nil {
		log.Fatalf("error generating cluster summary: %v", err)
	}

	summary := &model.NewsSummary{
		Paragraph:     "",
		Bullets:       []string{},
		ArticleCount:  len(articles),
		FromArticleID: articles[0].ID,
		ToArticleID:   articles[len(articles)-1].ID,
		ModelUsed:     result.ModelUsed,
	}

	err = summaryRepo.SaveSummary(summary)
	if err != nil {
		log.Fatalf("error saving summary: %v", err)
	}

	stories := make([]model.NewsStory, len(result.Stories))
	for i, s := range result.Stories {
		stories[i] = model.NewsStory{
			Headline:   s.Headline,
			Summary:    s.Summary,
			Angles:     s.Angles,
			Tickers:    s.Tickers,
			Publishers: s.Publishers,
			TimeRange:  s.TimeRange,
		}
	}

	err = summaryRepo.SaveStories(summary.ID, stories)
	if err != nil {
		log.Fatalf("error saving stories: %v", err)
	}

	slog.Info("summary saved successfully", "summary_id", summary.ID, "article_count", summary.ArticleCount, "story_count", len(stories))
}
