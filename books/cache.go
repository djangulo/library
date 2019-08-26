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

// IsAvailable checks whether a redis conection was made available on init
func (r *RedisCache) IsAvailable() error {
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
	keyMatch := fmt.Sprintf("book:*:%s", ID.String())
	keys := r.Client.Scan(0, keyMatch, 0).Iterator()
	for keys.Next() {
		_, id, err := parseBookID(keys.Val())
		if err != nil {
			return Book{}, errors.Wrap(err, fmt.Sprintf("could not parse book cache ID: %v", keys.Val()))
		}
		if id == ID {
			strBytes, err := r.Client.Get(keys.Val()).Result()
			if err != nil {
				return Book{}, errors.Wrap(err, fmt.Sprintf("could not GET id: %v", keys.Val()))
			}
			var book Book
			err = json.Unmarshal([]byte(strBytes), &book)
			if err != nil {
				return Book{}, errors.Wrap(err, fmt.Sprintf("could not unmarshal json into instance of Book, json: %v", strBytes))
			}
			return book, nil
		}

	}
	return Book{}, nil

}

// BookBySlug fetches a book by slug
func (r *RedisCache) BookBySlug(slug string) (Book, error) {
	keyMatch := fmt.Sprintf("book:%s:*", slug)
	keys := r.Client.Scan(0, keyMatch, 0).Iterator()
	for keys.Next() {
		bookSlug, _, err := parseBookID(keys.Val())
		if err != nil {
			return Book{}, errors.Wrap(err, fmt.Sprintf("could not parse book cache ID: %v", keys.Val()))
		}
		if bookSlug == slug {
			strBytes, err := r.Client.Get(keys.Val()).Result()
			if err != nil {
				return Book{}, errors.Wrap(err, fmt.Sprintf("could not GET id: %v", keys.Val()))
			}
			var book Book
			err = json.Unmarshal([]byte(strBytes), &book)
			if err != nil {
				return Book{}, errors.Wrap(err, fmt.Sprintf("could not unmarshal json into instance of Book, json: %v", strBytes))
			}
			return book, nil
		}

	}
	return Book{}, nil
}

// BooksByAuthor returns books by a given author
func (r *RedisCache) BooksByAuthor(name string) ([]Book, error) {
	name = Slugify(name, "-")

	var author Author
	authorKeyMatch := fmt.Sprintf("author:%s:*", name)
	authorKeys := r.Client.Scan(0, authorKeyMatch, 0).Iterator()
	for authorKeys.Next() {
		key := authorKeys.Val()
		slug, _, err := parseAuthorID(key)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not parse author cache ID: %v", key))
		}
		if strings.Contains(slug, name) {
			strBytes, err := r.Client.Get(key).Result()
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", authorKeys.Val()))
			}
			auteur := new(Author)
			err = json.Unmarshal([]byte(strBytes), &auteur)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", authorKeys.Val()))
			}
			author = *auteur
		}
	}
	if author.ID == uuid.Nil {
		// Author not in cache, let the store handle it
		return nil, ErrNotFoundInCache
	}

	books := make([]Book, 0)
	bookKeys := r.Client.Scan(0, "book:*", 0).Iterator()
	for bookKeys.Next() {
		strBytes, err := r.Client.Get(bookKeys.Val()).Result()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", bookKeys.Val()))
		}

		book := new(Book)
		err = json.Unmarshal([]byte(strBytes), &book)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", bookKeys.Val()))
		}
		if book.AuthorID.Valid {
			if book.AuthorID.UUID == author.ID {
				books = append(books, *book)
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
	keys := r.Client.Scan(0, "page:*", int64(lim)).Iterator()
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
	authorKeyMatch := fmt.Sprintf("author:*:%s", ID.String())
	keys := r.Client.Scan(0, authorKeyMatch, 0).Iterator()
	for keys.Next() {
		_, id, err := parseBookID(keys.Val())
		if err != nil {
			return Author{}, errors.Wrap(err, fmt.Sprintf("could not parse author cache ID: %v", keys.Val()))
		}
		if id == ID {
			strBytes, err := r.Client.Get(keys.Val()).Result()
			if err != nil {
				return Author{}, errors.Wrap(err, fmt.Sprintf("could not GET id: %v", keys.Val()))
			}
			var author Author
			err = json.Unmarshal([]byte(strBytes), &author)
			if err != nil {
				return Author{}, errors.Wrap(err, fmt.Sprintf("could not unmarshal json into instance of Author, json: %v", strBytes))
			}
			return author, nil
		}

	}
	return Author{}, nil
}

// AuthorBySlug fetches an author by slug
func (r *RedisCache) AuthorBySlug(slug string) (Author, error) {
	slug = Slugify(slug, "-")

	authorKeyMatch := fmt.Sprintf("author%s:*", slug)
	authorKeys := r.Client.Scan(0, authorKeyMatch, 0).Iterator()
	for authorKeys.Next() {
		key := authorKeys.Val()
		authorSlug, _, err := parseAuthorID(key)
		if err != nil {
			log.Println(errors.Wrap(err, fmt.Sprintf("could not parse author cache ID: %v", key)))
			continue
		}
		if strings.Contains(authorSlug, slug) {
			strBytes, err := r.Client.Get(key).Result()
			if err != nil {
				log.Println(errors.Wrap(err, fmt.Sprintf("could not scan key: %v", key)))
				return Author{}, errors.Wrap(err, fmt.Sprintf("could not scan key: %v", key))
			}
			var author Author
			err = json.Unmarshal([]byte(strBytes), &author)
			if err != nil {
				log.Println(errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", key)))
				return Author{}, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", key))
			}
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
	bookCacheID, err := serializeBookID(book)
	if err != nil {
		return errors.Wrap(err, "could not serialize book cache ID")
	}

	err = r.Client.Set(bookCacheID, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET book")
	}

	return nil
}

// InsertAuthor inserts author into the cache
func (r *RedisCache) InsertAuthor(author Author) error {

	b, err := json.Marshal(&author)
	if err != nil {
		return errors.Wrap(err, "could not marshal author")
	}
	authorCacheID, err := serializeAuthorID(author)
	if err != nil {
		return errors.Wrap(err, "could not serialize author cache ID")
	}

	err = r.Client.Set(authorCacheID, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET author")
	}

	return nil
}

// InsertPage inserts book into the cache
func (r *RedisCache) InsertPage(page Page) error {

	b, err := json.Marshal(&page)
	if err != nil {
		return errors.Wrap(err, "could not marshal page")
	}
	pageCacheID, err := serializePageID(page)
	if err != nil {
		return errors.Wrap(err, "could not serialize page cache ID")
	}

	err = r.Client.Set(pageCacheID, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET page")
	}

	return nil
}

// SaveBookQuery saves a query onto the cache for easy retrieval
func (r *RedisCache) SaveBookQuery(key string, books []Book) error {
	b, err := json.Marshal(&books)
	if err != nil {
		return errors.Wrap(err, "could not marshal array of type Book")
	}
	err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET book query")
	}
	return nil
}

// GetBookQuery retrieves a saved  query from the cache
func (r *RedisCache) GetBookQuery(key string) ([]Book, error) {
	str, err := r.Client.Get(key).Result()
	if err != nil {
		return nil, errors.Wrap(err, "could not GET author")
	}

	b := []byte(str)
	var books []Book
	err = json.Unmarshal(b, &books)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal books JSON")
	}

	return books, nil
}

// SavePageQuery saves a query onto the cache for easy retrieval
func (r *RedisCache) SavePageQuery(key string, pages []Page) error {
	b, err := json.Marshal(&pages)
	if err != nil {
		return errors.Wrap(err, "could not marshal array of type Page")
	}
	err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET page query")
	}
	return nil
}

// GetPageQuery retrieves a saved  query from the cache
func (r *RedisCache) GetPageQuery(key string) ([]Page, error) {
	str, err := r.Client.Get(key).Result()
	if err != nil {
		return nil, errors.Wrap(err, "could not GET pages")
	}

	b := []byte(str)
	var pages []Page
	err = json.Unmarshal(b, &pages)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal pages JSON")
	}

	return pages, nil
}

