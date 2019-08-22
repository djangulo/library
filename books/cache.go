package books

import (
	"encoding/json"
	"fmt"
	config "github.com/djangulo/library/config/books"
	"github.com/go-redis/redis"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"log"
	"strings"
	"time"
)

// RedisCache cache layer
type RedisCache struct {
	Available bool
	Client    *redis.Client
}

var (
	// ErrCacheUnavailable returned if cache is unavailable
	ErrCacheUnavailable = errors.New("attempted to access unavailable redis cache")
	// ErrNotFoundInCache returned if the query is not available
	ErrNotFoundInCache = errors.New("key not found in cache")
)

// Conn returns a connection from the pool and a convenience close method

// IsAvailable checks whether a redis conection was made available on init
func (r *RedisCache) IsAvailable() error {
	fmt.Printf("\nfrom cache.IsAvailable(): %+v\n", r)
	if !r.Available {
		return ErrCacheUnavailable
	}
	return nil
}

// NewRedisCache returns a `*RedisCache` object with the config provided
func NewRedisCache(config config.CacheConfig) (*RedisCache, error) {
	connStr := config.ConnStr()
	client := redis.NewClient(&redis.Options{
		Addr:     connStr,
		Password: config.Password,
		DB:       config.DB,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("Redis connection unavailable: %v", err)
		return &RedisCache{Available: false}, errors.Wrap(
			err,
			"redis connection unavailable",
		)
	}
	return &RedisCache{
		Available: true,
		Client:    client,
	}, nil
}

// Books fetches a list of books from the cache, offset is ignored
func (r *RedisCache) Books(limit, offset int) ([]Book, error) {
	var lim int
	if limit == -1 {
		lim = 1000
	} else {
		lim = limit
	}
	books := make([]Book, 0)
	keys := r.Client.Scan(0, "book:*", int64(lim)).Iterator()
	for keys.Next() {
		strBytes, err := r.Client.Get(keys.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", keys.Val()))
		}

		var book Book
		err = json.Unmarshal([]byte(strBytes), &book)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", keys.Val()))
		}
		books = append(books, book)
	}
	return books, nil
}

// BookByID fetches a book by ID
func (r *RedisCache) BookByID(ID uuid.UUID) (Book, error) {

	bookStr, err := r.Client.Get("book:" + ID.String()).Result()
	if err != nil {
		return Book{}, errors.Wrap(err, "could not GET book")
	}

	b := []byte(bookStr)

	var book Book
	err = json.Unmarshal(b, &book)
	if err != nil {
		return Book{}, errors.Wrap(err, "could not unmarshal book JSON")
	}

	return book, nil
}

// BookBySlug fetches a book by slug
func (r *RedisCache) BookBySlug(slug string) (Book, error) {
	keys := r.Client.Scan(0, "book:*", 0).Iterator()
	for keys.Next() {
		var book Book
		strBytes, _ := r.Client.Get(keys.Val()).Result()

		json.Unmarshal([]byte(strBytes), &book)
		if book.Slug == slug {
			return book, nil
		}
	}
	return Book{}, nil
}

// BooksByAuthor returns books by a given author
func (r *RedisCache) BooksByAuthor(name string) ([]Book, error) {
	authorKeys := r.Client.Scan(0, "author:*", 0).Iterator()
	var author Author
	for authorKeys.Next() {
		strBytes, err := r.Client.Get(authorKeys.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", authorKeys.Val()))
		}

		var auteur Author
		err = json.Unmarshal([]byte(strBytes), &auteur)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", authorKeys.Val()))
		}
		if strings.Contains(strings.ToLower(auteur.Name), strings.ToLower(name)) {
			author = auteur
			break
		}
	}

	if author.ID == uuid.Nil {
		// Author not in cache, let the store handle it
		return nil, ErrNotFoundInCache
	}
	// author exists in cache, try to find the books
	books := make([]Book, 0)
	bookKeys := r.Client.Scan(0, "book:*", 0).Iterator()
	for bookKeys.Next() {
		strBytes, err := r.Client.Get(bookKeys.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", bookKeys.Val()))
		}

		var book Book
		err = json.Unmarshal([]byte(strBytes), &book)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", bookKeys.Val()))
		}
		if book.AuthorID.Valid {
			if book.AuthorID.UUID.String() == author.ID.String() {
				books = append(books, book)
			}
		}
	}
	return books, nil
}

