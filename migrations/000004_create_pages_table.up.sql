CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY,
    page_number INT,
    body TEXT,
    book UUID REFERENCES books (id)
);
