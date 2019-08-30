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
	"time"
)

const (
	seekON int = iota
	seekOFF
)

var (
	// ErrSQLStoreUnavailable returned if SQL store is unavailable
	ErrSQLStoreUnavailable = errors.New("attempted to access unavailable SQL connection")
	// ErrNoResults If a query returns empty
	ErrNoResults = errors.New("no results from query")
	// ErrNilPointerPassed if a nil pointer is passed
	ErrNilPointerPassed = errors.New("nil pointer passed in")
)

// SQLStore houses the PostgreSQL connection
type SQLStore struct {
	DB *sqlx.DB
}

// type sqlModel struct {
// 	ID        uuid.UUID `json:"id" db:"id" redis:"id"`
// 	CreatedAt time.Time `json:"created_at" db:"created_at" redis:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at" db:"updated_at" redis:"updated_at"`
// 	DeletedAt time.Time `json:"deleted_at" db:"deleted_at" redis:"deleted_at"`
// }

// Book struct
type Book struct {
	ID              uuid.UUID  `json:"id" db:"id" redis:"id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at" redis:"updated_at"`
	DeletedAt       time.Time  `json:"deleted_at" db:"deleted_at" redis:"deleted_at"`
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
	ID        uuid.UUID `json:"id" db:"id" redis:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" redis:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" db:"deleted_at" redis:"deleted_at"`
	Name      string    `json:"name" db:"name" redis:"name"`
	Slug      string    `json:"slug" db:"slug" redis:"slug"`
}

// Page struct
type Page struct {
	ID         uuid.UUID  `json:"id" db:"id" redis:"id"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at" redis:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at" redis:"updated_at"`
	DeletedAt  time.Time  `json:"deleted_at" db:"deleted_at" redis:"deleted_at"`
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

// Books fetches a list of books
func (s *SQLStore) Books(
	books []*Book,
	limit *int,
	offset *int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {
	if *limit == -1 {
		*limit = 1000
	}
	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	var stmt string
	var seeking int
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		seek := "WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $3"
		// no offset if seeking
		seeking = seekON
		stmt = fmt.Sprintf(
			"SELECT %s FROM books %s %s;",
			strings.Join(fields, ","),
			seek,
			lim,
		)
	} else {
		seek := "ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $1"
		off := "OFFSET $2"
		seeking = seekOFF
		stmt = fmt.Sprintf(
			"SELECT %s FROM books %s %s %s;",
			strings.Join(fields, ","),
			seek,
			lim,
			off,
		)
	}
	log.Println(stmt)
	var err error
	var rows *sqlx.Rows
	if seeking == seekOFF {
		rows, err = s.DB.Queryx(stmt, limit, offset)
	} else {
		rows, err = s.DB.Queryx(stmt, lastCreated, lastID, limit)
	}

	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("Books: query failed - %s", stmt),
		)
	}

	for rows.Next() {
		var book Book
		if err = rows.StructScan(&book); err != nil {
			log.Printf("Books: error scanning row, %v\n", err)
			continue
		}
		log.Printf("%v\n", book)
		books = append(books, &book)
		for _, b := range books {
			log.Printf("%v\n", b)
		}
	}
	if err := rows.Close(); err != nil {
		log.Printf("Books: error closing rows, %v\n", err)
	}

	return nil
}

// BookByID fetches a book by ID
func (s *SQLStore) BookByID(
	book *Book,
	ID *uuid.UUID,
	fields []string,
) error {

	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	stmt := fmt.Sprintf(
		"SELECT %s FROM books WHERE id = $1 LIMIT 1;",
		strings.Join(fields, ","),
	)
	if err := s.DB.Get(&book, stmt, ID); err != nil {
		return errors.Wrap(err, "BookByID: query failed")
	}
	return nil
}

