package testutils

import (
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// NewStubCache noqa
func NewStubCache(available error) *StubCache {

	testBooks := TestBookData()
	testPages := TestPageData()
	testAuthors := TestAuthorData()

	return &StubCache{
		Available:     available,
		books:         testBooks,
		pages:         testPages,
		authors:       testAuthors,
		BookCalls:     map[string]int{},
		PageCalls:     map[string]int{},
		AuthorCalls:   map[string]int{},
		QueryCalls:    map[string]int{},
		BookQueries:   map[string]([]books.Book){},
		PageQueries:   map[string]([]books.Page){},
		AuthorQueries: map[string]([]books.Author){},
	}
}

// StubCache for testing
type StubCache struct {
	Available     error
	books         []books.Book
	pages         []books.Page
	authors       []books.Author
	QueryCalls    map[string]int
	BookCalls     map[string]int
	PageCalls     map[string]int
	AuthorCalls   map[string]int
	BookQueries   map[string][]books.Book
	PageQueries   map[string][]books.Page
	AuthorQueries map[string][]books.Author
}

func (s *StubCache) IsAvailable() error {
	return s.Available
}

// Books noqa
func (s *StubCache) Books(limit, offset int) ([]books.Book, error) {
	s.BookCalls["list"]++
	items := s.books
	length := len(items)
	if offset > length {
		return items[length:], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		return items[offset:], nil
	}
	if limit > length {
		limit = length
	}
	return items[(0 + offset):(offset + limit)], nil
}

// BookByID noqa
func (s *StubCache) BookByID(id uuid.UUID) (books.Book, error) {
	for _, b := range s.books {
		if id == b.ID {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Book{}, nil
}

// BookBySlug noqa
func (s *StubCache) BookBySlug(slug string) (books.Book, error) {
	for _, b := range s.books {
		if b.Slug == slug {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Book{}, nil
}

// BooksByAuthor noqa
func (s *StubCache) BooksByAuthor(name string) ([]books.Book, error) {
	s.BookCalls["list"]++
	var id *uuid.UUID
	for _, a := range s.authors {
		if a.Name == name {
			id = &a.ID
			break
		}
	}
	books := make([]books.Book, 0)
	for _, b := range s.books {
		if b.AuthorID.Valid {
			if b.AuthorID.UUID.String() == id.String() {
				books = append(books, b)
			}

		}
	}
	return books, nil
}

// Pages noqa
func (s *StubCache) Pages(limit, offset int) ([]books.Page, error) {
	s.PageCalls["list"]++
	items := s.pages
	length := len(items)
	if offset > length {
		return items[length:], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		return items[offset:], nil
	}
	if limit > length {
		limit = length
	}
	return items[(0 + offset):(offset + limit)], nil
}

// PageByID noqa
func (s *StubCache) PageByID(id uuid.UUID) (books.Page, error) {
	for _, p := range s.pages {
		if id.String() == p.ID.String() {
			s.PageCalls[p.ID.String()]++
			return p, nil
		}
	}
	return books.Page{}, nil
}

// PageByBookAndNumber noqa
func (s *StubCache) PageByBookAndNumber(bookID uuid.UUID, number int) (books.Page, error) {
	for _, p := range s.pages {
		if bookID.String() == p.BookID.String() && p.PageNumber == number {
			s.PageCalls[p.ID.String()]++
			return p, nil
		}
	}
	return books.Page{}, nil
}

// Authors noqa
func (s *StubCache) Authors(limit, offset int) ([]books.Author, error) {
	s.AuthorCalls["list"]++
	items := s.authors
	length := len(items)
	if offset > length {
		return items[length:], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		return items[offset:], nil
	}
	if limit > length {
		limit = length
	}
	return items[(0 + offset):(offset + limit)], nil
}

// AuthorByID noqa
func (s *StubCache) AuthorByID(id uuid.UUID) (books.Author, error) {
	for _, a := range s.authors {
		if id.String() == a.ID.String() {
			s.AuthorCalls[a.ID.String()]++
			return a, nil
		}
	}
	return books.Author{}, nil
}

// AuthorBySlug noqa
func (s *StubCache) AuthorBySlug(slug string) (books.Author, error) {
	slug = books.Slugify(slug, "-")
	for _, a := range s.authors {
		if a.Slug == slug {
			s.AuthorCalls[a.ID.String()]++
			return a, nil
		}
	}
	return books.Author{}, nil
}

// SaveBookQuery saves a query onto the cache for easy retrieval
func (s *StubCache) SaveBookQuery(key string, books []books.Book) error {
	s.QueryCalls[("SET:"+key)]++
	s.BookQueries[key] = books
	return nil
}

// GetBookQuery retrieves a saved  query from the cache
func (s *StubCache) GetBookQuery(key string) ([]books.Book, error) {
	s.QueryCalls[("GET:"+key)]++
	return s.BookQueries[key], nil
}

// SavePageQuery saves a query onto the cache for easy retrieval
func (s *StubCache) SavePageQuery(key string, pages []books.Page) error {
	s.QueryCalls[("SET:"+key)]++
	s.PageQueries[key] = pages
	return nil
}

// GetPageQuery retrieves a saved  query from the cache
func (s *StubCache) GetPageQuery(key string) ([]books.Page, error) {
	s.QueryCalls[("GET:"+key)]++
	return s.PageQueries[key], nil
}

// SaveAuthorQuery saves a query onto the cache for easy retrieval
func (s *StubCache) SaveAuthorQuery(key string, authors []books.Author) error {
	s.QueryCalls[("SET:"+key)]++
	s.AuthorQueries[key] = authors
	return nil
}

// GetAuthorQuery retrieves a saved  query from the cache
func (s *StubCache) GetAuthorQuery(key string) ([]books.Author, error) {
	s.QueryCalls[("GET:"+key)]++
	return s.AuthorQueries[key], nil
}

func (s *StubCache) InsertBook(book books.Book) error {
	if book.ID != uuid.Nil {
		s.books = append(s.books, book)
		return nil
	}
	return errors.New("Invalid book")
}

func (s *StubCache) InsertPage(page books.Page) error {
	if page.ID != uuid.Nil {
		s.pages = append(s.pages, page)
		return nil
	}
	return errors.New("Invalid page")
}

func (s *StubCache) InsertAuthor(author books.Author) error {
	if author.ID != uuid.Nil {
		s.authors = append(s.authors, author)
		return nil
	}
	return errors.New("Invalid author")
}
