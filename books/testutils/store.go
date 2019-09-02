package testutils

import (
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
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

	var tAuthors []*books.Author
	var tBooks []*books.Book
	var tPages []*books.Page
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
	books             []*books.Book
	pages             []*books.Page
	authors           []*books.Author
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
	booksArr []*books.Book,
	limit int,
	offset int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {
	if len(s.books) == 0 {
		return books.ErrNoResults
	}
	s.BookCalls["list"]++
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range s.books {
			if b.CreatedAt == *lastCreated && b.ID == *lastID {
				if itemsLeft := len(s.books[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				booksArr = s.books[(i + 1):(i + 1 + limit)]
				return nil
			}
		}
	}
	length := len(s.books)
	if offset > length {
		booksArr = s.books[length:]
		return nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		booksArr = s.books[offset:]
		return nil
	}
	if limit > length {
		limit = length
	}
	booksArr = s.books[(0 + offset):(offset + limit)]
	return nil
}

// BookByID noqa
func (s *StubStore) BookByID(
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
	return nil
}

// BookBySlug noqa
func (s *StubStore) BookBySlug(
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
	return nil
}

// BooksByAuthor noqa
func (s *StubStore) BooksByAuthor(
	booksArr []*books.Book,
	name string,
	limit int,
	offset int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {
	if len(s.books) == 0 {
		return books.ErrNoResults
	}
	s.BookCalls["by author"]++
	var id *uuid.UUID
	for _, a := range s.authors {
		if a.Name == name {
			id = &a.ID
			break
		}
	}
	for _, b := range s.books {
		if b.AuthorID.Valid {
			if b.AuthorID.UUID == *id {
				booksArr = append(booksArr, b)
			}
		}
	}
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range s.books {
			if b.CreatedAt == *lastCreated && b.ID == *lastID {
				if itemsLeft := len(booksArr[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				booksArr = booksArr[(i + 1):(i + 1 + limit)]
				return nil
			}
		}
	}
	length := len(booksArr)
	if offset > length {
		booksArr = booksArr[length:]
		return nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		booksArr = booksArr[offset:]
		return nil
	}
	if limit > length {
		limit = length
	}
	booksArr = booksArr[(0 + offset):(offset + limit)]
	return nil
}

// Pages noqa
func (s *StubStore) Pages(
	pages []*books.Page,
	limit int,
	offset int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {
	if len(s.pages) == 0 {
		return books.ErrNoResults
	}
	s.PageCalls["list"]++
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range s.pages {
			if b.CreatedAt == *lastCreated && b.ID == *lastID {
				if itemsLeft := len(s.pages[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				pages = s.pages[(i + 1):(i + 1 + limit)]
				return nil
			}
		}
	}
	length := len(s.pages)
	if offset > length {
		pages = s.pages[length:]
		return nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		pages = s.pages[offset:]
		return nil
	}
	if limit > length {
		limit = length
	}
	pages = s.pages[(0 + offset):(offset + limit)]
	return nil
}

// PageByID noqa
func (s *StubStore) PageByID(
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
	return nil
}

// PageByBookAndNumber noqa
func (s *StubStore) PageByBookAndNumber(
	page *books.Page,
	bookID *uuid.UUID,
	number int,
	fields []string,
) error {
	for _, p := range s.pages {
		if *bookID == p.ID && number == p.PageNumber {
			s.PageCalls[p.ID.String()]++
			page = p
			return nil
		}
	}
	return nil
}

// Authors noqa
func (s *StubStore) Authors(
	authors []*books.Author,
	limit int,
	offset int,
	lastID *uuid.UUID,
	lastCreated *time.Time,
	fields []string,
) error {
	if len(s.authors) == 0 {
		return books.ErrNoResults
	}
	s.AuthorCalls["list"]++
	if *lastID != uuid.Nil && !lastCreated.IsZero() {
		for i, b := range s.authors {
			if b.CreatedAt == *lastCreated && b.ID == *lastID {
				if itemsLeft := len(s.authors[(i + 1):]); limit > itemsLeft {
					limit = itemsLeft - 1
				}
				authors = s.authors[(i + 1):(i + 1 + limit)]
				return nil
			}
		}
	}
	length := len(s.authors)
	if offset > length {
		authors = s.authors[length:]
		return nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > length {
		authors = s.authors[offset:]
		return nil
	}
	if limit > length {
		limit = length
	}
	authors = s.authors[(0 + offset):(offset + limit)]
	return nil
}

// AuthorByID noqa
func (s *StubStore) AuthorByID(
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
	return nil
}

// AuthorBySlug noqa
func (s *StubStore) AuthorBySlug(
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
	return nil
}

// InsertBook noqa
func (s *StubStore) InsertBook(book *books.Book) error {
	if book == nil {
		return books.ErrNilPointerPassed
	}
	s.InsertBookCalls[book.ID.String()]++
	s.books = append(s.books, book)
	return nil
}

// InsertPage noqa
func (s *StubStore) InsertPage(page *books.Page) error {
	if page == nil {
		return books.ErrNilPointerPassed
	}
	s.InsertPageCalls[page.ID.String()]++
	s.pages = append(s.pages, page)
	return nil
}

// InsertAuthor noqa
func (s *StubStore) InsertAuthor(author *books.Author) error {
	if author == nil {
		return books.ErrNilPointerPassed
	}
	s.InsertAuthorCalls[author.ID.String()]++
	s.authors = append(s.authors, author)
	return nil
}

// BulkInsertBooks noqa
func (s *StubStore) BulkInsertBooks(books []*books.Book) error {
	for range books {
		s.InsertBookCalls["bulk"]++
	}
	s.books = append(s.books, books...)
	return nil
}

// BulkInsertPages noqa
func (s *StubStore) BulkInsertPages(pages []*books.Page) error {
	for range pages {
		s.InsertPageCalls["bulk"]++
	}
	s.pages = append(s.pages, pages...)
	return nil
}

// BulkInsertAuthors noqa
func (s *StubStore) BulkInsertAuthors(authors []*books.Author) error {
	for range authors {
		s.InsertAuthorCalls["bulk"]++
	}
	s.authors = append(s.authors, authors...)
	return nil
}
