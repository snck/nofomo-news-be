CREATE TABLE news_summary (
    id SERIAL PRIMARY KEY,
    paragraph TEXT NOT NULL,
    bullets JSONB NOT NULL,
    article_count INTEGER NOT NULL,
    from_article_id INTEGER REFERENCES original_article(id),
    to_article_id INTEGER REFERENCES original_article(id),
    model_used VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_news_summary_created_at ON news_summary(created_at);
