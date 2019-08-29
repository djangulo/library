BEGIN;

CREATE TABLE IF NOT EXISTS authors (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE
) INHERITS (stamps);

-- id index is created by default, skipping
CREATE UNIQUE INDEX index_authors_on_created_at_id
ON authors
USING btree (created_at, id)
INCLUDE (name, slug);

CREATE INDEX index_authors_on_created_at_name
ON authors
USING btree (created_at, name);

COMMIT;
