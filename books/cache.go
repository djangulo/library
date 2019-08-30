package books

import (
	"encoding/json"
	"fmt"
	config "github.com/djangulo/library/config/books"
	"github.com/go-redis/redis"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"log"
	"strconv"
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

// BookByID fetches a book by ID
func (r *RedisCache) BookByID(
	book *Book,
	ID *uuid.UUID,
	fields []string,
) error {
	match := fmt.Sprintf("book:*:%s", ID.String())
	keys := r.Client.Scan(0, match, 0).Iterator()
	for keys.Next() {
		key := keys.Val()
		availableFields, err := r.Client.HKeys(key).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HKeys key %v", key))
		}
		if err := IsSubset(fields, availableFields); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf(
					"missing requested fields in key %v, has: %v, want %v",
					key,
					availableFields,
					fields,
				),
			)
		}
		result, err := r.Client.HMGet(key, fields...).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HMGET key: %v", key))
		}
		err = UnhashBook(book, fields, result)
		if err != nil {
			return errors.Wrap(err, "could parse HMGet results")
		}
	}
	return nil
}

// BookBySlug fetches a book by ID
func (r *RedisCache) BookBySlug(
	book *Book,
	slug *string,
	fields []string,
) error {
	*slug = Slugify(*slug, "-")
	match := fmt.Sprintf("book:%s:*", *slug)
	keys := r.Client.Scan(0, match, 0).Iterator()
	for keys.Next() {
		key := keys.Val()
		availableFields, err := r.Client.HKeys(key).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HKeys key %v", key))
		}
		if err := IsSubset(fields, availableFields); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf(
					"missing requested fields in key %v, has: %v, want %v",
					key,
					availableFields,
					fields,
				),
			)
		}
		result, err := r.Client.HMGet(key, fields...).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HMGET key: %v", key))
		}
		err = UnhashBook(book, fields, result)
		if err != nil {
			return errors.Wrap(err, "could parse HMGet results")
		}
	}
	return nil
}

// PageByID fetches a page by ID
func (r *RedisCache) PageByID(
	page *Page,
	ID *uuid.UUID,
	fields []string,
) error {
	match := fmt.Sprintf("page:%s:*", ID.String())
	keys := r.Client.Scan(0, match, 0).Iterator()
	for keys.Next() {
		key := keys.Val()
		availableFields, err := r.Client.HKeys(key).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HKeys key %v", key))
		}
		if err := IsSubset(fields, availableFields); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf(
					"missing requested fields in key %v, has: %v, want %v",
					key,
					availableFields,
					fields,
				),
			)
		}
		result, err := r.Client.HMGet(key, fields...).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HMGET key: %v", key))
		}
		err = UnhashPage(page, fields, result)
		if err != nil {
			return errors.Wrap(err, "could parse HMGet results")
		}
	}
	return nil
}

// PageByBookAndNumber returns a page by book id and number
func (r *RedisCache) PageByBookAndNumber(
	page *Page,
	bookID *uuid.UUID,
	number *int,
	fields []string,
) error {
	match := fmt.Sprintf("page:*:%s:%d", bookID.String(), *number)
	keys := r.Client.Scan(0, match, 0).Iterator()
	for keys.Next() {
		key := keys.Val()
		availableFields, err := r.Client.HKeys(key).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HKeys key %v", key))
		}
		if err := IsSubset(fields, availableFields); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf(
					"missing requested fields in key %v, has: %v, want %v",
					key,
					availableFields,
					fields,
				),
			)
		}
		result, err := r.Client.HMGet(key, fields...).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HMGET key: %v", key))
		}
		err = UnhashPage(page, fields, result)
		if err != nil {
			return errors.Wrap(err, "could parse HMGet results")
		}
	}
	return nil
}

// AuthorByID fetches a book by ID
func (r *RedisCache) AuthorByID(
	author *Author,
	ID *uuid.UUID,
	fields []string,
) error {
	match := fmt.Sprintf("author:*:%s", ID.String())
	keys := r.Client.Scan(0, match, 0).Iterator()
	for keys.Next() {
		key := keys.Val()
		availableFields, err := r.Client.HKeys(key).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HKeys key %v", key))
		}
		if err := IsSubset(fields, availableFields); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf(
					"missing requested fields in key %v, has: %v, want %v",
					key,
					availableFields,
					fields,
				),
			)
		}
		result, err := r.Client.HMGet(key, fields...).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HMGET key: %v", key))
		}
		err = UnhashAuthor(author, fields, result)
		if err != nil {
			return errors.Wrap(err, "could parse HMGet results")
		}
	}
	return nil
}

