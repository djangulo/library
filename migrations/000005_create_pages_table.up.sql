BEGIN;

CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY,
    page_number INT,
    body TEXT,
    book_id UUID REFERENCES books (id)
) INHERITS (stamps);

-- id index is created by default, skipping
CREATE UNIQUE INDEX index_pages_on_created_at_id
ON pages
USING btree (created_at, id);

CREATE UNIQUE INDEX index_pages_on_book_id_page_number
ON pages
USING btree (book_id, page_number);

COMMIT;
