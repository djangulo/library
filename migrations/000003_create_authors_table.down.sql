BEGIN;

DROP INDEX IF EXISTS index_authors_on_created_at_name RESTRICT;
DROP INDEX IF EXISTS index_authors_on_created_at_id RESTRICT;
DROP TABLE IF EXISTS authors;

COMMIT;
