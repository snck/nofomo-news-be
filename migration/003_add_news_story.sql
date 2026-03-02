CREATE TABLE news_story (
    id SERIAL PRIMARY KEY,
    summary_id INTEGER NOT NULL REFERENCES news_summary(id) ON DELETE CASCADE,
    rank INTEGER NOT NULL,
    headline TEXT NOT NULL,
    summary TEXT NOT NULL,
    angles JSONB NOT NULL DEFAULT '[]',
    tickers JSONB NOT NULL DEFAULT '[]',
    publishers JSONB NOT NULL DEFAULT '[]',
    time_range VARCHAR(100)
);
CREATE INDEX idx_news_story_summary_id ON news_story(summary_id);
