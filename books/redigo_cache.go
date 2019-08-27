package books

import (
	// "encoding/json"
	"fmt"
	config "github.com/djangulo/library/config/books"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"log"
	"strconv"
	"time"
)

type RedigoCache struct {
	Available bool
	Pool      *redis.Pool
}

// IsAvailable checks whether a redis conection was made available on init
func (r *RedigoCache) IsAvailable() error {
	if !r.Available {
		return ErrCacheUnavailable
	}
	return nil
}

func NewRedigoCache(config config.CacheConfig) (*RedigoCache, error) {
	connStr := config.ConnStr()
	conn, err := redis.Dial("tcp", connStr)
	defer conn.Close()
	if err != nil {
		log.Printf("Redis connection unavailable: %v", err)
		return &RedigoCache{Available: false}, errors.Wrap(
			err,
			"redis connection unavailable",
		)
	}
	return &RedigoCache{
		Available: true,
		Pool: &redis.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", connStr)
			},
		},
	}, nil
}

func (r *RedigoCache) GetKeys(pattern string) ([]string, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, errors.Wrap(
				err,
				fmt.Sprintf(
					"error retrieving '%s' keys",
					pattern,
				),
			)
		}

		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return keys, errors.Wrap(err, "could not get new cursor")
		}
		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return keys, errors.Wrap(err, "could not get parse string")
		}
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}

// Books fetches a list of books from the cache, offset is ignored
func (r *RedigoCache) Books(limit, offset int) ([]Book, error) {
	// var lim int
	// if limit == -1 {
	// 	lim = 1000
	// } else {
	// 	lim = limit
	// }
	books := make([]Book, 0)
	// keys := r.Client.Scan(0, "book:*", int64(lim)).Iterator()
	// for keys.Next() {
	// 	strBytes, err := r.Client.Get(keys.Val()).Result()
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", keys.Val()))
	// 	}

	// 	var book Book
	// 	err = json.Unmarshal([]byte(strBytes), &book)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", keys.Val()))
	// 	}
	// 	books = append(books, book)
	// }
	return books, nil
}

// BookByID fetches a book by ID
func (r *RedigoCache) BookByID(ID uuid.UUID) (Book, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	pattern := fmt.Sprintf("book:*:%s", ID.String())
	keys, err := r.GetKeys(pattern)
	if err != nil {
		return Book{}, errors.Wrap(
			err,
			fmt.Sprintf(
				"could not GetKeys with pattern %v",
				pattern,
			),
		)
	}
	if len(keys) > 1 {
		return Book{}, errors.New("multiple results from ID")
	}

	values, err := redis.Values(conn.Do("HGETALL", keys[0]))
	if err != nil {
		return Book{}, errors.Wrap(err, "could not HGETALL")
	}
	str, err := redis.Strings(values, nil)
	if err != nil {
		return Book{}, errors.Wrap(err, "could not stringify")
	}
	var book Book
	err = unhashBookFromStrings(str, &book)
	// err = redis.ScanStruct(values, &book)
	if err != nil {
		return Book{}, errors.Wrap(err, "could not unhash book")
	}
	return book, nil

}

// BookBySlug fetches a book by slug
func (r *RedigoCache) BookBySlug(slug string) (Book, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	slug = Slugify(slug, "-")
	pattern := fmt.Sprintf("book:%s:*", slug)
	keys, err := r.GetKeys(pattern)
	if err != nil {
		return Book{}, errors.Wrap(
			err,
			fmt.Sprintf(
				"could not GetKeys with pattern %v",
				pattern,
			),
		)
	}
	if len(keys) > 1 {
		return Book{}, errors.New("multiple results from ID")
	}

	values, err := redis.Values(conn.Do("HGETALL", keys[0]))
	if err != nil {
		return Book{}, errors.Wrap(err, "could not HGETALL")
	}
	str, err := redis.Strings(values, nil)
	if err != nil {
		return Book{}, errors.Wrap(err, "could not stringify")
	}
	var book Book
	err = unhashBookFromStrings(str, &book)
	// err = redis.ScanStruct(values, &book)
	if err != nil {
		return Book{}, errors.Wrap(err, "could not unhash book")
	}
	return book, nil
}

