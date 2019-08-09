package books

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"log"
)

// SQLStore houses the PostgreSQL connection
type SQLStore struct {
	DB *sqlx.DB
}

// NewSQLStore Returns a new SQL store with a postgres database connection.
func NewSQLStore(host, port, user, dbname, pass string) (*SQLStore, func()) {
	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		user,
		pass,
		host,
		port,
		dbname,
	)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}

	_, errCreate := db.Exec(CreateTables)
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
func (s *SQLStore) BookByID(ID uuid.UUID) (*Book, error) {
	var book Book
	stmt := `
	SELECT * FROM books
	WHERE id = $1
	LIMIT 1;
	`

	if err := s.DB.Get(&book, stmt, ID); err != nil {
		return nil, errors.Wrap(err, "error querying database")
	}
	return &book, nil
}

// BookBySlug fetches a book by slug
func (s *SQLStore) BookBySlug(slug string) (*Book, error) {
	var book Book
	stmt := `
	SELECT * FROM books
	WHERE slug = $1
	LIMIT 1;
	`
	if err := s.DB.Get(&book, stmt, slug); err != nil {
		return nil, errors.Wrap(err, "error querying database")
	}
	return &book, nil
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
