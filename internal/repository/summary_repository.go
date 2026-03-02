package repository

import (
	"database/sql"
	"encoding/json"
	"zennews/internal/model"
)

type SummaryRepository struct {
	db *sql.DB
}

func NewSummaryRepository(db *sql.DB) *SummaryRepository {
	return &SummaryRepository{db: db}
}

func (r *SummaryRepository) GetLastToArticleID() (int64, error) {
	var id int64
	err := r.db.QueryRow(`
		SELECT COALESCE(MAX(to_article_id), 0) FROM news_summary
	`).Scan(&id)
	return id, err
}

func (r *SummaryRepository) GetArticlesForSummary(fromID int64) ([]model.OriginalArticle, error) {
	rows, err := r.db.Query(`
		SELECT id, headline, detail, url, source, publisher, published_at, external_id
		FROM original_article
		WHERE id > $1
		ORDER BY id ASC
	`, fromID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.OriginalArticle
	for rows.Next() {
		var a model.OriginalArticle
		err := rows.Scan(&a.ID, &a.Headline, &a.Detail, &a.URL, &a.Source, &a.Publisher, &a.PublishedAt, &a.ExternalID)
		if err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *SummaryRepository) SaveSummary(summary *model.NewsSummary) error {
	bullets, err := json.Marshal(summary.Bullets)
	if err != nil {
		return err
	}

	return r.db.QueryRow(`
		INSERT INTO news_summary(paragraph, bullets, article_count, from_article_id, to_article_id, model_used)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, summary.Paragraph, bullets, summary.ArticleCount, summary.FromArticleID, summary.ToArticleID, summary.ModelUsed).Scan(&summary.ID)
}

func (r *SummaryRepository) GetSummaries(limit, offset int) ([]model.NewsSummary, error) {
	rows, err := r.db.Query(`
		SELECT id, paragraph, bullets, article_count, from_article_id, to_article_id, model_used, created_at
		FROM news_summary
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []model.NewsSummary
	for rows.Next() {
		var s model.NewsSummary
		var bulletsJSON []byte
		err := rows.Scan(&s.ID, &s.Paragraph, &bulletsJSON, &s.ArticleCount, &s.FromArticleID, &s.ToArticleID, &s.ModelUsed, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bulletsJSON, &s.Bullets); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *SummaryRepository) GetSummaryTotal() (int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM news_summary`).Scan(&total)
	return total, err
}

func (r *SummaryRepository) SaveStories(summaryID int64, stories []model.NewsStory) error {
	for i, s := range stories {
		anglesJSON, err := json.Marshal(s.Angles)
		if err != nil {
			return err
		}
		tickersJSON, err := json.Marshal(s.Tickers)
		if err != nil {
			return err
		}
		publishersJSON, err := json.Marshal(s.Publishers)
		if err != nil {
			return err
		}
		_, err = r.db.Exec(`
			INSERT INTO news_story(summary_id, rank, headline, summary, angles, tickers, publishers, time_range)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		`, summaryID, i+1, s.Headline, s.Summary, anglesJSON, tickersJSON, publishersJSON, s.TimeRange)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SummaryRepository) GetStoriesBySummaryID(summaryID int64) ([]model.NewsStory, error) {
	rows, err := r.db.Query(`
		SELECT id, summary_id, rank, headline, summary, angles, tickers, publishers, time_range
		FROM news_story
		WHERE summary_id = $1
		ORDER BY rank ASC
	`, summaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStories(rows)
}

func (r *SummaryRepository) GetLatestStories() ([]model.NewsStory, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.summary_id, s.rank, s.headline, s.summary, s.angles, s.tickers, s.publishers, s.time_range
		FROM news_story s
		INNER JOIN news_summary ns ON ns.id = s.summary_id
		WHERE ns.id = (SELECT id FROM news_summary ORDER BY created_at DESC LIMIT 1)
		ORDER BY s.rank ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanStories(rows)
}

func scanStories(rows *sql.Rows) ([]model.NewsStory, error) {
	var stories []model.NewsStory
	for rows.Next() {
		var s model.NewsStory
		var anglesJSON, tickersJSON, publishersJSON []byte
		var timeRange sql.NullString
		err := rows.Scan(&s.ID, &s.SummaryID, &s.Rank, &s.Headline, &s.Summary,
			&anglesJSON, &tickersJSON, &publishersJSON, &timeRange)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(anglesJSON, &s.Angles); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(tickersJSON, &s.Tickers); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(publishersJSON, &s.Publishers); err != nil {
			return nil, err
		}
		if timeRange.Valid {
			s.TimeRange = timeRange.String
		}
		stories = append(stories, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stories, nil
}

func (r *SummaryRepository) GetLatestSummary() (*model.NewsSummary, error) {
	var s model.NewsSummary
	var bulletsJSON []byte
	err := r.db.QueryRow(`
		SELECT id, paragraph, bullets, article_count, from_article_id, to_article_id, model_used, created_at
		FROM news_summary
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&s.ID, &s.Paragraph, &bulletsJSON, &s.ArticleCount, &s.FromArticleID, &s.ToArticleID, &s.ModelUsed, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bulletsJSON, &s.Bullets); err != nil {
		return nil, err
	}

	return &s, nil
}
