package testutils

import (
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
)

// NewStubStore noqa
func NewStubStore(available bool) *StubStore {

	testBooks := TestBookData()
	testPages := TestPageData()
	testAuthors := TestAuthorData()

	return &StubStore{
		Available:   available,
		books:       testBooks,
		pages:       testPages,
		authors:     testAuthors,
		BookCalls:   map[string]int{},
		PageCalls:   map[string]int{},
		AuthorCalls: map[string]int{},
	}
}

// StubStore for testing
type StubStore struct {
	Available   bool
	books       []books.Book
	pages       []books.Page
	authors     []books.Author
	BookCalls   map[string]int
	PageCalls   map[string]int
	AuthorCalls map[string]int
}

func (s *StubStore) IsAvailable() error {
	if !s.Available {
		return books.ErrSQLStoreUnavailable
	}
	return nil
}

// Books noqa
func (s *StubStore) Books(limit, offset int) ([]books.Book, error) {
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
func (s *StubStore) BookByID(id uuid.UUID) (books.Book, error) {
	for _, b := range s.books {
		if id.String() == b.ID.String() {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Book{}, nil
}

// BookBySlug noqa
func (s *StubStore) BookBySlug(slug string) (books.Book, error) {
	for _, b := range s.books {
		if b.Slug == slug {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Book{}, nil
}

// BooksByAuthor noqa
func (s *StubStore) BooksByAuthor(name string) ([]books.Book, error) {
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
func (s *StubStore) Pages(limit, offset int) ([]books.Page, error) {
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
func (s *StubStore) PageByID(id uuid.UUID) (books.Page, error) {
	for _, p := range s.pages {
		if id.String() == p.ID.String() {
			s.PageCalls[p.ID.String()]++
			return p, nil
		}
	}
	return books.Page{}, nil
}

// PageByBookAndNumber noqa
func (s *StubStore) PageByBookAndNumber(bookID uuid.UUID, number int) (books.Page, error) {
	for _, p := range s.pages {
		if bookID.String() == p.BookID.String() && p.PageNumber == number {
			s.PageCalls[p.ID.String()]++
			return p, nil
		}
	}
	return books.Page{}, nil
}

// Authors noqa
func (s *StubStore) Authors(limit, offset int) ([]books.Author, error) {
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
func (s *StubStore) AuthorByID(id uuid.UUID) (books.Author, error) {
	for _, b := range s.authors {
		if id.String() == b.ID.String() {
			s.AuthorCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Author{}, nil
}

// AuthorBySlug noqa
func (s *StubStore) AuthorBySlug(slug string) (books.Author, error) {
	slug = books.Slugify(slug, "-")
	for _, a := range s.authors {
		if a.Slug == slug {
			s.AuthorCalls[a.ID.String()]++
			return a, nil
		}
	}
	return books.Author{}, nil
}

func (s *StubStore) InsertBook(book books.Book) error {
	s.books = append(s.books, book)
	return nil
}

func (s *StubStore) InsertPage(page books.Page) error {
	s.pages = append(s.pages, page)
	return nil
}

func (s *StubStore) InsertAuthor(author books.Author) error {
	s.authors = append(s.authors, author)
	return nil
}

func (s *StubStore) BulkInsertBooks(books []books.Book) error {
	s.books = append(s.books, books...)
	return nil
}

func (s *StubStore) BulkInsertPages(pages []books.Page) error {
	s.pages = append(s.pages, pages...)
	return nil
}

func (s *StubStore) BulkInsertAuthors(authors []books.Author) error {
	s.authors = append(s.authors, authors...)
	return nil
}