// BooksByAuthor returns books by a given author
func (r *RedigoCache) BooksByAuthor(name string) ([]Book, error) {

	// conn := r.Pool.Get()
	// defer conn.Close()

	// name = Slugify(name, "-")
	// pattern := fmt.Sprintf("author:%s:*", name)
	books := make([]Book, 0)
	// keys, err := r.GetKeys(pattern)
	// if err != nil {
	// 	return books, errors.Wrap(
	// 		err,
	// 		fmt.Sprintf(
	// 			"could not GetKeys with pattern %v",
	// 			pattern,
	// 		),
	// 	)
	// }

	// _, authorID, err := parseAuthorID(keys[0])
	// if err != nil {
	// 	return books, errors.Wrap(err, "could not parse author cache ID")
	// }

	// values, err := redis.Values(conn.Do("HGETALL", keys[0]))
	// if err != nil {
	// 	return Book{}, errors.Wrap(err, "could not HGETALL")
	// }
	// str, err := redis.Strings(values, nil)
	// if err != nil {
	// 	return Book{}, errors.Wrap(err, "could not stringify")
	// }
	// var book Book
	// err = unhashBookFromStrings(str, &book)
	// // err = redis.ScanStruct(values, &book)
	// if err != nil {
	// 	return Book{}, errors.Wrap(err, "could not unhash book")
	// }
	return books, nil

	// var author Author
	// authorKeyMatch := fmt.Sprintf("author:%s:*", name)
	// authorKeys := r.Client.Scan(0, authorKeyMatch, 0).Iterator()
	// for authorKeys.Next() {
	// 	key := authorKeys.Val()
	// 	slug, _, err := parseAuthorID(key)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("could not parse author cache ID: %v", key))
	// 	}
	// 	if strings.Contains(slug, name) {
	// 		strBytes, err := r.Client.Get(key).Result()
	// 		if err != nil {
	// 			return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", authorKeys.Val()))
	// 		}
	// 		auteur := new(Author)
	// 		err = json.Unmarshal([]byte(strBytes), &auteur)
	// 		if err != nil {
	// 			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", authorKeys.Val()))
	// 		}
	// 		author = *auteur
	// 	}
	// }
	// if author.ID == uuid.Nil {
	// 	// Author not in cache, let the store handle it
	// 	return nil, ErrNotFoundInCache
	// }

	// bookKeys := r.Client.Scan(0, "book:*", 0).Iterator()
	// for bookKeys.Next() {
	// 	strBytes, err := r.Client.Get(bookKeys.Val()).Result()
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", bookKeys.Val()))
	// 	}

	// 	book := new(Book)
	// 	err = json.Unmarshal([]byte(strBytes), &book)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", bookKeys.Val()))
	// 	}
	// 	if book.AuthorID.Valid {
	// 		if book.AuthorID.UUID == author.ID {
	// 			books = append(books, *book)
	// 		}
	// 	}
	// }
}

// Pages fetches a list of pages, offset is ignored
func (r *RedigoCache) Pages(limit, offset int) ([]Page, error) {
	// var lim int
	// if limit == -1 {
	// 	lim = 1000
	// } else {
	// 	lim = limit
	// }
	pages := make([]Page, 0)
	// keys := r.Client.Scan(0, "page:*", int64(lim)).Iterator()
	// for keys.Next() {
	// 	strBytes, err := r.Client.Get(keys.Val()).Result()
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", keys.Val()))
	// 	}

	// 	var page Page
	// 	err = json.Unmarshal([]byte(strBytes), &page)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", keys.Val()))
	// 	}
	// 	pages = append(pages, page)
	// }
	return pages, nil
}

// PageByID fetches a page by ID
func (r *RedigoCache) PageByID(ID uuid.UUID) (Page, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	pattern := fmt.Sprintf("page:%s", ID.String())
	keys, err := r.GetKeys(pattern)
	if err != nil {
		return Page{}, errors.Wrap(
			err,
			fmt.Sprintf(
				"could not GetKeys with pattern %v",
				pattern,
			),
		)
	}
	if len(keys) > 1 {
		return Page{}, errors.New("multiple results from ID")
	}

	values, err := redis.Values(conn.Do("HGETALL", keys[0]))
	if err != nil {
		return Page{}, errors.Wrap(err, "could not HGETALL")
	}
	str, err := redis.Strings(values, nil)
	if err != nil {
		return Page{}, errors.Wrap(err, "could not stringify")
	}
	var page Page
	err = unhashPageFromStrings(str, &page)
	// err = redis.ScanStruct(values, &book)
	if err != nil {
		return Page{}, errors.Wrap(err, "could not unhash page")
	}
	return page, nil
}

