package books

import (
	"database/sql"
	"fmt"
	"github.com/djangulo/library/config"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // unneded namespace
	"github.com/pkg/errors"
	"log"
)

// SQLStore houses the PostgreSQL connection
type SQLStore struct {
	DB *sqlx.DB
}

type Book struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	Title           string         `json:"title" db:"title"`
	Slug            string         `json:"slug" db:"slug"`
	Author          sql.NullString `json:"author" db:"author"`
	PublicationYear sql.NullInt64  `json:"publication_year" db:"publication_year"`
	PageCount       int            `json:"page_count" db:"page_count"`
	Pages           []Page         `json:"pages" db:"pages"`
}

type Author struct {
	ID    uuid.UUID `json:"id" db:"id"`
	Name  string    `json:"name" db:"name"`
	Slug  string    `json:"slug" db:"slug"`
	Books []Book    `json:"books" db:"books"`
}

type Page struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PageNumber int       `json:"page_number" db:"page_number"`
	Body       string    `json:"body" db:"body"`
	BookID     uuid.UUID `json:"book_id" db:"book_id"`
}

// NewSQLStore Returns a new SQL store with a postgres database connection.
func NewSQLStore(config config.DatabaseConfig) (*SQLStore, func()) {
	db, err := sqlx.Open("postgres", config.ConnStr())
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}

	// This is temporary until go-migrate is implemented
	_, errCreate := db.Exec(`
	CREATE SCHEMA IF NOT EXISTS library;
	CREATE TABLE IF NOT EXISTS books (
		id UUID PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		slug VARCHAR(255) NOT NULL,
		author VARCHAR(100),
		publication_year INTEGER,
		page_count INTEGER,
		file VARCHAR(255)
	);
	CREATE TABLE IF NOT EXISTS pages (
		id UUID PRIMARY KEY,
		page_number INT,
		body TEXT,
		book_id UUID REFERENCES books (id)
	);
	`)
	if errCreate != nil {
		log.Fatalf("failed to create tables %v", errCreate)
	}

	removeDatabase := func() {
		db.Close()
	}

	return &SQLStore{db}, removeDatabase
}

// Books fetches a list of books
func (s *SQLStore) Books(limit int) ([]Book, error) {
	books := make([]Book, 0)
	var lim int
	if limit == -1 {
		lim = 1000
	} else {
		lim = limit
	}
	stmt := `SELECT * FROM books LIMIT $1;`
	rows, err := s.DB.Queryx(stmt, lim)

	if err != nil {
		return nil, errors.Wrap(err, "database query failed")
	}

	for rows.Next() {
		var book Book
		if err = rows.StructScan(&book); err != nil {
			return nil, errors.Wrap(err, "error scanning database rows")
		}
		books = append(books, book)
	}

	return books, nil
}

// BookByID fetches a book by ID
func (s *SQLStore) BookByID(ID uuid.UUID) (Book, error) {
	var book Book
	stmt := `
	SELECT * FROM books
	WHERE id = $1
	LIMIT 1;
	`

	if err := s.DB.Get(&book, stmt, ID); err != nil {
		return book, errors.Wrap(err, "error querying database")
	}
	return book, nil
}

// BookBySlug fetches a book by slug
func (s *SQLStore) BookBySlug(slug string) (Book, error) {
	var book Book
	stmt := `
	SELECT * FROM books
	WHERE slug = $1
	LIMIT 1;
	`
	err := s.DB.Get(&book, stmt, slug)
	if err != nil {
		return book, errors.Wrap(err, "error querying database")
	}
	return book, nil
}

func (s *SQLStore) BooksByAuthor(author string) ([]Book, error) {
	books := make([]Book, 0)
	stmt := `SELECT * FROM books WHERE author = $1 LIMIT 1000;`
	rows, err := s.DB.Queryx(stmt, author)

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("database query failed\n\t%s", stmt))
	}

	for rows.Next() {
		var book Book
		if err = rows.StructScan(&book); err != nil {
			return nil, errors.Wrap(err, "error scanning database rows")
		}
		books = append(books, book)
	}

	return books, nil
}

// func (s *SQLStore) Page(bookId uuid.UUID, number int) Page {
// 	var page Page
// 	stmt := `
// 	SELECT * FROM pages
// 	WHERE book_id = $1
// 	AND
// 	page_number = $2
// 	LIMIT 1;`
// 	err := s.DB.Get(&page, stmt, bookId, number)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return page

// }
