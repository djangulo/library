package books

import (
	"fmt"
	config "github.com/djangulo/library/config/books"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // unneded namespace
	"github.com/pkg/errors"
	"log"
	"strings"
)

var (
	// ErrSQLStoreUnavailable returned if SQL store is unavailable
	ErrSQLStoreUnavailable = errors.New("Attempted to access unavailable SQL connection")
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
}

// Author struct
type Author struct {
	ID   uuid.UUID `json:"id" db:"id" redis:"id"`
	Name string    `json:"name" db:"name" redis:"name"`
	Slug string    `json:"slug" db:"slug" redis:"slug"`
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
	if err != nil {
		log.Fatalf("failed to prepare statements (prepareStatements) %v", err)
	}
	removeDatabase := func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("error closing connection to database: %v", err)
		}
	}

	return &SQLStore{db}, removeDatabase
}

// IsAvailable checks whether it's possible to connect to the DB or not
func (s *SQLStore) IsAvailable() error {
	return s.DB.Ping()
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

	stmt := `
	SELECT
		id,
		title,
		slug,
		publication_year,
		page_count,
		file,
		author_id,
		source
	FROM books ORDER BY title LIMIT $1 OFFSET $2;`
	rows, err := s.DB.Queryx(stmt, lim, off)

	if err != nil {
		return nil, errors.Wrap(err, "Books: query failed")
	}

	for rows.Next() {
		var book Book
		if err = rows.StructScan(&book); err != nil {
			log.Printf("Books: error scanning row, %v\n", err)
			continue
		}
		books = append(books, book)
	}

	return books, nil
}

// BookByID fetches a book by ID
func (s *SQLStore) BookByID(ID uuid.UUID) (Book, error) {
	var book Book

	stmt := `
	SELECT
		id,
		title,
		slug,
		publication_year,
		page_count,
		file,
		author_id,
		source
	FROM books WHERE id = $1 LIMIT 1;`
	if err := s.DB.Get(&book, stmt, ID); err != nil {
		return book, errors.Wrap(err, "BookByID: query failed")
	}
	return book, nil
}

// BookBySlug fetches a book by slug
func (s *SQLStore) BookBySlug(slug string) (Book, error) {
	var book Book
	stmt := `
	SELECT
		id,
		title,
		slug,
		publication_year,
		page_count,
		file,
		author_id,
		source
	FROM books WHERE slug = $1 LIMIT 1;`
	if err := s.DB.Get(&book, stmt, slug); err != nil {
		return book, errors.Wrap(err, "BookBySlug: query failed")
	}
	return book, nil
}

// BooksByAuthor returns books by a given author
func (s *SQLStore) BooksByAuthor(author string) ([]Book, error) {
	books := make([]Book, 0)
	author = strings.ToLower(author)

	stmt := `
	SELECT
		id, title, book_slug AS slug, publication_year, page_count, file,
		source
	FROM (
		SELECT
			b.id, b.title, b.slug AS book_slug, b.publication_year,
			b.page_count, b.file, b.source, b.author_id,
			a.slug AS author_slug, a.name
		FROM books AS b
		JOIN authors AS a
		ON b.author_id = a.id
	) AS books_authors
	WHERE lower(name) = $1 OR author_slug = $2;
	`
	rows, err := s.DB.Queryx(stmt, author, author)
	if err != nil {
		return nil, errors.Wrap(err, "BooksByAuthor: query failed")
	}

	for rows.Next() {
		var book Book
		if err = rows.StructScan(&book); err != nil {
			log.Printf("BooksByAuthor: error scanning row, %v\n", err)
			continue
		}
		books = append(books, book)
	}

	return books, nil
}

// Pages fetches a list of pages
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

	stmt := `SELECT id, page_number, book_id, body FROM pages LIMIT $1 OFFSET $2;`
	rows, err := s.DB.Queryx(stmt, lim, off)
	if err != nil {
		return nil, errors.Wrap(err, "Pages: query failed")
	}

	for rows.Next() {
		var page Page
		if err = rows.StructScan(&page); err != nil {
			log.Printf("Pages: error scanning row, %v\n", err)
			continue
		}
		pages = append(pages, page)
	}

	return pages, nil
}

