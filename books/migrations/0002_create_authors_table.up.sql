CREATE TABLE IF NOT EXISTS authors (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE,
    slug VARCHAR(100) UNIQUE,
    
)


CREATE TABLE IF NOT EXISTS books (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    author VARCHAR(100),
    pub_year INTEGER,
    page_count INTEGER,
    file VARCHAR(255)
);