// AuthorBySlug fetches a book by ID
func (r *RedisCache) AuthorBySlug(
	author *Author,
	slug *string,
	fields []string,
) error {
	*slug = Slugify(*slug, "-")
	match := fmt.Sprintf("author:%s:*", *slug)
	keys := r.Client.Scan(0, match, 0).Iterator()
	for keys.Next() {
		key := keys.Val()
		availableFields, err := r.Client.HKeys(key).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HKeys key %v", key))
		}
		if err := IsSubset(fields, availableFields); err != nil {
			return errors.Wrap(
				err,
				fmt.Sprintf(
					"missing requested fields in key %v, has: %v, want %v",
					key,
					availableFields,
					fields,
				),
			)
		}
		result, err := r.Client.HMGet(key, fields...).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not HMGET key: %v", key))
		}
		err = UnhashAuthor(author, fields, result)
		if err != nil {
			return errors.Wrap(err, "could parse HMGet results")
		}
	}
	return nil
}

// InsertBook inserts book into the cache
func (r *RedisCache) InsertBook(book *Book) error {
	hash, _ := HashBook(book)
	cacheID, err := serializeBookID(book)
	if err != nil {
		return errors.Wrap(err, "could not serialize book cache ID")
	}
	err = r.Client.HMSet(cacheID, hash).Err()
	if err != nil {
		return errors.Wrap(err, "could not HMSet book")
	}
	return nil
}

// InsertAuthor inserts author into the cache
func (r *RedisCache) InsertAuthor(author *Author) error {
	hash, _ := HashAuthor(author)
	cacheID, err := serializeAuthorID(author)
	if err != nil {
		return errors.Wrap(err, "could not serialize author cache ID")
	}
	err = r.Client.HMSet(cacheID, hash).Err()
	if err != nil {
		return errors.Wrap(err, "could not HMSet author")
	}
	return nil
}

// InsertPage inserts page into the cache
func (r *RedisCache) InsertPage(page *Page) error {
	hash, _ := HashPage(page)
	cacheID, err := serializePageID(page)
	if err != nil {
		return errors.Wrap(err, "could not serialize page cache ID")
	}
	err = r.Client.HMSet(cacheID, hash).Err()
	if err != nil {
		return errors.Wrap(err, "could not HMSet page")
	}
	return nil
}

// SaveBookQuery saves a query onto the cache for easy retrieval
func (r *RedisCache) SaveBookQuery(key string, books []*Book) error {
	b, err := json.Marshal(&books)
	if err != nil {
		return errors.Wrap(err, "could not marshal array of type []*Book")
	}
	err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET book query")
	}
	return nil
}

// BookQuery retrieves a saved  query from the cache
func (r *RedisCache) BookQuery(books *[]*Book, key string) error {
	str, err := r.Client.Get(key).Result()
	if err != nil {
		return errors.Wrap(err, "could not GET book query")
	}

	b := []byte(str)
	err = json.Unmarshal(b, &books)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal books JSON")
	}

	return nil
}

// SavePageQuery saves a query onto the cache for easy retrieval
func (r *RedisCache) SavePageQuery(key string, pages []*Page) error {
	b, err := json.Marshal(&pages)
	if err != nil {
		return errors.Wrap(err, "could not marshal array of type []*Page")
	}
	err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	if err != nil {
		return errors.Wrap(err, "could not SET page query")
	}
	return nil
}

// PageQuery retrieves a saved  query from the cache
func (r *RedisCache) PageQuery(pages *[]*Page, key string) error {
	str, err := r.Client.Get(key).Result()
	if err != nil {
		return errors.Wrap(err, "could not GET pages")
	}

	b := []byte(str)
	err = json.Unmarshal(b, &pages)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal pages JSON")
	}

	return nil
}

// SaveAuthorQuery saves a query onto the cache for easy retrieval
func (r *RedisCache) SaveAuthorQuery(key string, authors []*Author) error {
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

// AuthorQuery retrieves a saved  query from the cache
func (r *RedisCache) AuthorQuery(authors *[]*Author, key string) error {
	str, err := r.Client.Get(key).Result()
	if err != nil {
		return errors.Wrap(err, "could not GET Authors")
	}

	b := []byte(str)
	err = json.Unmarshal(b, &authors)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal authors JSON")
	}

	return nil
}

// HashBook safely hash a book into a Redis HMSet compatible map
func HashBook(book *Book) (map[string]interface{}, error) {
	b := map[string]interface{}{
		"id":         book.ID.String(),
		"title":      book.Title,
		"slug":       book.Slug,
		"page_count": strconv.Itoa(book.PageCount),
		"created_at": book.CreatedAt.Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"deleted_at": book.DeletedAt.Format(time.RFC3339),
	}
	if book.PublicationYear.Valid {
		b["publication_year"] = strconv.Itoa(int(book.PublicationYear.Int64))
	}
	if book.File.Valid {
		b["file"] = book.File.String
	}
	if book.Source.Valid {
		b["source"] = book.Source.String
	}
	if book.AuthorID.Valid {
		b["author_id"] = book.AuthorID.UUID.String()
	}
	return b, nil
}