// SaveAuthorQuery saves a query onto the cache for easy retrieval
func (r *RedisCache) SaveAuthorQuery(key string, authors []Author) error {
	b, err := json.Marshal(&authors)
	if err != nil {
		return errors.Wrap(err, "could not marshal array of type Author")
	}
	err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET book query")
	}
	return nil
}

// GetAuthorQuery retrieves a saved  query from the cache
func (r *RedisCache) GetAuthorQuery(key string) ([]Author, error) {
	str, err := r.Client.Get(key).Result()
	if err != nil {
		return nil, errors.Wrap(err, "could not GET Authors")
	}

	b := []byte(str)
	var authors []Author
	err = json.Unmarshal(b, &authors)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal authors JSON")
	}

	return authors, nil
}

func serializeBookID(book Book) (string, error) {
	if book.ID == uuid.Nil {
		return "", errors.New(fmt.Sprintf("invalid book ID: %v", book.ID))
	}
	return fmt.Sprintf("book:%s:%s", book.Slug, book.ID.String()), nil
}

func parseBookID(bookCacheID string) (string, uuid.UUID, error) {
	arr := strings.Split(bookCacheID, ":")
	slug, idString := arr[1], arr[2]
	uid, err := uuid.FromString(idString)
	if err != nil {
		return "", uuid.Nil, errors.Wrap(err, "could not parse uuid")
	}
	return slug, uid, nil
}

func serializeAuthorID(author Author) (string, error) {
	if author.ID == uuid.Nil {
		return "", errors.New(fmt.Sprintf("invalid author ID: %v", author.ID))
	}
	return fmt.Sprintf("author:%s:%s", author.Slug, author.ID.String()), nil
}

func parseAuthorID(authorCacheID string) (string, uuid.UUID, error) {
	arr := strings.Split(authorCacheID, ":")
	slug, idString := arr[1], arr[2]
	uid, err := uuid.FromString(idString)
	if err != nil {
		return "", uuid.Nil, errors.Wrap(err, "could not parse uuid")
	}
	return slug, uid, nil
}

func serializePageID(page Page) (string, error) {
	if page.ID == uuid.Nil {
		return "", errors.New(fmt.Sprintf("invalid page ID: %v", page.ID))
	}
	return fmt.Sprintf("page:%s", page.ID.String()), nil
}

func parsePageID(pageCacheID string) (uuid.UUID, error) {
	arr := strings.Split(pageCacheID, ":")
	idString := arr[1]
	uid, err := uuid.FromString(idString)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "could not parse uuid")
	}
	return uid, nil
}
