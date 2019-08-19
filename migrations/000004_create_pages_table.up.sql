BEGIN;

CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY,
    page_number INT,
    body TEXT,
    book_id UUID REFERENCES books (id)
);
CREATE UNIQUE INDEX page_book_idx ON pages (page_number, book_id);

COMMIT;