// PageByBookAndNumber returns a page by book id and number
func (r *RedigoCache) PageByBookAndNumber(bookID uuid.UUID, number int) (Page, error) {
	// pageKeys := r.Client.Scan(0, "page:*", 0).Iterator()
	// for pageKeys.Next() {
	// 	strBytes, err := r.Client.Get(pageKeys.Val()).Result()
	// 	if err != nil {
	// 		return Page{}, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", pageKeys.Val()))
	// 	}

	// 	var page Page
	// 	err = json.Unmarshal([]byte(strBytes), &page)
	// 	if err != nil {
	// 		return Page{}, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", pageKeys.Val()))
	// 	}
	// 	if page.BookID.String() == bookID.String() && page.PageNumber == number {
	// 		return page, nil
	// 	}
	// }
	return Page{}, ErrNotFoundInCache
}

// Authors fetches a list of authors from the cache, offset is ignored
func (r *RedigoCache) Authors(limit, offset int) ([]Author, error) {
	// var lim int
	// if limit == -1 {
	// 	lim = 1000
	// } else {
	// 	lim = limit
	// }
	authors := make([]Author, 0)

	// keys := r.Client.Scan(0, "author:*", int64(lim)).Iterator()
	// for keys.Next() {
	// 	strBytes, err := r.Client.Get(keys.Val()).Result()
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("could not scan key '%v'", keys.Val()))
	// 	}

	// 	var author Author
	// 	err = json.Unmarshal([]byte(strBytes), &author)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal key '%v'", keys.Val()))
	// 	}
	// 	authors = append(authors, author)
	// }
	return authors, nil
}

// AuthorByID fetches an auhtor by ID
func (r *RedigoCache) AuthorByID(ID uuid.UUID) (Author, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	pattern := fmt.Sprintf("author:*:%s", ID.String())
	keys, err := r.GetKeys(pattern)
	if err != nil {
		return Author{}, errors.Wrap(
			err,
			fmt.Sprintf(
				"could not GetKeys with pattern %v",
				pattern,
			),
		)
	}
	if len(keys) > 1 {
		return Author{}, errors.New("multiple results from ID")
	}

	values, err := redis.Values(conn.Do("HGETALL", keys[0]))
	if err != nil {
		return Author{}, errors.Wrap(err, "could not HGETALL")
	}
	str, err := redis.Strings(values, nil)
	if err != nil {
		return Author{}, errors.Wrap(err, "could not stringify")
	}
	var author Author
	err = unhashAuthorFromStrings(str, &author)
	if err != nil {
		return Author{}, errors.Wrap(err, "could not unhash author")
	}
	return author, nil
}

// AuthorBySlug fetches an author by slug
func (r *RedigoCache) AuthorBySlug(slug string) (Author, error) {

	conn := r.Pool.Get()
	defer conn.Close()

	slug = Slugify(slug, "-")
	pattern := fmt.Sprintf("author:%s:*", slug)
	keys, err := r.GetKeys(pattern)
	if err != nil {
		return Author{}, errors.Wrap(
			err,
			fmt.Sprintf(
				"could not GetKeys with pattern %v",
				pattern,
			),
		)
	}
	if len(keys) > 1 {
		return Author{}, errors.New("multiple results from ID")
	}

	values, err := redis.Values(conn.Do("HGETALL", keys[0]))
	if err != nil {
		return Author{}, errors.Wrap(err, "could not HGETALL")
	}
	str, err := redis.Strings(values, nil)
	if err != nil {
		return Author{}, errors.Wrap(err, "could not stringify")
	}
	var author Author
	err = unhashAuthorFromStrings(str, &author)
	if err != nil {
		return Author{}, errors.Wrap(err, "could not unhash author")
	}
	return author, nil
}

// InsertBook inserts book into the cache
func (r *RedigoCache) InsertBook(book Book) error {

	conn := r.Pool.Get()
	defer conn.Close()

	cacheID, err := serializeBookID(book)
	if err != nil {
		return errors.Wrap(err, "could not serialize book cache ID")
	}
	if _, err := conn.Do("HMSET", redis.Args{}.Add(cacheID).AddFlat(&book)...); err != nil {
		return errors.Wrap(err, "could not HMSet book")
	}

	return nil
}

// InsertAuthor inserts author into the cache
func (r *RedigoCache) InsertAuthor(author Author) error {

	conn := r.Pool.Get()
	defer conn.Close()

	cacheID, err := serializeAuthorID(author)
	if err != nil {
		return errors.Wrap(err, "could not serialize author cache ID")
	}
	if _, err := conn.Do("HMSET", redis.Args{}.Add(cacheID).AddFlat(&author)...); err != nil {
		return errors.Wrap(err, "could not HMSet author")
	}

	return nil
}