// HashPage safely hash a page into a Redis HMSet compatible map
func HashPage(page *Page) (map[string]interface{}, error) {
	return map[string]interface{}{
		"id":          page.ID.String(),
		"page_number": page.PageNumber,
		"book_id":     page.BookID.String(),
		"body":        page.Body,
		"created_at":  page.CreatedAt.Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
		"deleted_at":  page.DeletedAt.Format(time.RFC3339),
	}, nil
}

// HashAuthor safely hash an author into a Redis HMSet compatible map
func HashAuthor(author *Author) (map[string]interface{}, error) {
	return map[string]interface{}{
		"id":         author.ID.String(),
		"name":       author.Name,
		"slug":       author.Slug,
		"created_at": author.CreatedAt.Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"deleted_at": author.DeletedAt.Format(time.RFC3339),
	}, nil
}

// UnhashBook unhash the HMGet.Results interface into a book
func UnhashBook(
	book *Book,
	fields []string,
	results []interface{},
) error {
	for i, r := range results {
		switch v := r.(type) {
		case string:
			switch fields[i] {
			case "title":
				book.Title = v
			case "slug":
				book.Slug = v
			case "page_count":
				inty, err := strconv.Atoi(v)
				if err != nil {
					return errors.Wrap(err, "could not convert page count to int")
				}
				book.PageCount = inty
			case "id":
				uid, err := uuid.FromString(v)
				if err != nil {
					return errors.Wrap(err, "could not convert id to UUID")
				}
				book.ID = uid
			case "publication_year":
				inty, err := strconv.Atoi(v)
				if err != nil {
					return errors.Wrap(err, "could not convert page count to int")
				}
				book.PublicationYear = NewNullInt64(int64(inty))
			case "file":
				book.File = NewNullString(v)
			case "source":
				book.Source = NewNullString(v)
			case "author_id":
				book.AuthorID = NewNullUUID(v)
			case "created_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				book.CreatedAt = t
			case "updated_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				book.UpdatedAt = t
			case "deleted_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				book.DeletedAt = t
			}
		}
	}
	return nil
}

// UnhashPage unhash the HMGet.Results interface into a page
func UnhashPage(
	page *Page,
	fields []string,
	results []interface{},
) error {
	for i, r := range results {
		switch v := r.(type) {
		case string:
			switch fields[i] {
			case "id":
				uid, err := uuid.FromString(v)
				if err != nil {
					return errors.Wrap(err, "could not convert id to UUID")
				}
				page.ID = uid
			case "book_id":
				uid, err := uuid.FromString(v)
				if err != nil {
					return errors.Wrap(err, "could not convert id to UUID")
				}
				page.BookID = &uid
			case "page_number":
				inty, err := strconv.Atoi(v)
				if err != nil {
					return errors.Wrap(err, "could not convert page count to int")
				}
				page.PageNumber = inty
			case "body":
				page.Body = v
			case "created_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				page.CreatedAt = t
			case "updated_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				page.UpdatedAt = t
			case "deleted_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				page.DeletedAt = t
			}
		}
	}
	return nil
}

// UnhashAuthor unhash the HMGet.Results interface into a author
func UnhashAuthor(
	author *Author,
	fields []string,
	results []interface{},
) error {
	for i, r := range results {
		switch v := r.(type) {
		case string:
			switch fields[i] {
			case "id":
				uid, err := uuid.FromString(v)
				if err != nil {
					return errors.Wrap(err, "could not convert id to UUID")
				}
				author.ID = uid
			case "slug":
				author.Slug = v
			case "name":
				author.Name = v
			case "created_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				author.CreatedAt = t
			case "updated_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				author.UpdatedAt = t
			case "deleted_at":
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return errors.Wrap(err, "could not parse timestamp onto time.Time")
				}
				author.DeletedAt = t
			}
		}
	}
	return nil
}

func serializeBookID(book *Book) (string, error) {
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

func serializeAuthorID(author *Author) (string, error) {
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

func serializePageID(page *Page) (string, error) {
	if page.ID == uuid.Nil ||
		*page.BookID == uuid.Nil ||
		page.PageNumber == 0 {
		return "", errors.New(fmt.Sprintf("invalid page ID: %v", page.ID))
	}
	return fmt.Sprintf(
		"page:%s:%s:%d",
		page.ID.String(),
		page.BookID.String(),
		page.PageNumber,
	), nil
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
