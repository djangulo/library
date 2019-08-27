package books

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // unneded namespace
	"log"
)

type SQLiteInMemoryStore struct {
	SQLStore
	Available bool
}

// NewInMemoryStore Returns a new SQLite in-memory database connection.
func NewInMemoryStore(available bool) (*SQLiteInMemoryStore, func()) {
	db, err := sqlx.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	removeDatabase := func() {
		db.Close()
	}

	return &SQLiteInMemoryStore{db, Available: available}, removeDatabase
}

func (s *SQLiteInMemoryStore) IsAvailable() error {
	if !s.Available {
		return ErrSQLStoreNotAvailable
	}
	return nil
}