// InsertPage inserts book into the cache
func (r *RedigoCache) InsertPage(page Page) error {

	conn := r.Pool.Get()
	defer conn.Close()

	cacheID, err := serializePageID(page)
	if err != nil {
		return errors.Wrap(err, "could not serialize page cache ID")
	}
	if _, err := conn.Do("HMSET", redis.Args{}.Add(cacheID).AddFlat(&page)...); err != nil {
		return errors.Wrap(err, "could not HMSet page")
	}

	return nil
}

// SaveBookQuery saves a query onto the cache for easy retrieval
func (r *RedigoCache) SaveBookQuery(key string, books []Book) error {
	// b, err := json.Marshal(&books)
	// if err != nil {
	// 	return errors.Wrap(err, "could not marshal array of type Book")
	// }
	// err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	// if err != nil {
	// 	return errors.Wrap(err, "could not SET book query")
	// }
	return nil
}

// GetBookQuery retrieves a saved  query from the cache
func (r *RedigoCache) GetBookQuery(key string) ([]Book, error) {
	// str, err := r.Client.Get(key).Result()
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not GET author")
	// }

	// b := []byte(str)
	var books []Book
	// err = json.Unmarshal(b, &books)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not unmarshal books JSON")
	// }

	return books, nil
}

// SavePageQuery saves a query onto the cache for easy retrieval
func (r *RedigoCache) SavePageQuery(key string, pages []Page) error {
	// b, err := json.Marshal(&pages)
	// if err != nil {
	// 	return errors.Wrap(err, "could not marshal array of type Page")
	// }
	// err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	// if err != nil {
	// 	return errors.Wrap(err, "could not SET page query")
	// }
	return nil
}

// GetPageQuery retrieves a saved  query from the cache
func (r *RedigoCache) GetPageQuery(key string) ([]Page, error) {
	// 	str, err := r.Client.Get(key).Result()
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "could not GET pages")
	// 	}

	// 	b := []byte(str)
	var pages []Page
	// err = json.Unmarshal(b, &pages)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not unmarshal pages JSON")
	// }

	return pages, nil
}

// SaveAuthorQuery saves a query onto the cache for easy retrieval
func (r *RedigoCache) SaveAuthorQuery(key string, authors []Author) error {
	// b, err := json.Marshal(&authors)
	// if err != nil {
	// 	return errors.Wrap(err, "could not marshal array of type Author")
	// }
	// err = r.Client.Set(key, string(b), 1*time.Hour).Err()
	// if err != nil {
	// 	return errors.Wrap(err, "could not SET book query")
	// }
	return nil
}

// GetAuthorQuery retrieves a saved  query from the cache
func (r *RedigoCache) GetAuthorQuery(key string) ([]Author, error) {
	// str, err := r.Client.Get(key).Result()
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not GET Authors")
	// }

	// b := []byte(str)
	var authors []Author
	// err = json.Unmarshal(b, &authors)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "could not unmarshal authors JSON")
	// }

	return authors, nil
}

// func serializeBookID(book Book) (string, error) {
// 	if book.ID == uuid.Nil {
// 		return "", errors.New(fmt.Sprintf("invalid book ID: %v", book.ID))
// 	}
// 	return fmt.Sprintf("book:%s:%s", book.Slug, book.ID.String()), nil
// }

// func parseBookID(bookCacheID string) (string, uuid.UUID, error) {
// 	arr := strings.Split(bookCacheID, ":")
// 	slug, idString := arr[1], arr[2]
// 	uid, err := uuid.FromString(idString)
// 	if err != nil {
// 		return "", uuid.Nil, errors.Wrap(err, "could not parse uuid")
// 	}
// 	return slug, uid, nil
// }

// func serializeAuthorID(author Author) (string, error) {
// 	if author.ID == uuid.Nil {
// 		return "", errors.New(fmt.Sprintf("invalid author ID: %v", author.ID))
// 	}
// 	return fmt.Sprintf("author:%s:%s", author.Slug, author.ID.String()), nil
// }

// func parseAuthorID(authorCacheID string) (string, uuid.UUID, error) {
// 	arr := strings.Split(authorCacheID, ":")
// 	slug, idString := arr[1], arr[2]
// 	uid, err := uuid.FromString(idString)
// 	if err != nil {
// 		return "", uuid.Nil, errors.Wrap(err, "could not parse uuid")
// 	}
// 	return slug, uid, nil
// }

// func serializePageID(page Page) (string, error) {
// 	if page.ID == uuid.Nil {
// 		return "", errors.New(fmt.Sprintf("invalid page ID: %v", page.ID))
// 	}
// 	return fmt.Sprintf("page:%s", page.ID.String()), nil
// }

