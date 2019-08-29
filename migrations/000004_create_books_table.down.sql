BEGIN;

DROP INDEX IF EXISTS index_books_on_created_at_slug RESTRICT;
DROP INDEX IF EXISTS index_books_on_created_at_id RESTRICT;
DROP TABLE IF EXISTS books;

COMMIT;