// BookBySlug fetches a book by slug
func (s *SQLStore) BookBySlug(
	book *Book,
	slug *string,
	fields []string,
) error {
	*slug = Slugify(*slug, "-")

	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	stmt := fmt.Sprintf(
		"SELECT %s FROM books WHERE slug = $1 LIMIT 1;",
		strings.Join(fields, ","),
	)
	if err := s.DB.Get(&book, stmt, slug); err != nil {
		return errors.Wrap(err, "BookBySlug: query failed")
	}
	return nil
}

// BooksByAuthor returns books by a given author
func (s *SQLStore) BooksByAuthor(
	books []*Book,
	author *string,
	limit *int,
	offset *int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {
	*author = strings.ToLower(*author)

	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	innerFields := make([]string, len(fields), len(fields))
	for i := range fields {
		innerFields[i] = fmt.Sprintf("b.%s", fields[i])
	}

	if *limit == -1 {
		*limit = 1000
	}

	var stmt string
	var seeking int
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		seek := "AND (created_at, id) < ($3, $4) ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $5"
		// no offset if seeking
		seeking = seekON
		stmt = fmt.Sprintf(
			`
			SELECT
				%s
			FROM (
				SELECT
					a.slug AS author_slug, a.name, %s	
				FROM books AS b
				JOIN authors AS a
				ON b.author_id = a.id
			) AS books_authors
			WHERE
				lower(name) = $1 OR author_slug = $2
				%s %s;
			`,
			strings.Join(fields, ","),
			strings.Join(innerFields, ","),
			seek,
			lim,
		)
	} else {
		seek := "ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $3"
		off := "OFFSET $4"
		seeking = seekOFF
		stmt = fmt.Sprintf(
			`
			SELECT
				%s
			FROM (
				SELECT
					a.slug AS author_slug, a.name, %s	
				FROM books AS b
				JOIN authors AS a
				ON b.author_id = a.id
			) AS books_authors
			WHERE
				lower(name) = $1 OR author_slug = $2
				%s %s %s;
			`,
			strings.Join(fields, ","),
			strings.Join(innerFields, ","),
			seek,
			lim,
			off,
		)
	}

	var err error
	var rows *sqlx.Rows
	if seeking == seekOFF {
		rows, err = s.DB.Queryx(stmt, author, author, limit, offset)
	} else {
		rows, err = s.DB.Queryx(stmt, author, author, lastCreated, lastID, limit)
	}

	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("BooksByAuthor: query failed - %s", stmt),
		)
	}

	for rows.Next() {
		var book Book
		if err = rows.StructScan(&book); err != nil {
			log.Printf("BooksByAuthor: error scanning row, %v\n", err)
			continue
		}
		books = append(books, &book)
	}
	if err := rows.Close(); err != nil {
		log.Printf("BooksByAuthor: error closing rows, %v\n", err)
	}

	return nil
}

// Pages fetches a list of pages
func (s *SQLStore) Pages(
	pages []*Page,
	limit *int,
	offset *int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {
	if *limit == -1 {
		*limit = 1000
	}
	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	var stmt string
	var seeking int
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		seek := "WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $3"
		// no offset if seeking
		seeking = seekON
		stmt = fmt.Sprintf(
			"SELECT %s FROM pages %s %s;",
			strings.Join(fields, ","),
			seek,
			lim,
		)
	} else {
		seek := "ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $1"
		off := "OFFSET $2"
		seeking = seekOFF
		stmt = fmt.Sprintf(
			"SELECT %s FROM pages %s %s %s;",
			strings.Join(fields, ","),
			seek,
			lim,
			off,
		)
	}
	var err error
	var rows *sqlx.Rows
	if seeking == seekOFF {
		rows, err = s.DB.Queryx(stmt, limit, offset)
	} else {
		rows, err = s.DB.Queryx(stmt, lastCreated, lastID, limit)
	}

	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("Pages: query failed - %s", stmt),
		)
	}

	for rows.Next() {
		var page Page
		if err = rows.StructScan(&page); err != nil {
			log.Printf("Pages: error scanning row, %v\n", err)
			continue
		}
		pages = append(pages, &page)
	}
	if err := rows.Close(); err != nil {
		log.Printf("Pages: error closing rows, %v\n", err)
	}

	return nil
}

