package books

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // unneded namespace
	"log"
)

// NewInMemoryStore Returns a new SQLite in-memory database connection.
func NewInMemoryStore() (*SQLStore, func()) {
	db, err := sqlx.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	removeDatabase := func() {
		db.Close()
	}

	return &SQLStore{db}, removeDatabase
}
