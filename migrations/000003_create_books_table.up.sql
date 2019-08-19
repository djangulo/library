CREATE TABLE IF NOT EXISTS books (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,
    publication_year INTEGER,
    page_count INTEGER,
    file VARCHAR(255),
    author_id UUID REFERENCES authors (id)
);