// Pages fetches a list of pages, offset is ignored
func (r *RedisCache) Pages(limit, offset int) ([]Page, error) {
	var lim int
	if limit == -1 {
		lim = 1000
	} else {
		lim = limit
	}
	pages := make([]Page, 0)
	keys := r.Client.Scan(0, "book:*", int64(lim)).Iterator()
	for keys.Next() {
		strBytes, err := r.Client.Get(keys.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", keys.Val()))
		}

		var page Page
		err = json.Unmarshal([]byte(strBytes), &page)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", keys.Val()))
		}
		pages = append(pages, page)
	}
	return pages, nil
}

// PageByID fetches a page by ID
func (r *RedisCache) PageByID(ID uuid.UUID) (Page, error) {

	str, err := r.Client.Get("page:" + ID.String()).Result()
	if err != nil {
		return Page{}, errors.Wrap(err, "could not GET page")
	}

	b := []byte(str)

	var page Page
	err = json.Unmarshal(b, &page)
	if err != nil {
		return Page{}, errors.Wrap(err, "could not unmarshal page JSON")
	}

	return page, nil
}

// PageByBookAndNumber returns a page by book id and number
func (r *RedisCache) PageByBookAndNumber(bookID uuid.UUID, number int) (Page, error) {
	pageKeys := r.Client.Scan(0, "page:*", 0).Iterator()
	for pageKeys.Next() {
		strBytes, err := r.Client.Get(pageKeys.Val()).Result()
		if err != nil {
			return Page{}, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", pageKeys.Val()))
		}

		var page Page
		err = json.Unmarshal([]byte(strBytes), &page)
		if err != nil {
			return Page{}, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", pageKeys.Val()))
		}
		if page.BookID.String() == bookID.String() && page.PageNumber == number {
			return page, nil
		}
	}
	return Page{}, ErrNotFoundInCache
}

// Authors fetches a list of authors from the cache, offset is ignored
func (r *RedisCache) Authors(limit, offset int) ([]Author, error) {
	var lim int
	if limit == -1 {
		lim = 1000
	} else {
		lim = limit
	}
	authors := make([]Author, 0)
	keys := r.Client.Scan(0, "author:*", int64(lim)).Iterator()
	for keys.Next() {
		strBytes, err := r.Client.Get(keys.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", keys.Val()))
		}

		var author Author
		err = json.Unmarshal([]byte(strBytes), &author)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", keys.Val()))
		}
		authors = append(authors, author)
	}
	return authors, nil
}

// AuthorByID fetches an auhtor by ID
func (r *RedisCache) AuthorByID(ID uuid.UUID) (Author, error) {
	str, err := r.Client.Get("author:" + ID.String()).Result()
	if err != nil {
		return Author{}, errors.Wrap(err, "could not GET author")
	}

	b := []byte(str)

	var author Author
	err = json.Unmarshal(b, &author)
	if err != nil {
		return Author{}, errors.Wrap(err, "could not unmarshal author JSON")
	}

	return author, nil
}

// AuthorBySlug fetches an author by slug
func (r *RedisCache) AuthorBySlug(slug string) (Author, error) {
	keys := r.Client.Scan(0, "author:*", 0).Iterator()
	for keys.Next() {
		var author Author
		strBytes, _ := r.Client.Get(keys.Val()).Result()

		json.Unmarshal([]byte(strBytes), &author)
		if author.Slug == slug {
			return author, nil
		}
	}
	return Author{}, nil
}

// InsertBook inserts book into the cache
func (r *RedisCache) InsertBook(book Book) error {

	b, err := json.Marshal(&book)
	if err != nil {
		return errors.Wrap(err, "could not marshal book")
	}

	err = r.Client.Set("book:"+book.ID.String(), string(b), 24*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET book")
	}

	return nil
}

// func MapBytesToBook(bytes [][]byte) *Book {
// 	var book Book
// 	for i := 0; i < len(bytes); i += 2 {
// 		k := string(bytes[i])
// 		switch k {
// 		case "title":
// 			book.Title = string(bytes[i+1])
// 		case "slug":
// 			book.Slug = string(bytes[i+1])
// 		case "publication_year":
// 			inty, _ := strconv.Atoi(string(bytes[i+1]))
// 			book.PublicationYear = NewNullInt64(int64(inty))
// 		case "page_count":
// 			inty, _ := strconv.Atoi(string(bytes[i+1]))
// 			book.PageCount = inty
// 		case "author_id":
// 			fmt.Printf("%T", bytes[i+1])
// 			fmt.Printf("%T", string(bytes[i+1]))
// 			book.AuthorID = NewNullUUID(string(bytes[i+1]))
// 		case "file":
// 			book.File = NewNullString(string(bytes[i+1]))
// 		case "source":
// 			book.Source = NewNullString(string(bytes[i+1]))
// 		case "id":
// 			book.ID = uuid.Must(uuid.FromString(string(bytes[i+1])))
// 		}
// 	}
// 	return &book
// }
