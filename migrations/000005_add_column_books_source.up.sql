BEGIN;

CREATE TYPE enum_source AS ENUM (
    'nltk-gutenberg',
    'open-library'
);
ALTER TABLE books ADD COLUMN source enum_source;

COMMIT;
