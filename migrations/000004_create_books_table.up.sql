BEGIN;

CREATE TABLE IF NOT EXISTS books (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,
    publication_year INTEGER,
    page_count INTEGER,
    file VARCHAR(255),
    author_id UUID REFERENCES authors (id)
) INHERITS (stamps);

-- id index is created by default, skipping
CREATE UNIQUE INDEX index_books_on_created_at_id
ON books
USING btree (created_at, id);

CREATE INDEX index_books_on_created_at_slug
ON books
USING btree (created_at, slug);

COMMIT;
