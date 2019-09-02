package testutils

import (
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// NewStubCache noqa
func NewStubCache(available error, prepopulate bool) *StubCache {

	if available != nil {
		return &StubCache{Available: books.ErrCacheUnavailable}
	}

	testBooks := make([]*books.Book, 0)
	testAuthors := make([]*books.Author, 0)
	testPages := make([]*books.Page, 0)
	if prepopulate {
		testBooks = TestBookData()
		testPages = TestPageData()
		testAuthors = TestAuthorData()
	}

	return &StubCache{
		Available:     available,
		books:         testBooks,
		pages:         testPages,
		authors:       testAuthors,
		BookCalls:     map[string]int{},
		PageCalls:     map[string]int{},
		AuthorCalls:   map[string]int{},
		QueryCalls:    map[string]int{},
		BookQueries:   map[string]([]*books.Book){},
		PageQueries:   map[string]([]*books.Page){},
		AuthorQueries: map[string]([]*books.Author){},
	}
}

// StubCache for testing
type StubCache struct {
	Available     error
	books         []*books.Book
	pages         []*books.Page
	authors       []*books.Author
	QueryCalls    map[string]int
	BookCalls     map[string]int
	PageCalls     map[string]int
	AuthorCalls   map[string]int
	BookQueries   map[string][]*books.Book
	PageQueries   map[string][]*books.Page
	AuthorQueries map[string][]*books.Author
}

// IsAvailable noqa
func (s *StubCache) IsAvailable() error {
	return s.Available
}

// BookByID noqa
func (s *StubCache) BookByID(
	book *books.Book,
	ID *uuid.UUID,
	fields []string,
) error {
	for _, b := range s.books {
		if *ID == b.ID {
			s.BookCalls[b.ID.String()]++
			book = b
			return nil
		}
	}
	return books.ErrNotFoundInCache
}

// BookBySlug noqa
func (s *StubCache) BookBySlug(
	book *books.Book,
	slug string,
	fields []string,
) error {
	for _, b := range s.books {
		if slug == b.Slug {
			s.BookCalls[b.ID.String()]++
			book = b
			return nil
		}
	}
	return books.ErrNotFoundInCache
}

// PageByID noqa
func (s *StubCache) PageByID(
	page *books.Page,
	ID *uuid.UUID,
	fields []string,
) error {
	for _, p := range s.pages {
		if *ID == p.ID {
			s.PageCalls[p.ID.String()]++
			page = p
			return nil
		}
	}
	return books.ErrNotFoundInCache
}

// PageByBookAndNumber noqa
func (s *StubCache) PageByBookAndNumber(
	page *books.Page,
	bookID *uuid.UUID,
	number int,
	fields []string,
) error {
	for _, p := range s.pages {
		if bookID == p.BookID && number == p.PageNumber {
			s.PageCalls[p.ID.String()]++
			page = p
			return nil
		}
	}
	return books.ErrNotFoundInCache
}

// AuthorByID noqa
func (s *StubCache) AuthorByID(
	author *books.Author,
	ID *uuid.UUID,
	fields []string,
) error {
	for _, a := range s.authors {
		if *ID == a.ID {
			s.AuthorCalls[a.ID.String()]++
			author = a
			return nil
		}
	}
	return books.ErrNotFoundInCache
}

// AuthorBySlug noqa
func (s *StubCache) AuthorBySlug(
	author *books.Author,
	slug string,
	fields []string,
) error {
	for _, a := range s.authors {
		if slug == a.Slug {
			s.AuthorCalls[a.ID.String()]++
			author = a
			return nil
		}
	}
	return books.ErrNotFoundInCache
}

// SaveBookQuery saves a query onto the cache for easy retrieval
func (s *StubCache) SaveBookQuery(key string, books []*books.Book) error {
	s.QueryCalls[("SET:"+key)]++
	s.BookQueries[key] = books
	return nil
}

// BookQuery retrieves a saved  query from the cache
func (s *StubCache) BookQuery(books *[]*books.Book, key string) error {
	s.QueryCalls[("GET:"+key)]++
	result := s.BookQueries[key]
	books = &result
	return nil
}

// SavePageQuery saves a query onto the cache for easy retrieval
func (s *StubCache) SavePageQuery(key string, pages []*books.Page) error {
	s.QueryCalls[("SET:"+key)]++
	s.PageQueries[key] = pages
	return nil
}

// PageQuery retrieves a saved  query from the cache
func (s *StubCache) PageQuery(pages *[]*books.Page, key string) error {
	s.QueryCalls[("GET:"+key)]++
	result := s.PageQueries[key]
	pages = &result
	return nil
}

// SaveAuthorQuery saves a query onto the cache for easy retrieval
func (s *StubCache) SaveAuthorQuery(key string, authors []*books.Author) error {
	s.QueryCalls[("SET:"+key)]++
	s.AuthorQueries[key] = authors
	return nil
}

// AuthorQuery retrieves a saved  query from the cache
func (s *StubCache) AuthorQuery(authors *[]*books.Author, key string) error {
	s.QueryCalls[("GET:"+key)]++
	result := s.AuthorQueries[key]
	authors = &result
	return nil
}

// InsertBook noqa
func (s *StubCache) InsertBook(book *books.Book) error {
	if book.ID != uuid.Nil {
		s.books = append(s.books, book)
		return nil
	}
	return errors.New("Invalid book")
}

// InsertPage noqa
func (s *StubCache) InsertPage(page *books.Page) error {
	if page.ID != uuid.Nil {
		s.pages = append(s.pages, page)
		return nil
	}
	return errors.New("Invalid page")
}

// InsertAuthor noqa
func (s *StubCache) InsertAuthor(author *books.Author) error {
	if author.ID != uuid.Nil {
		s.authors = append(s.authors, author)
		return nil
	}
	return errors.New("Invalid author")
}
