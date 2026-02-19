CREATE TABLE original_article (
    id SERIAL PRIMARY KEY,
    headline TEXT NOT NULL,
    detail TEXT,
    url TEXT NOT NULL UNIQUE,
    source VARCHAR(50) NOT NULL,
    publisher VARCHAR(255),
    published_at TIMESTAMP,
    fetched_at TIMESTAMP DEFAULT NOW(),
    external_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending'
);

CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE transformed_article (
    id SERIAL PRIMARY KEY,
    headline TEXT NOT NULL,
    detail TEXT,
    original_id INTEGER REFERENCES original_article(id),
    category_id INTEGER REFERENCES category(id),
    sentiment_score INTEGER,
    prompt_version VARCHAR(20),
    model_used VARCHAR(50),
    transformed_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE article_symbol (
    id SERIAL PRIMARY KEY,
    article_id INTEGER REFERENCES original_article(id),
    symbol VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE processing_error (
    id SERIAL PRIMARY KEY,
    article_id INTEGER REFERENCES original_article(id),
    error_message TEXT,
    error_type VARCHAR(50),
    attempt_count INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_original_url ON original_article(url);
CREATE INDEX idx_original_status ON original_article(status);
CREATE INDEX idx_original_fetched ON original_article(fetched_at);
CREATE INDEX idx_transformed_date ON transformed_article(transformed_at);
CREATE INDEX idx_article_symbol ON article_symbol(symbol);
CREATE INDEX idx_article_symbol_article_id ON article_symbol(article_id);

-- Seed categories
INSERT INTO category (name) VALUES 
    ('Earnings'),
    ('Market Movement'),
    ('Economy'),
    ('Crypto'),
    ('Mergers & Acquisitions'),
    ('Policy & Regulation'),
    ('Company News'),
    ('Analysis'),
    ('Others');