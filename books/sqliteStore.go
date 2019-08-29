package books

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // unneded namespace
	"github.com/pkg/errors"
	"log"
	"strings"
	"time"
)

// SQLiteInMemoryStore embed of SQLStore with some SQLite3 syntax edits
type SQLiteInMemoryStore struct {
	SQLStore
	Available bool
}

func migrate(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS authors (
		id BLOB PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TEXT NULL
	)
	`)
	if err != nil {
		return errors.Wrap(err, "SQLiteInMemoryStore: failed to create 'authors' table")
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS books (
		id BLOB PRIMARY KEY,
		title TEXT NOT NULL,
		slug TEXT UNIQUE,
		publication_year INTEGER NULL,
		page_count INTEGER,
		file TEXT,
		source TEXT CHECK( source IN ( 'nltk-gutenberg','open-library','manual-insert') )  DEFAULT 'nltk-gutenberg',
		author_id TEXT REFERENCES authors (id) NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TEXT NULL
	);
	`)
	if err != nil {
		return errors.Wrap(err, "SQLiteInMemoryStore: failed to create 'books' table")
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pages (
		id BLOB PRIMARY KEY,
		page_number INTEGER,
		body TEXT,
		book_id BLOB REFERENCES books(id),
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at TEXT NULL
	);
	`)
	if err != nil {
		return errors.Wrap(err, "SQLiteInMemoryStore: failed to create 'pages' table")
	}
	return nil
}

// NewInMemoryStore Returns a new SQLite in-memory database connection.
func NewInMemoryStore(addressID string, available bool) (*SQLiteInMemoryStore, func()) {
	// cnf := config.Get()
	connStr := fmt.Sprintf("file:%s?mode=memory&cache=shared", addressID)
	// connStr := "file::memory:?cache=shared"
	// connStr := fmt.Sprintf("%s.sqlite3", addressID)
	// connStr := "main.sqlite3"
	db, err := sqlx.Open("sqlite3", connStr)
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	removeDatabase := func() {
		db.Close()
	}

	s := SQLiteInMemoryStore{}
	s.DB = db
	s.Available = available
	err = migrate(s.DB)
	if err != nil {
		log.Fatalf("failed to migrate SQLiteInMemoryStore: %v", err)
	}

	// migrateConn, err := sql.Open("sqlite3", connStr)
	// if err != nil {
	// 	log.Fatalf("failed to connect to sqlite database %v", err)
	// }
	// defer migrateConn.Close()
	// driver, err := sqlite3.WithInstance(migrateConn, &sqlite3.Config{})
	// m, err := migrate.NewWithDatabaseInstance(
	// 	"file://"+cnf.Project.Dirs.Migrations,
	// 	"sqlite3",
	// 	driver,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// m.Up()

	return &s, removeDatabase
}

// IsAvailable noqa
func (s *SQLiteInMemoryStore) IsAvailable() error {
	if !s.Available {
		return ErrSQLStoreUnavailable
	}
	return nil
}

// SeedSQLite noqa
func SeedSQLite(db *SQLiteInMemoryStore, authors []Author, books []Book, pages []Page) error {
	err := db.BulkInsertAuthors(authors)
	if err != nil {
		return errors.Wrap(err, "failed to bulk insert authors")
	}
	err = db.BulkInsertBooks(books)
	if err != nil {
		return errors.Wrap(err, "failed to bulk insert books")
	}
	nPages := len(pages)
	for i := 0; i < nPages; i += 200 {
		var err error
		if diff := nPages - i; diff < 200 {
			err = db.BulkInsertPages(pages[i:(i + diff)])
		} else {
			err = db.BulkInsertPages(pages[i:(i + 200)])
		}
		if err != nil {
			return errors.Wrap(err, "failed to bulk insert pages")
		}
	}
	return nil
}

// InsertBook noqa
func (s *SQLiteInMemoryStore) InsertBook(book Book) error {

	stmt := `
	INSERT OR REPLACE INTO books (
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
		book.CreatedAt.Format(time.RFC3339),
		time.Now().Format(time.RFC3339), // updated_at
		book.DeletedAt.Format(time.RFC3339),
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertBook: failed on book %v", book))
	}
	return nil
}

// InsertPage noqa
func (s *SQLiteInMemoryStore) InsertPage(page Page) error {

	stmt := `
	INSERT OR REPLACE INTO pages (
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
		page.CreatedAt.Format(time.RFC3339),
		time.Now().Format(time.RFC3339), // updated_at
		page.DeletedAt.Format(time.RFC3339),
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertPage: failed on page %v", page))
	}
	return nil
}

// InsertAuthor noqa
func (s *SQLiteInMemoryStore) InsertAuthor(author Author) error {
	stmt := `
	INSERT OR REPLACE INTO authors (
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
		author.CreatedAt.Format(time.RFC3339),
		time.Now().Format(time.RFC3339), // updated_at
		author.DeletedAt.Format(time.RFC3339),
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("InsertAuthor: failed on author %v", author))
	}
	return nil
}

// BulkInsertBooks noqa
func (s *SQLiteInMemoryStore) BulkInsertBooks(books []Book) error {

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
		valueArgs = append(valueArgs, book.CreatedAt.Format(time.RFC3339))
		valueArgs = append(valueArgs, now.Format(time.RFC3339)) //updated_at
		valueArgs = append(valueArgs, book.DeletedAt.Format(time.RFC3339))
	}

	stmt := fmt.Sprintf(`
	INSERT OR REPLACE INTO books (
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

	if commitErr := tx.Commit(); commitErr != nil {
		return errors.Wrap(err, "BulkInsertBooks: commit failed")
	}

	return nil
}

// BulkInsertPages noqa
func (s *SQLiteInMemoryStore) BulkInsertPages(pages []Page) error {

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
		valueArgs = append(valueArgs, page.CreatedAt.Format(time.RFC3339))
		valueArgs = append(valueArgs, now.Format(time.RFC3339)) //updated_at
		valueArgs = append(valueArgs, page.DeletedAt.Format(time.RFC3339))
	}

	stmt := fmt.Sprintf(`
	INSERT OR REPLACE INTO pages (
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

	if commitErr := tx.Commit(); commitErr != nil {
		return errors.Wrap(err, "BulkInsertPages: commit failed")
	}

	return nil
}

// BulkInsertAuthors noqa
func (s *SQLiteInMemoryStore) BulkInsertAuthors(authors []Author) error {

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
		valueArgs = append(valueArgs, author.CreatedAt.Format(time.RFC3339))
		valueArgs = append(valueArgs, now.Format(time.RFC3339)) //updated_at
		valueArgs = append(valueArgs, author.DeletedAt.Format(time.RFC3339))
	}

	stmt := fmt.Sprintf(`
	INSERT OR REPLACE INTO authors (
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

	if commitErr := tx.Commit(); commitErr != nil {
		return errors.Wrap(err, "BulkInsertAuthors: commit failed")
	}

	return nil
}
