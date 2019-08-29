
BEGIN;

DROP INDEX IF EXISTS index_pages_on_book_id_page_number RESTRICT;
DROP INDEX IF EXISTS index_pages_on_created_at_id RESTRICT;
DROP TABLE IF EXISTS pages;

COMMIT;
