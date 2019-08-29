package testutils

import (
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
	"sort"
	"time"
)

const (
	// Available to use in tests for legibility
	Available = true
	// Unavailable to use in tests for legibility
	Unavailable = false
	// PrepopulateStore to use in tests for legibility
	PrepopulateStore = true
	// NOPrepopulateStore to use in tests for legibility
	NOPrepopulateStore = false
)

// NewStubStore noqa
func NewStubStore(available bool, prepopulate bool) *StubStore {

	var tAuthors []books.Author
	var tBooks []books.Book
	var tPages []books.Page
	if prepopulate {
		tAuthors = TestAuthorData()
		tBooks = TestBookData()
		tPages = TestPageData()
	}

	return &StubStore{
		Available:         available,
		authors:           tAuthors,
		books:             tBooks,
		pages:             tPages,
		BookCalls:         map[string]int{},
		PageCalls:         map[string]int{},
		AuthorCalls:       map[string]int{},
		InsertBookCalls:   map[string]int{},
		InsertPageCalls:   map[string]int{},
		InsertAuthorCalls: map[string]int{},
	}
}

// StubStore for testing
type StubStore struct {
	Available         bool
	books             []books.Book
	pages             []books.Page
	authors           []books.Author
	BookCalls         map[string]int
	PageCalls         map[string]int
	AuthorCalls       map[string]int
	InsertBookCalls   map[string]int
	InsertPageCalls   map[string]int
	InsertAuthorCalls map[string]int
}

// IsAvailable noqa
func (s *StubStore) IsAvailable() error {
	if !s.Available {
		return books.ErrSQLStoreUnavailable
	}
	return nil
}

