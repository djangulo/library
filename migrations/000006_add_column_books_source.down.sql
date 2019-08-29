BEGIN;

ALTER TABLE books DROP COLUMN IF EXISTS source;
DROP TYPE enum_source;

COMMIT;
