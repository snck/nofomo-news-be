package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"
	"zennews/db"
	"zennews/internal/model"
	"zennews/internal/repository"
	"zennews/pkg/llm"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	const maxRetries = 3

	err := db.ConnectRedis()
	if err != nil {
		log.Fatalf("error connecting to Redis: %v", err)
	}
	defer db.CloseRedis()

	err = db.Connect()
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}
	defer db.Close()

	articleRepository := repository.NewArticleRepository(db.DB)

	openAIClient := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))

	for {
		id, err := db.PopFromQueue(db.TransformQueueKey)
		if err != nil {
			slog.Error("error popping from Redis queue", "error", err)
			break
		}

		articleId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			slog.Error("invalid article id in queue", "id", id, "error", err)
			continue
		}

		errorCount, err := articleRepository.GetErrorCount(articleId)
		if err != nil {
			slog.Error("error getting error count", "error", err, "article_id", articleId)
			continue
		}

		if errorCount >= maxRetries {
			slog.Warn("article exceeded max retries, marking as failed", "article_id", articleId, "error_count", errorCount)
			articleRepository.UpdateStatus(articleId, model.StatusFailed)
			continue
		}

		article, err := articleRepository.GetOriginalByID(articleId)
		if err != nil {
			slog.Error("error getting article from DB", "error", err, "article_id", articleId)
			continue
		}

		if article == nil {
			slog.Warn("article not found in DB", "article_id", articleId)
			continue
		}

		input := llm.TransformInput{
			Headline: article.Headline,
			Detail:   article.Detail,
		}

		result, err := openAIClient.Transform(input)
		if err != nil {
			slog.Error("error transforming article", "error", err, "article_id", articleId)

			articleRepository.SaveError(articleId, err.Error(), "llm_error")

			db.PushToQueue(db.TransformQueueKey, strconv.FormatInt(articleId, 10))

			time.Sleep(5 * time.Second)
			continue
		}

		category, err := articleRepository.GetCategoryByName(result.Category)
		if err != nil {
			slog.Error("error getting category", "error", err, "category", result.Category)
		}

		if category == nil {
			slog.Warn("LLM returned unknown category, falling back to Others", "category", result.Category, "article_id", articleId)
			category, err = articleRepository.GetCategoryByName(model.OthersCategory)

			if err != nil {
				slog.Error("error getting Others category", "error", err, "article_id", articleId)
				continue
			}
		}

		transformedArticle := model.TransformedArticle{
			Headline:       result.Headline,
			Detail:         result.Detail,
			OriginalID:     article.ID,
			CategoryID:     category.ID,
			SentimentScore: result.SentimentScore,
			PromptVersion:  result.PromptVersion,
			ModelUsed:      result.ModelUsed,
			TransformedAt:  time.Now(),
		}

		err = articleRepository.SaveTransformedAndComplete(&transformedArticle, article.ID)
		if err != nil {
			slog.Error("error saving transformed article", "error", err, "article_id", articleId)
			continue
		}

		slog.Info("article transformed successfully", "article_id", article.ID)
	}

}