// Books noqa
func (s *StubStore) Books(
	limit int,
	offset int,
	lastID uuid.UUID,
	lastCreated time.Time,
	fields []string,
) ([]books.Book, error) {
	if len(s.books) == 0 {
		return nil, books.ErrNoResults
	}
	sort.SliceStable(s.books, func(i, j int) bool {
		return s.books[i].CreatedAt.After(s.books[j].CreatedAt)
	})
	sort.SliceStable(s.books, func(i, j int) bool {
		return s.books[j].ID.String() < s.books[i].ID.String()
	})
	s.BookCalls["list"]++
	if lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range s.books {
			if b.CreatedAt == lastCreated && b.ID == lastID {
				if itemsLeft := len(s.books[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				return s.books[(i + 1):(i + 1 + limit)], nil
			}
		}
	}
	length := len(s.books)
	if offset > length {
		return s.books[length:], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		return s.books[offset:], nil
	}
	if limit > length {
		limit = length
	}
	return s.books[(0 + offset):(offset + limit)], nil
}

// BookByID noqa
func (s *StubStore) BookByID(id uuid.UUID, fields []string) (books.Book, error) {
	for _, b := range s.books {
		if id == b.ID {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Book{}, nil
}

// BookBySlug noqa
func (s *StubStore) BookBySlug(slug string, fields []string) (books.Book, error) {
	for _, b := range s.books {
		if b.Slug == slug {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Book{}, nil
}

// BooksByAuthor noqa
func (s *StubStore) BooksByAuthor(
	name string,
	limit int,
	offset int,
	lastID uuid.UUID,
	lastCreated time.Time,
	fields []string,
) ([]books.Book, error) {
	if len(s.books) == 0 {
		return nil, books.ErrNoResults
	}
	sort.SliceStable(s.books, func(i, j int) bool {
		return s.books[i].CreatedAt.After(s.books[j].CreatedAt)
	})
	sort.SliceStable(s.books, func(i, j int) bool {
		return s.books[j].ID.String() < s.books[i].ID.String()
	})
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
			if b.AuthorID.UUID == *id {
				books = append(books, b)
			}
		}
	}
	if lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range books {
			if b.CreatedAt == lastCreated && b.ID == lastID {
				if itemsLeft := len(books[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				return books[(i + 1):(i + 1 + limit)], nil
			}
		}
	}
	length := len(books)
	if offset > length {
		return books[length:], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		return books[offset:], nil
	}
	if limit > length {
		limit = length
	}
	return books[(0 + offset):(offset + limit)], nil
}

// Pages noqa
func (s *StubStore) Pages(
	limit int,
	offset int,
	lastID uuid.UUID,
	lastCreated time.Time,
	fields []string,
) ([]books.Page, error) {
	if len(s.pages) == 0 {
		return nil, books.ErrNoResults
	}
	sort.SliceStable(s.pages, func(i, j int) bool {
		return s.pages[i].CreatedAt.After(s.pages[j].CreatedAt)
	})
	sort.SliceStable(s.pages, func(i, j int) bool {
		return s.pages[j].ID.String() < s.pages[i].ID.String()
	})
	s.PageCalls["list"]++
	if lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range s.pages {
			if b.CreatedAt == lastCreated && b.ID == lastID {
				if itemsLeft := len(s.pages[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				return s.pages[(i + 1):(i + 1 + limit)], nil
			}
		}
	}
	length := len(s.pages)
	if offset > length {
		return s.pages[length:], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		return s.pages[offset:], nil
	}
	if limit > length {
		limit = length
	}
	return s.pages[(0 + offset):(offset + limit)], nil
}

// PageByID noqa
func (s *StubStore) PageByID(id uuid.UUID, fields []string) (books.Page, error) {
	for _, p := range s.pages {
		if id == p.ID {
			s.PageCalls[p.ID.String()]++
			return p, nil
		}
	}
	return books.Page{}, nil
}

// PageByBookAndNumber noqa
func (s *StubStore) PageByBookAndNumber(bookID uuid.UUID, number int, fields []string) (books.Page, error) {
	for _, p := range s.pages {
		if bookID == *p.BookID && p.PageNumber == number {
			s.PageCalls[p.ID.String()]++
			return p, nil
		}
	}
	return books.Page{}, nil
}

// Authors noqa
func (s *StubStore) Authors(
	limit int,
	offset int,
	lastID uuid.UUID,
	lastCreated time.Time,
	fields []string,
) ([]books.Author, error) {
	if len(s.authors) == 0 {
		return nil, books.ErrNoResults
	}
	sort.SliceStable(s.authors, func(i, j int) bool {
		return s.pages[i].CreatedAt.After(s.authors[j].CreatedAt)
	})
	sort.SliceStable(s.authors, func(i, j int) bool {
		return s.pages[j].ID.String() < s.authors[i].ID.String()
	})
	s.PageCalls["list"]++
	if lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range s.authors {
			if b.CreatedAt == lastCreated && b.ID == lastID {
				if itemsLeft := len(s.authors[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				return s.authors[(i + 1):(i + 1 + limit)], nil
			}
		}
	}
	length := len(s.authors)
	if offset > length {
		return s.authors[length:], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		return s.authors[offset:], nil
	}
	if limit > length {
		limit = length
	}
	return s.authors[(0 + offset):(offset + limit)], nil
}

// AuthorByID noqa
func (s *StubStore) AuthorByID(id uuid.UUID, fields []string) (books.Author, error) {
	for _, b := range s.authors {
		if id == b.ID {
			s.AuthorCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Author{}, nil
}

// AuthorBySlug noqa
func (s *StubStore) AuthorBySlug(slug string, fields []string) (books.Author, error) {
	slug = books.Slugify(slug, "-")
	for _, a := range s.authors {
		if a.Slug == slug {
			s.AuthorCalls[a.ID.String()]++
			return a, nil
		}
	}
	return books.Author{}, nil
}

// InsertBook noqa
func (s *StubStore) InsertBook(book books.Book) error {
	s.InsertBookCalls[book.ID.String()]++
	s.books = append(s.books, book)
	return nil
}

// InsertPage noqa
func (s *StubStore) InsertPage(page books.Page) error {
	s.InsertPageCalls[page.ID.String()]++
	s.pages = append(s.pages, page)
	return nil
}

// InsertAuthor noqa
func (s *StubStore) InsertAuthor(author books.Author) error {
	s.InsertAuthorCalls[author.ID.String()]++
	s.authors = append(s.authors, author)
	return nil
}

// BulkInsertBooks noqa
func (s *StubStore) BulkInsertBooks(books []books.Book) error {
	for range books {
		s.InsertBookCalls["bulk"]++
	}
	s.books = append(s.books, books...)
	return nil
}

// BulkInsertPages noqa
func (s *StubStore) BulkInsertPages(pages []books.Page) error {
	for range pages {
		s.InsertPageCalls["bulk"]++
	}
	s.pages = append(s.pages, pages...)
	return nil
}

// BulkInsertAuthors noqa
func (s *StubStore) BulkInsertAuthors(authors []books.Author) error {
	for range authors {
		s.InsertAuthorCalls["bulk"]++
	}
	s.authors = append(s.authors, authors...)
	return nil
}