// func parsePageID(pageCacheID string) (uuid.UUID, error) {
// 	arr := strings.Split(pageCacheID, ":")
// 	idString := arr[1]
// 	uid, err := uuid.FromString(idString)
// 	if err != nil {
// 		return uuid.Nil, errors.Wrap(err, "could not parse uuid")
// 	}
// 	return uid, nil
// }

// func stringMapOfBook(book Book) (map[string]interface{}, error) {
// 	b := map[string]interface{}{
// 		"id":         book.ID.String(),
// 		"title":      book.Title,
// 		"slug":       book.Slug,
// 		"page_count": strconv.Itoa(book.PageCount),
// 	}
// 	if book.PublicationYear.Valid {
// 		b["publication_year"] = strconv.Itoa(int(book.PublicationYear.Int64))
// 	}
// 	if book.File.Valid {
// 		b["file"] = book.File.String
// 	}
// 	if book.Source.Valid {
// 		b["source"] = book.Source.String
// 	}
// 	if book.AuthorID.Valid {
// 		b["author_id"] = book.AuthorID.UUID.String()
// 	}
// 	return b, nil
// }

func unhashBookFromStrings(tuples []string, book *Book) error {
	for i := 0; i < len(tuples); i += 2 {
		key, value := tuples[i], tuples[i+1]
		switch key {
		case "title":
			book.Title = value
		case "slug":
			book.Slug = value
		case "page_count":
			inty, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrap(err, "could not convert page count to int")
			}
			book.PageCount = inty
		case "id":
			uid, err := uuid.FromString(value)
			if err != nil {
				return errors.Wrap(err, "could not convert id to UUID")
			}
			book.ID = uid
		case "publication_year":
			inty, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrap(err, "could not convert page count to int")
			}
			book.PublicationYear = NewNullInt64(int64(inty))
		case "file":
			book.File = NewNullString(value)
		case "source":
			book.Source = NewNullString(value)
		case "author_id":
			book.AuthorID = NewNullUUID(value)
		}
	}
	return nil
}

func unhashPageFromStrings(tuples []string, page *Page) error {
	for i := 0; i < len(tuples); i += 2 {
		key, value := tuples[i], tuples[i+1]
		switch key {
		case "id":
			uid, err := uuid.FromString(value)
			if err != nil {
				return errors.Wrap(err, "could not convert id to UUID")
			}
			page.ID = uid
		case "body":
			page.Body = value
		case "page_number":
			inty, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrap(err, "could not convert page count to int")
			}
			page.PageNumber = inty
		case "book_id":
			uid, err := uuid.FromString(value)
			if err != nil {
				return errors.Wrap(err, "could not parse UUID string")
			}
			page.BookID = &uid
		}
	}
	return nil
}

func unhashAuthorFromStrings(tuples []string, author *Author) error {
	for i := 0; i < len(tuples); i += 2 {
		key, value := tuples[i], tuples[i+1]
		switch key {
		case "name":
			author.Name = value
		case "slug":
			author.Slug = value
		case "id":
			uid, err := uuid.FromString(value)
			if err != nil {
				return errors.Wrap(err, "could not convert id to UUID")
			}
			author.ID = uid
		}
	}
	return nil
}

// func unvectorizeBook(fields []string, results []interface{}) (Book, error) {
// 	var book Book
// 	for i, r := range results {
// 		switch v := r.(type) {
// 		case string:
// 			switch fields[i] {
// 			case "title":
// 				book.Title = v
// 			case "slug":
// 				book.Slug = v
// 			case "page_count":
// 				inty, err := strconv.Atoi(v)
// 				if err != nil {
// 					return Book{}, errors.Wrap(err, "could not convert page count to int")
// 				}
// 				book.PageCount = inty
// 			case "id":
// 				uid, err := uuid.FromString(v)
// 				if err != nil {
// 					return Book{}, errors.Wrap(err, "could not convert id to UUID")
// 				}
// 				book.ID = uid
// 			case "publication_year":
// 				inty, err := strconv.Atoi(v)
// 				if err != nil {
// 					return Book{}, errors.Wrap(err, "could not convert page count to int")
// 				}
// 				book.PublicationYear = NewNullInt64(int64(inty))
// 			case "file":
// 				book.File = NewNullString(v)
// 			case "source":
// 				book.Source = NewNullString(v)
// 			case "author_id":
// 				book.AuthorID = NewNullUUID(v)
// 			}
// 		}
// 	}
// 	return book, nil
// }
