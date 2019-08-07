CREATE TABLE IF NOT EXISTS books (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    author VARCHAR(100),
    pub_year INTEGER,
    page_count INTEGER
);
