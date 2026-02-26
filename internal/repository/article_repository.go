package repository

import (
	"database/sql"
	"zennews/internal/model"

	"github.com/lib/pq"
)

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) SaveOriginal(article *model.OriginalArticle) (bool, error) {
	var id int64
	err := r.db.QueryRow(`
		INSERT INTO original_article(headline, detail, url, source, publisher, published_at, external_id, status) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8) 
		ON CONFLICT (url) DO NOTHING
		RETURNING id
	`, article.Headline, article.Detail, article.URL, article.Source, article.Publisher, article.PublishedAt, article.ExternalID, model.StatusPending).Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	article.ID = id
	return true, nil
}

func (r *ArticleRepository) GetPendingArticle(limit int) ([]model.OriginalArticle, error) {
	rows, err := r.db.Query(`
		SELECT id, headline, detail, url, source, published_at, fetched_at, external_id, status 
		FROM original_article 
		WHERE status = $1
		ORDER by fetched_at ASC 
		LIMIT $2
	`, model.StatusPending, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.OriginalArticle
	for rows.Next() {
		var a model.OriginalArticle
		err := rows.Scan(&a.ID, &a.Headline, &a.Detail, &a.URL, &a.Source, &a.PublishedAt, &a.FetchedAt, &a.ExternalID, &a.Status)
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

func (r *ArticleRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`
		UPDATE original_article SET status = $1 WHERE id = $2
	`, status, id)
	return err
}

func (r *ArticleRepository) GetOriginalByID(id int64) (*model.OriginalArticle, error) {
	var a model.OriginalArticle
	err := r.db.QueryRow(`
		SELECT id, headline, detail, url, source, published_at, fetched_at, external_id, status 
		FROM original_article 
		WHERE id = $1
	`, id).Scan(&a.ID, &a.Headline, &a.Detail, &a.URL, &a.Source, &a.PublishedAt, &a.FetchedAt, &a.ExternalID, &a.Status)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *ArticleRepository) SaveTransformed(article *model.TransformedArticle) error {
	return r.db.QueryRow(`
		INSERT INTO transformed_article(headline, detail, original_id, category_id, sentiment_score, prompt_version, model_used)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, article.Headline, article.Detail, article.OriginalID, article.CategoryID, article.SentimentScore, article.PromptVersion, article.ModelUsed).Scan(&article.ID)
}

func (r *ArticleRepository) SaveTransformedAndComplete(article *model.TransformedArticle, originalID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(`
		INSERT INTO transformed_article(headline, detail, original_id, category_id, sentiment_score, prompt_version, model_used)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, article.Headline, article.Detail, article.OriginalID, article.CategoryID, article.SentimentScore, article.PromptVersion, article.ModelUsed).Scan(&article.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE original_article SET status = $1 WHERE id = $2
	`, model.StatusCompleted, originalID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ArticleRepository) GetTransformedFeed(limit int, offset int) ([]model.TransformedArticle, error) {
	rows, err := r.db.Query(`
		SELECT id, headline, detail, original_id, category_id, sentiment_score, prompt_version, model_used, transformed_at 
		FROM transformed_article 
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.TransformedArticle
	for rows.Next() {
		var a model.TransformedArticle
		err := rows.Scan(&a.ID, &a.Headline, &a.Detail, &a.OriginalID, &a.CategoryID, &a.SentimentScore, &a.PromptVersion, &a.ModelUsed, &a.TransformedAt)
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

func (r *ArticleRepository) SaveOriginalWithSymbols(article *model.OriginalArticle, symbols []string) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRow(`
		INSERT INTO original_article(headline, detail, url, source, publisher, published_at, external_id, status)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (url) DO NOTHING
		RETURNING id
	`, article.Headline, article.Detail, article.URL, article.Source, article.Publisher, article.PublishedAt, article.ExternalID, model.StatusPending).Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	article.ID = id

	if len(symbols) > 0 {
		_, err = tx.Exec(`
			INSERT INTO article_symbol(article_id, symbol)
			SELECT $1, unnest($2::text[])
		`, id, pq.Array(symbols))
		if err != nil {
			return false, err
		}
	}

	return true, tx.Commit()
}

func (r *ArticleRepository) SaveSymbols(articleID int64, symbols []string) error {
	_, err := r.db.Exec(`
		INSERT INTO article_symbol(article_id, symbol)
		SELECT $1, unnest($2::text[])
	`, articleID, pq.Array(symbols))
	return err
}

func (r *ArticleRepository) SaveError(articleID int64, errMsg string, errType string) error {
	_, err := r.db.Exec(`
		INSERT INTO processing_error(article_id, error_message, error_type) 
		VALUES($1, $2, $3)
	`, articleID, errMsg, errType)

	return err
}

func (r *ArticleRepository) GetCategoryByName(name string) (*model.Category, error) {
	var category model.Category
	err := r.db.QueryRow(`
		SELECT id, name 
		FROM category 
		WHERE name = $1
	`, name).Scan(&category.ID, &category.Name)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *ArticleRepository) GetAllCategories() ([]model.Category, error) {
	rows, err := r.db.Query(`
		SELECT id, name 
		FROM category 
		ORDER BY name
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var c model.Category
		err := rows.Scan(&c.ID, &c.Name)
		if err != nil {
			return nil, err
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *ArticleRepository) GetFeed(limit int, offset int) ([]model.FeedArticle, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.headline, t.detail, t.sentiment_score, 
			o.publisher, o.published_at, o.url, 
			c.id, c.name,
			o.id
		FROM transformed_article t 
		JOIN original_article o ON o.id = t.original_id 
		JOIN category c on c.id = t.category_id 
		ORDER BY o.published_at DESC 
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.FeedArticle
	for rows.Next() {
		var a model.FeedArticle
		err = rows.Scan(
			&a.ID, &a.Headline, &a.Detail, &a.SentimentScore,
			&a.Publisher, &a.PublishedAt, &a.URL, &a.CategoryID,
			&a.CategoryName, &a.OriginalID,
		)

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

func (r *ArticleRepository) GetFeedTotal() (int, error) {
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM transformed_article
	`).Scan(&total)
	return total, err
}

func (r *ArticleRepository) GetSymbolsByOriginalID(id int64) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT symbol FROM article_symbol 
		WHERE article_id = $1
	`, id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return symbols, nil
}

func (r *ArticleRepository) GetTransformedByID(id int64) (*model.SingleArticle, error) {

	var a model.SingleArticle
	err := r.db.QueryRow(`
		SELECT t.id, t.headline, t.detail, t.sentiment_score, 
			o.publisher, o.published_at, o.url, 
			c.id, c.name,
			o.id, o.headline, o.detail 
		FROM transformed_article t 
		JOIN original_article o ON o.id = t.original_id 
		JOIN category c on c.id = t.category_id 
		WHERE t.id = $1
	`, id).Scan(&a.ID, &a.Headline, &a.Detail, &a.SentimentScore, &a.Publisher, &a.PublishedAt,
		&a.URL, &a.CategoryID, &a.CategoryName, &a.OriginalID, &a.OriginalHeadline, &a.OriginalDetail)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *ArticleRepository) GetSymbolsByOriginalIDs(ids []int64) (map[int64][]string, error) {
	rows, err := r.db.Query(`
		SELECT article_id, symbol FROM article_symbol WHERE article_id = ANY($1)
	`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]string)
	for rows.Next() {
		var id int64
		var symbol string
		if err := rows.Scan(&id, &symbol); err != nil {
			return nil, err
		}
		result[id] = append(result[id], symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *ArticleRepository) GetOriginalFeed(limit, offset int) ([]model.OriginalArticle, error) {
	rows, err := r.db.Query(`
		SELECT id, headline, detail, url, source, publisher, published_at, fetched_at, external_id, status
		FROM original_article
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.OriginalArticle
	for rows.Next() {
		var a model.OriginalArticle
		err := rows.Scan(&a.ID, &a.Headline, &a.Detail, &a.URL, &a.Source, &a.Publisher, &a.PublishedAt, &a.FetchedAt, &a.ExternalID, &a.Status)
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

func (r *ArticleRepository) GetOriginalFeedTotal() (int, error) {
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM original_article
	`).Scan(&total)
	return total, err
}

func (r *ArticleRepository) GetErrorCount(id int64) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM processing_error 
		WHERE article_id = $1
	`, id).Scan(&count)

	return count, err
}