// PageByID fetches a page by ID
func (s *SQLStore) PageByID(ID uuid.UUID) (Page, error) {
	var page Page
	stmt := `SELECT id, page_number, book_id, body FROM pages WHERE id = $1 LIMIT 1;`
	if err := s.DB.Get(&page, stmt, ID); err != nil {
		return page, errors.Wrap(err, "PageByID: query failed")
	}
	return page, nil
}

// PageByBookAndNumber returns a page by book id and number
func (s *SQLStore) PageByBookAndNumber(bookID uuid.UUID, number int) (Page, error) {
	var page Page

	stmt := `
	SELECT
		id, page_number, book_id, body
	FROM (
		SELECT
			b.id AS books_book_id,
			p.id, p.page_number, p.book_id, p.body
		FROM pages AS p
		JOIN books AS b
		ON b.id = p.book_id
	) AS pages_books
	WHERE book_id = $1 AND page_number = $2;
	`
	if err := s.DB.Get(&page, stmt, bookID, number); err != nil {
		return page, errors.Wrap(err, "PageByBookAndNumber: query failed")
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

	stmt := `SELECT id, name, slug FROM authors ORDER BY name LIMIT $1 OFFSET $2;`
	rows, err := s.DB.Queryx(stmt, lim, off)
	if err != nil {
		return nil, errors.Wrap(err, "Authors: query failed")
	}

	for rows.Next() {
		var author Author
		if err = rows.StructScan(&author); err != nil {
			log.Printf("Authors: error scanning row, %v\n", err)
			continue
		}
		authors = append(authors, author)
	}

	return authors, nil
}

// AuthorByID fetches an auhtor by ID
func (s *SQLStore) AuthorByID(ID uuid.UUID) (Author, error) {
	var author Author

	stmt := `SELECT id, name, slug FROM authors WHERE id = $1 LIMIT 1;`
	if err := s.DB.Get(&author, stmt, ID); err != nil {
		return author, errors.Wrap(err, "AuthorByID: query failed")
	}
	return author, nil
}

// AuthorBySlug fetches an author by slug
func (s *SQLStore) AuthorBySlug(slug string) (Author, error) {
	slug = Slugify(slug, "-")
	var author Author

	stmt := `SELECT id, name, slug FROM authors WHERE slug = $1 LIMIT 1;`
	if err := s.DB.Get(&author, stmt, slug); err != nil {
		return author, errors.Wrap(err, "AuthorByID: query failed")
	}
	return author, nil
}

func (s *SQLStore) InsertBook(book Book) error {

	stmt := `
	INSERT INTO books (
		id,
		title,
		slug,
		publication_year,
		page_count,
		file,
		author_id,
		source
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err := s.DB.Queryx(
		stmt,
		book.ID,
		book.Title,
		book.Slug,
		book.PublicationYear,
		book.PageCount,
		book.File,
		book.AuthorID,
		book.Source,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertBook: failed on book %v", book))
	}
	return nil
}

func (s *SQLStore) InsertPage(page Page) error {

	stmt := `
	INSERT INTO pages (
		id,
		book_id,
		page_number,
		body
	)
	VALUES ($1, $2, $3, $4);
	`
	_, err := s.DB.Queryx(
		stmt,
		page.ID,
		page.BookID,
		page.PageNumber,
		page.Body,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertPage: failed on page %v", page))
	}
	return nil
}

func (s *SQLStore) InsertAuthor(author Author) error {
	stmt := `
	INSERT INTO authors (
		id,
		slug,
		name
	)
	VALUES ($1, $2, $3, $4);
	`
	_, err := s.DB.Queryx(
		stmt,
		author.ID,
		author.Slug,
		author.Name,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertAuthor: failed on author %v", author))
	}
	return nil
}

func (s *SQLStore) BulkInsertBooks(books []Book) error {

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	for i, book := range books {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8,
			),
		)
		valueArgs = append(valueArgs, book.ID)
		valueArgs = append(valueArgs, book.Title)
		valueArgs = append(valueArgs, book.Slug)
		valueArgs = append(valueArgs, book.PublicationYear)
		valueArgs = append(valueArgs, book.PageCount)
		valueArgs = append(valueArgs, book.File)
		valueArgs = append(valueArgs, book.AuthorID)
		valueArgs = append(valueArgs, book.Source)
	}

	stmt := fmt.Sprintf(`
	INSERT INTO books (
		id,
		title,
		slug,
		publication_year,
		page_count,
		file,
		author_id,
		source
	) VALUES %s;`, strings.Join(valueStrings, ","))

	tx, err := s.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "BulkInsertBooks: begin transaction failed")
	}

	_, err = tx.Exec(`SET CLIENT_ENCODING TO 'LATIN2';`)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				rollbackErr,
				"BulkInsertBooks: set encoding, unable to rollback",
			)
		}
		return errors.Wrap(err, "BulkInsertBooks: set encoding failed")
	}
	_, err = tx.Exec(stmt, valueArgs...)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				err,
				"BulkInsertBooks: insert, unable to rollback",
			)
		}
		return errors.Wrap(
			err,
			"BulkInsertBooks: insert failed",
		)
	}
	_, err = tx.Exec(`RESET CLIENT_ENCODING;`)
	if err != nil {

		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				err,
				"BulkInsertBooks: reset encoding, unable to rollback",
			)
		}
		return errors.Wrap(
			err,
			"BulkInsertBooks: reset encoding failed",
		)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return errors.Wrap(err, "BulkInsertBooks: commit failed")
	}

	return nil
}

func (s *SQLStore) BulkInsertPages(pages []Page) error {

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	for i, page := range pages {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d)",
				i*8+1, i*8+2, i*8+3, i*8+4,
			),
		)
		valueArgs = append(valueArgs, page.ID)
		valueArgs = append(valueArgs, page.BookID)
		valueArgs = append(valueArgs, page.PageNumber)
		valueArgs = append(valueArgs, page.Body)
	}

	stmt := fmt.Sprintf(`
	INSERT INTO pages (
		id,
		book_id,
		page_number,
		body
	) VALUES %s;`, strings.Join(valueStrings, ","))

	tx, err := s.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "BulkInsertPages: begin transaction failed")
	}

	_, err = tx.Exec(`SET CLIENT_ENCODING TO 'LATIN2';`)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				rollbackErr,
				"BulkInsertPages: set encoding, unable to rollback",
			)
		}
		return errors.Wrap(err, "BulkInsertPages: set encoding failed")
	}
	_, err = tx.Exec(stmt, valueArgs...)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				err,
				"BulkInsertPages: insert, unable to rollback",
			)
		}
		return errors.Wrap(
			err,
			"BulkInsertPages: insert failed",
		)
	}
	_, err = tx.Exec(`RESET CLIENT_ENCODING;`)
	if err != nil {

		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				err,
				"BulkInsertPages: reset encoding, unable to rollback",
			)
		}
		return errors.Wrap(
			err,
			"BulkInsertPages: reset encoding failed",
		)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return errors.Wrap(err, "BulkInsertPages: commit failed")
	}

	return nil
}

func (s *SQLStore) BulkInsertAuthors(authors []Author) error {

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	for i, author := range authors {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d)",
				i*8+1, i*8+2, i*8+3,
			),
		)
		valueArgs = append(valueArgs, author.ID)
		valueArgs = append(valueArgs, author.Slug)
		valueArgs = append(valueArgs, author.Name)
	}

	stmt := fmt.Sprintf(`
	INSERT INTO authors (
		id,
		slug,
		name
	) VALUES %s;`, strings.Join(valueStrings, ","))

	tx, err := s.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "BulkInsertAuthors: begin transaction failed")
	}

	_, err = tx.Exec(`SET CLIENT_ENCODING TO 'LATIN2';`)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				rollbackErr,
				"BulkInsertAuthors: set encoding, unable to rollback",
			)
		}
		return errors.Wrap(err, "BulkInsertAuthors: set encoding failed")
	}
	_, err = tx.Exec(stmt, valueArgs...)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				err,
				"BulkInsertAuthors: insert, unable to rollback",
			)
		}
		return errors.Wrap(
			err,
			"BulkInsertAuthors: insert failed",
		)
	}
	_, err = tx.Exec(`RESET CLIENT_ENCODING;`)
	if err != nil {

		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				err,
				"BulkInsertAuthors: reset encoding, unable to rollback",
			)
		}
		return errors.Wrap(
			err,
			"BulkInsertAuthors: reset encoding failed",
		)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return errors.Wrap(err, "BulkInsertAuthors: commit failed")
	}

	return nil
}