// PageByID fetches a page by ID
func (s *SQLStore) PageByID(
	page *Page,
	ID *uuid.UUID,
	fields []string,
) error {

	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	stmt := fmt.Sprintf(
		"SELECT %s FROM pages WHERE id = $1 LIMIT 1;",
		strings.Join(fields, ","),
	)
	if err := s.DB.Get(&page, stmt, ID); err != nil {
		return errors.Wrap(err, "PageByID: query failed")
	}
	return nil
}

// PageByBookAndNumber returns a page by book id and number
func (s *SQLStore) PageByBookAndNumber(
	page *Page,
	bookID *uuid.UUID,
	number *int,
	fields []string,
) error {
	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	innerFields := make([]string, len(fields), len(fields))
	for i := range fields {
		innerFields[i] = fmt.Sprintf("p.%s", fields[i])
	}

	stmt := fmt.Sprintf(`
	SELECT
		%s
	FROM (
		SELECT
			b.id AS books_book_id,
			%s
		FROM pages AS p
		JOIN books AS b
		ON b.id = p.book_id
	) AS pages_books
	WHERE book_id = $1 AND page_number = $2 LIMIT 1;
	`, strings.Join(fields, ","), strings.Join(innerFields, ","))
	if err := s.DB.Get(&page, stmt, bookID, number); err != nil {
		return errors.Wrap(err, "PageByBookAndNumber: query failed")
	}
	return nil
}

// Authors fetches a list of authors
func (s *SQLStore) Authors(
	authors []*Author,
	limit *int,
	offset *int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {

	if *limit == -1 {
		*limit = 1000
	}
	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	var stmt string
	var seeking int
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		seek := "WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $3"
		// no offset if seeking
		seeking = seekON
		stmt = fmt.Sprintf(
			"SELECT %s FROM authors %s %s;",
			strings.Join(fields, ","),
			seek,
			lim,
		)
	} else {
		seek := "ORDER BY created_at DESC, id DESC"
		lim := "LIMIT $1"
		off := "OFFSET $2"
		seeking = seekOFF
		stmt = fmt.Sprintf(
			"SELECT %s FROM authors %s %s %s;",
			strings.Join(fields, ","),
			seek,
			lim,
			off,
		)
	}
	var err error
	var rows *sqlx.Rows
	if seeking == seekOFF {
		rows, err = s.DB.Queryx(stmt, limit, offset)
	} else {
		rows, err = s.DB.Queryx(stmt, lastCreated, lastID, limit)
	}

	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf("Authors: query failed - %s", stmt),
		)
	}

	for rows.Next() {
		var author Author
		if err = rows.StructScan(&author); err != nil {
			log.Printf("Authors: error scanning row, %v\n", err)
			continue
		}
		authors = append(authors, &author)
	}
	if err := rows.Close(); err != nil {
		log.Printf("Authors: error closing rows, %v\n", err)
	}

	return nil
}

// AuthorByID fetches an author by ID
func (s *SQLStore) AuthorByID(
	author *Author,
	ID *uuid.UUID,
	fields []string,
) error {
	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	stmt := fmt.Sprintf(
		"SELECT %s FROM authors WHERE id = $1 LIMIT 1;",
		strings.Join(fields, ","),
	)
	if err := s.DB.Get(&author, stmt, ID); err != nil {
		return errors.Wrap(err, "AuthorByID: query failed")
	}
	return nil
}

// AuthorBySlug fetches an author by slug
func (s *SQLStore) AuthorBySlug(
	author *Author,
	slug *string,
	fields []string,
) error {
	*slug = Slugify(*slug, "-")

	if len(fields) == 0 || fields == nil {
		fields = []string{"*"}
	}

	stmt := fmt.Sprintf(
		"SELECT %s FROM authors WHERE slug = $1 LIMIT 1;",
		strings.Join(fields, ","),
	)
	if err := s.DB.Get(&author, stmt, slug); err != nil {
		return errors.Wrap(err, "AuthorBySlug: query failed")
	}
	return nil
}

