package books

import (
	"fmt"
	config "github.com/djangulo/library/config/books"
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

// Book struct
type Book struct {
	ID              uuid.UUID  `json:"id" db:"id" redis:"id"`
	Title           string     `json:"title" db:"title" redis:"title"`
	Slug            string     `json:"slug" db:"slug" redis:"slug"`
	PublicationYear NullInt64  `json:"publication_year" db:"publication_year" redis:"publication_year"`
	PageCount       int        `json:"page_count" db:"page_count" redis:"page_count"`
	File            NullString `json:"file" db:"file" redis:"file"`
	Source          NullString `json:"source" db:"source" redis:"source"`
	AuthorID        NullUUID   `json:"author_id" db:"author_id" redis:"author_id"`
	Pages           []Page     `json:"pages"`
}

// Author struct
type Author struct {
	ID    uuid.UUID `json:"id" db:"id" redis:"id"`
	Name  string    `json:"name" db:"name" redis:"name"`
	Slug  string    `json:"slug" db:"slug" redis:"slug"`
	Books []Book    `json:"books"`
}

// Page struct
type Page struct {
	ID         uuid.UUID  `json:"id" db:"id" redis:"id"`
	PageNumber int        `json:"page_number" db:"page_number" redis:"page_number"`
	Body       string     `json:"body" db:"body" redis:"body"`
	BookID     *uuid.UUID `json:"book_id" db:"book_id" redis:"book_id"`
}

// NewSQLStore Returns a new SQL store with a postgres database connection.
func NewSQLStore(config config.DatabaseConfig) (*SQLStore, func()) {
	db, err := sqlx.Open("postgres", config.ConnStr())
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	removeDatabase := func() {
		db.Close()
	}

	return &SQLStore{db}, removeDatabase
}

// Books fetches a list of books
func (s *SQLStore) Books(limit, offset int) ([]Book, error) {
	books := make([]Book, 0)
	var lim int
	if limit == -1 {
		lim = 1000
	} else {
		lim = limit
	}
	var off int
	if offset == -1 || offset == 0 {
		off = 0
	} else {
		off = offset
	}
	stmt := `SELECT * FROM books ORDER BY title LIMIT $1 OFFSET $2;`
	rows, err := s.DB.Queryx(stmt, lim, off)

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

// BooksByAuthor returns books by a given author
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

// Pages fetches a list of books
func (s *SQLStore) Pages(limit, offset int) ([]Page, error) {
	pages := make([]Page, 0)
	var lim int
	if limit == -1 {
		lim = 1000
	} else {
		lim = limit
	}
	var off int
	if offset == -1 || offset == 0 {
		off = 0
	} else {
		off = offset
	}
	stmt := `SELECT * FROM pages ORDER BY page_number LIMIT $1 OFFSET $2;`
	rows, err := s.DB.Queryx(stmt, lim, off)

	if err != nil {
		return nil, errors.Wrap(err, "database query failed")
	}

	for rows.Next() {
		var page Page
		if err = rows.StructScan(&page); err != nil {
			return nil, errors.Wrap(err, "error scanning database rows")
		}
		pages = append(pages, page)
	}

	return pages, nil
}

// PageByID fetches a page by ID
func (s *SQLStore) PageByID(ID uuid.UUID) (Page, error) {
	var page Page
	stmt := `
	SELECT * FROM pages
	WHERE id = $1
	LIMIT 1;
	`
	if err := s.DB.Get(&page, stmt, ID); err != nil {
		return page, errors.Wrap(err, "error querying database")
	}
	return page, nil
}

// PageByBookAndNumber returns a page by book id and number
func (s *SQLStore) PageByBookAndNumber(bookID uuid.UUID, number int) (Page, error) {
	var page Page
	stmt := `
	SELECT * FROM pages
	WHERE book = $1
	AND
	page_number = $2
	LIMIT 1;`
	if err := s.DB.Get(&page, stmt, bookID, number); err != nil {
		return page, errors.Wrap(err, "error querying database")
	}
	return page, nil
}

// Authors fetches a list of authors
func (s *SQLStore) Authors(limit, offset int) ([]Author, error) {
	authors := make([]Author, 0)
	var lim int
	if limit == -1 {
		lim = 1000
	} else {
		lim = limit
	}
	var off int
	if offset == -1 || offset == 0 {
		off = 0
	} else {
		off = offset
	}
	stmt := `SELECT * FROM authors ORDER BY name LIMIT $1 OFFSET $2;`
	rows, err := s.DB.Queryx(stmt, lim, off)

	if err != nil {
		return nil, errors.Wrap(err, "database query failed")
	}

	for rows.Next() {
		var author Author
		if err = rows.StructScan(&author); err != nil {
			return nil, errors.Wrap(err, "error scanning database rows")
		}
		authors = append(authors, author)
	}

	return authors, nil
}

// AuthorByID fetches an auhtor by ID
func (s *SQLStore) AuthorByID(ID uuid.UUID) (Author, error) {
	var author Author
	stmt := `
	SELECT * FROM authors
	WHERE id = $1
	LIMIT 1;
	`

	if err := s.DB.Get(&author, stmt, ID); err != nil {
		return author, errors.Wrap(err, "error querying database")
	}
	return author, nil
}

// AuthorBySlug fetches an author by slug
func (s *SQLStore) AuthorBySlug(slug string) (Author, error) {
	var author Author
	stmt := `
	SELECT * FROM authors
	WHERE slug = $1
	LIMIT 1;
	`
	err := s.DB.Get(&author, stmt, slug)
	if err != nil {
		return author, errors.Wrap(err, "error querying database")
	}
	return author, nil
}
