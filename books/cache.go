package books

import (
	"github.com/gomodule/redigo/redis"
	config "github.com/djangulo/library/config/books"
	"log"
	"time"
	"github.com/pkg/errors"
)

type RedisCache struct {
	Available bool
	Pool      *redis.Pool
}

func (r *RedisCache) IsAvailable() bool {
	if !r.Available {
		log.Println("Attempted to access unavailable redis cache")
		return false
	}
	return true
}

func NewRedisCache(config config.CacheConfig) (*RedisCache, error) {
	connStr := config.ConnStr()
	conn, err := redis.Dial("tcp", connStr)
	defer conn.Close()
	if err != nil {
		log.Printf("Redis connection unavailable: %v", err)
		return &RedisCache{Available: false}
	}
	return &RedisCache{
		Available: true,
		Pool: &redis.Pool{
			MaxIdle: 10,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial(tcp, connStr)
			}
		}
	}
}


// // Books fetches a list of books
// func (r *RedisCache) Books(limit, offset int) ([]Book, error) {
// 	books := make([]Book, 0)
// 	if !r.Available {
// 		log.Println("Attempted to access unavailable redis cache")
// 		return books, errors.New("attempted to access unavailable redis cache")
// 	}
// 	conn := r.Pool.Get()
// 	var lim int
// 	if limit == -1 {
// 		lim = 1000
// 	} else {
// 		lim = limit
// 	}
// 	var off int
// 	if offset == -1 || offset == 0 {
// 		off = 0
// 	} else {
// 		off = offset
// 	}
// 	stmt := `SELECT * FROM books ORDER BY title LIMIT $1 OFFSET $2;`
// 	rows, err := s.DB.Queryx(stmt, lim, off)

// 	if err != nil {
// 		return nil, errors.Wrap(err, "database query failed")
// 	}

// 	for rows.Next() {
// 		var book Book
// 		if err = rows.StructScan(&book); err != nil {
// 			return nil, errors.Wrap(err, "error scanning database rows")
// 		}
// 		books = append(books, book)
// 	}

// 	return books, nil
// }

// // BookByID fetches a book by ID
// func (r *RedisCache) BookByID(ID uuid.UUID) (Book, error) {
// 	var book Book
// 	stmt := `
// 	SELECT * FROM books
// 	WHERE id = $1
// 	LIMIT 1;
// 	`

// 	if err := s.DB.Get(&book, stmt, ID); err != nil {
// 		return book, errors.Wrap(err, "error querying database")
// 	}
// 	return book, nil
// }

// // BookBySlug fetches a book by slug
// func (r *RedisCache) BookBySlug(slug string) (Book, error) {
// 	var book Book
// 	stmt := `
// 	SELECT * FROM books
// 	WHERE slug = $1
// 	LIMIT 1;
// 	`
// 	err := s.DB.Get(&book, stmt, slug)
// 	if err != nil {
// 		return book, errors.Wrap(err, "error querying database")
// 	}
// 	return book, nil
// }

// // BooksByAuthor returns books by a given author
// func (r *RedisCache) BooksByAuthor(author string) ([]Book, error) {
// 	books := make([]Book, 0)
// 	stmt := `SELECT * FROM books WHERE author = $1 LIMIT 1000;`
// 	rows, err := s.DB.Queryx(stmt, author)

// 	if err != nil {
// 		return nil, errors.Wrap(err, fmt.Sprintf("database query failed\n\t%s", stmt))
// 	}

// 	for rows.Next() {
// 		var book Book
// 		if err = rows.StructScan(&book); err != nil {
// 			return nil, errors.Wrap(err, "error scanning database rows")
// 		}
// 		books = append(books, book)
// 	}

// 	return books, nil
// }

// // Pages fetches a list of books
// func (r *RedisCache) Pages(limit, offset int) ([]Page, error) {
// 	pages := make([]Page, 0)
// 	var lim int
// 	if limit == -1 {
// 		lim = 1000
// 	} else {
// 		lim = limit
// 	}
// 	var off int
// 	if offset == -1 || offset == 0 {
// 		off = 0
// 	} else {
// 		off = offset
// 	}
// 	stmt := `SELECT * FROM pages ORDER BY page_number LIMIT $1 OFFSET $2;`
// 	rows, err := s.DB.Queryx(stmt, lim, off)