// InsertBook noqa
func (s *SQLStore) InsertBook(book *Book) error {
	if book == nil {
		return ErrNilPointerPassed
	}

	stmt := `
	INSERT INTO books (
		id,
		title,
		slug,
		publication_year,
		page_count,
		file,
		author_id,
		source,
		created_at,
		updated_at,
		deleted_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
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
		book.CreatedAt,
		time.Now(), // updated_at
		book.DeletedAt,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertBook: failed on book %v", book))
	}
	return nil
}

// InsertPage noqa
func (s *SQLStore) InsertPage(page *Page) error {

	stmt := `
	INSERT INTO pages (
		id,
		book_id,
		page_number,
		body,
		created_at,
		updated_at,
		deleted_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err := s.DB.Queryx(
		stmt,
		page.ID,
		page.BookID,
		page.PageNumber,
		page.Body,
		page.CreatedAt,
		time.Now(), // updated_at
		page.DeletedAt,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertPage: failed on page %v", page))
	}
	return nil
}

// InsertAuthor noqa
func (s *SQLStore) InsertAuthor(author *Author) error {
	stmt := `
	INSERT INTO authors (
		id,
		slug,
		name,
		created_at,
		updated_at,
		deleted_at
	)
	VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := s.DB.Queryx(
		stmt,
		author.ID,
		author.Slug,
		author.Name,
		author.CreatedAt,
		time.Now(), // updated_at
		author.DeletedAt,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertAuthor: failed on author %v", author))
	}
	return nil
}

// BulkInsertBooks noqa
func (s *SQLStore) BulkInsertBooks(books []*Book) error {

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	now := time.Now()
	for i, book := range books {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, %d, %d, %d)",
				i*11+1, i*11+2, i*11+3, i*11+4, i*11+5, i*11+6, i*11+7, i*11+8,
				i*11+9, i*11+10, i*11+11,
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
		valueArgs = append(valueArgs, book.CreatedAt)
		valueArgs = append(valueArgs, now) //updated_at
		valueArgs = append(valueArgs, book.DeletedAt)
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
		source,
		created_at,
		updated_at,
		deleted_at
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

// BulkInsertPages noqa
func (s *SQLStore) BulkInsertPages(pages []*Page) error {

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	now := time.Now()
	for i, page := range pages {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7,
			),
		)
		valueArgs = append(valueArgs, page.ID)
		valueArgs = append(valueArgs, page.PageNumber)
		valueArgs = append(valueArgs, page.Body)
		valueArgs = append(valueArgs, page.BookID)
		valueArgs = append(valueArgs, page.CreatedAt)
		valueArgs = append(valueArgs, now) //updated_at
		valueArgs = append(valueArgs, page.DeletedAt)
		// fmt.Printf("id: %v\n", page.ID)
	}

	stmt := fmt.Sprintf(`
	INSERT INTO pages (
		id,
		page_number,
		body,
		book_id,
		created_at,
		updated_at,
		deleted_at
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

// BulkInsertAuthors noqa
func (s *SQLStore) BulkInsertAuthors(authors []*Author) error {

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	now := time.Now()
	for i, author := range authors {
		valueStrings = append(
			valueStrings,
			fmt.Sprintf(
				"($%d, $%d, $%d, $%d, $%d, $%d)",
				i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6,
			),
		)
		valueArgs = append(valueArgs, author.ID)
		valueArgs = append(valueArgs, author.Slug)
		valueArgs = append(valueArgs, author.Name)
		valueArgs = append(valueArgs, author.CreatedAt)
		valueArgs = append(valueArgs, now) //updated_at
		valueArgs = append(valueArgs, author.DeletedAt)
	}

	stmt := fmt.Sprintf(`
	INSERT INTO authors (
		id,
		slug,
		name,
		created_at,
		updated_at,
		deleted_at
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
