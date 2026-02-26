package main

import (
	"log"
	"log/slog"
	"os"
	"zennews/db"
	"zennews/internal/handler"
	"zennews/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	err := db.Connect()
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}
	defer db.Close()

	articleRepo := repository.NewArticleRepository(db.DB)
	articleHandler := handler.NewArticleHandler(articleRepo)

	summaryRepo := repository.NewSummaryRepository(db.DB)
	summaryHandler := handler.NewSummaryHandler(summaryRepo)

	r := gin.Default()

	allowedOrigins := []string{"http://localhost:3000"}

	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}

	slog.Info("AllowOrigins URL:", "urls", allowedOrigins)

	r.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	r.GET("/feed/:id", articleHandler.GetArticle)
	r.GET("/feed", articleHandler.GetFeed)
	r.GET("/articles", articleHandler.GetOriginalFeed)
	r.GET("/categories", articleHandler.GetCategories)
	r.GET("/summaries/latest", summaryHandler.GetLatestSummary)
	r.GET("/summaries", summaryHandler.GetSummaries)
	r.GET("/health", articleHandler.GetHealth)

	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