// 	if err != nil {
// 		return nil, errors.Wrap(err, "database query failed")
// 	}

// 	for rows.Next() {
// 		var page Page
// 		if err = rows.StructScan(&page); err != nil {
// 			return nil, errors.Wrap(err, "error scanning database rows")
// 		}
// 		pages = append(pages, page)
// 	}

// 	return pages, nil
// }

// // PageByID fetches a page by ID
// func (r *RedisCache) PageByID(ID uuid.UUID) (Page, error) {
// 	var page Page
// 	stmt := `
// 	SELECT * FROM pages
// 	WHERE id = $1
// 	LIMIT 1;
// 	`
// 	if err := s.DB.Get(&page, stmt, ID); err != nil {
// 		return page, errors.Wrap(err, "error querying database")
// 	}
// 	return page, nil
// }

// // PageByBookAndNumber returns a page by book id and number
// func (r *RedisCache) PageByBookAndNumber(bookID uuid.UUID, number int) (Page, error) {
// 	var page Page
// 	stmt := `
// 	SELECT * FROM pages
// 	WHERE book = $1
// 	AND
// 	page_number = $2
// 	LIMIT 1;`
// 	if err := s.DB.Get(&page, stmt, bookID, number); err != nil {
// 		return page, errors.Wrap(err, "error querying database")
// 	}
// 	return page, nil
// }

// // Authors fetches a list of authors
// func (r *RedisCache) Authors(limit, offset int) ([]Author, error) {
// 	authors := make([]Author, 0)
// 	var lim int
// 	if limit == -1 {
// 		lim = 1000
// 	} else {
// 		lim = limit
// 	}
// 	var off int
// 	if offset == -1 || offset == 0 {
// 		off = 0
// 	} else {
// 		off = offset
// 	}
// 	stmt := `SELECT * FROM authors ORDER BY name LIMIT $1 OFFSET $2;`
// 	rows, err := s.DB.Queryx(stmt, lim, off)

// 	if err != nil {
// 		return nil, errors.Wrap(err, "database query failed")
// 	}

// 	for rows.Next() {
// 		var author Author
// 		if err = rows.StructScan(&author); err != nil {
// 			return nil, errors.Wrap(err, "error scanning database rows")
// 		}
// 		authors = append(authors, author)
// 	}

// 	return authors, nil
// }

// // AuthorByID fetches an auhtor by ID
// func (r *RedisCache) AuthorByID(ID uuid.UUID) (Author, error) {
// 	var author Author
// 	stmt := `
// 	SELECT * FROM authors
// 	WHERE id = $1
// 	LIMIT 1;
// 	`

// 	if err := s.DB.Get(&author, stmt, ID); err != nil {
// 		return author, errors.Wrap(err, "error querying database")
// 	}
// 	return author, nil
// }

// // AuthorBySlug fetches an author by slug
// func (r *RedisCache) AuthorBySlug(slug string) (Author, error) {
// 	var author Author
// 	stmt := `
// 	SELECT * FROM authors
// 	WHERE slug = $1
// 	LIMIT 1;
// 	`
// 	err := s.DB.Get(&author, stmt, slug)
// 	if err != nil {
// 		return author, errors.Wrap(err, "error querying database")
// 	}
// 	return author, nil
// }


// func (r *RedisCache) InsertBook(book Book) (*Book, error) {
// 	var book Book
// 	if !r.IsAvailable() {
// 		return book, errors.New("attempted to access unavailable redis cache")
// 	}
// 	conn, err := r.Pool.Get()
// 	defer conn.Close()

// 	var pubYear int64
// 	if book.PublicationYear.Valid {
// 		pubYear = book.PublicationYear.Int64
// 	}

// 	_, err := conn.Do(
// 		"HMSET",
// 		"book:"+book.ID.String(),
// 		"title", book.Title,
// 		"slug", book.Slug,
// 		"publication_year", book.PublicationYear
// 	)



// }