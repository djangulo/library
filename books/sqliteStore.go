package books

import (
	"fmt"
	// "database/sql"
	// config "github.com/djangulo/library/config/books"
	// "github.com/golang-migrate/migrate/v4"
	// "github.com/golang-migrate/migrate/v4/database/sqlite3"
	// _ "github.com/golang-migrate/migrate/v4/source/file" // unneeded namespace
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // unneded namespace
	"log"
)

type SQLiteInMemoryStore struct {
	SQLStore
	Available bool
}

func migrate(db *sqlx.DB) {
	db.Exec(`CREATE TABLE IF NOT EXISTS authors (
		id BLOB PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT UNIQUE
	)
	`)
	db.Exec(`CREATE TABLE IF NOT EXISTS books (
		id BLOB PRIMARY KEY,
		title TEXT NOT NULL,
		slug TEXT UNIQUE,
		publication_year INTEGER NULL,
		page_count INTEGER,
		file TEXT,
		source TEXT CHECK( source IN (  'nltk-gutenberg','open-library','manual-insert') )  DEFAULT 'nltk-gutenberg',
		author_id TEXT REFERENCES authors (id) NULL
	);
	`)
	db.Exec(`CREATE TABLE IF NOT EXISTS pages (
		id BLOB PRIMARY KEY,
		page_number INTEGER,
		body TEXT,
		book_id BLOB REFERENCES books(id)
	);
	`)
}

// NewInMemoryStore Returns a new SQLite in-memory database connection.
func NewInMemoryStore(addressID string, available bool) (*SQLiteInMemoryStore, func()) {
	// cnf := config.Get()
	connStr := fmt.Sprintf("file:%s?mode=memory&cache=shared", addressID)
	// connStr := "file::memory:?cache=shared"
	// connStr := fmt.Sprintf("%s.sqlite3", addressID)
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
	migrate(s.DB)

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

func (s *SQLiteInMemoryStore) IsAvailable() error {
	if !s.Available {
		return ErrSQLStoreUnavailable
	}
	return nil
}
