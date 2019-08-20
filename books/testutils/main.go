package testutils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/djangulo/library/books"
	config "github.com/djangulo/library/config/books"
	"github.com/gofrs/uuid"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // unneded namespace
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // unneded namespace
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sync"
	"testing"
)

// TestBookData reads books json data and returns as a slice
func TestBookData() (books []books.Book) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakeBooks.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &books)
	return
}

// TestPageData reads pages json data and returns as a slice
func TestPageData() (pages []books.Page) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakePages.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &pages)
	return
}

// TestAuthorsData reads pages json data and returns as a slice
func TestAuthorsData() (authors []books.Author) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakeAuthors.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &authors)
	return
}

// NewStubStore noqa
func NewStubStore() *StubStore {

	books := TestBookData()
	pages := TestPageData()
	authors := TestAuthorsData()

	return &StubStore{
		books:       books,
		pages:       pages,
		authors:     authors,
		BookCalls:   map[string]int{},
		PageCalls:   map[string]int{},
		AuthorCalls: map[string]int{},
	}
}

// StubStore for testing
type StubStore struct {
	books       []books.Book
	pages       []books.Page
	authors     []books.Author
	BookCalls   map[string]int
	PageCalls   map[string]int
	AuthorCalls map[string]int
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
	for _, b := range s.authors {
		if b.Slug == slug {
			s.AuthorCalls[b.ID.String()]++
			return b, nil
		}
	}
	return books.Author{}, nil
}

// GraphQLResponse server response object
type GraphQLResponse struct {
	Data   GraphQLDataResponse      `json:"data"`
	Errors []map[string]interface{} `json:"errors"`
}

// GraphQLDataResponse noqa
type GraphQLDataResponse struct {
	Book      books.Book     `json:"book"`
	AllBook   []books.Book   `json:"allBook"`
	Page      books.Page     `json:"page"`
	AllPage   []books.Page   `json:"allPage"`
	Author    books.Author   `json:"author"`
	AllAuthor []books.Author `json:"allAuthor"`
}

// Utils

// FlattenJSON write your test JSON multiline, this will one-line it
func FlattenJSON(raw []byte) []byte {
	re := regexp.MustCompile(`[\n\t]+`)
	return re.ReplaceAll(raw, []byte(""))
}

// NewJSONPostRequest noqa
func NewJSONPostRequest(url string, rawJSON []byte) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(rawJSON))
	request.Header.Set("Content-Type", "application/json")
	return request
}

// ParseGraphQLResponse noqa
func ParseGraphQLResponse(t *testing.T, body io.Reader) (gqlResponse GraphQLResponse) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&gqlResponse)
	if err != nil {
		t.Fatalf("Unable to parse GraphQL response from server '%s': '%v'", body, err)
	}
	return
}

// GetBookFromGraphQLResponse noqa
func GetBookFromGraphQLResponse(t *testing.T, body io.Reader) books.Book {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.Book
}

// GetAllBookFromGraphQLResponse noqa
func GetAllBookFromGraphQLResponse(t *testing.T, body io.Reader) []books.Book {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.AllBook
}

// GetPageFromGraphQLResponse noqa
func GetPageFromGraphQLResponse(t *testing.T, body io.Reader) books.Page {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.Page
}

// GetAllPageFromGraphQLResponse noqa
func GetAllPageFromGraphQLResponse(t *testing.T, body io.Reader) []books.Page {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.AllPage
}

// Assertions

// AssertBooks noqa
func AssertBooks(t *testing.T, got, want []books.Book) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertPages noqa
func AssertPages(t *testing.T, got, want []books.Page) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertBookStoreCalls noqa
func AssertBookStoreCalls(t *testing.T, store *StubStore, id string, want int) {
	t.Helper()
	got := store.BookCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertPageStoreCalls noqa
func AssertPageStoreCalls(t *testing.T, store *StubStore, id string, want int) {
	t.Helper()
	got := store.PageCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertStatus noqa
func AssertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	got := response.Code
	if got != want {
		t.Errorf("response status is wrong, got %d want %d", got, want)
	}
}

func getVal(x interface{}) reflect.Value {
	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

// AssertEqual noqa
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	gotVal := getVal(got)
	wantVal := getVal(want)
	if gotVal.Kind() != wantVal.Kind() {
		t.Errorf("cannot compare type %T to %T", got, want)
	}
	switch gotVal.Kind() {
	case reflect.Struct, reflect.Array:
		if !reflect.DeepEqual(gotVal, wantVal) {
			t.Errorf("got %v want %v", gotVal, wantVal)
		}
	default:
		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	}
}

// AssertUUIDsEqual noqa
func AssertUUIDsEqual(t *testing.T, got, want uuid.UUID) {
	t.Helper()
	g, err := got.Value()
	if err != nil {
		t.Errorf("failed to read value from got type UUID %v: %v", got, err)
	}
	w, err := want.Value()
	if err != nil {
		t.Errorf("failed to read value from got type UUID %v: %v", got, err)
	}
	if g != w {
		t.Errorf("got %v want %v", g, w)
	}
}

// Other helpers

// NewTestSQLStore Creates and returns a test database. Intended for use with
// integration tests.
func NewTestSQLStore(config config.Config) (*books.SQLStore, func()) {
	db, err := sqlx.Open("postgres", config.Database["main"].ConnStr())
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	stmt := fmt.Sprintf("CREATE DATABASE %s;", config.Database["test"].Name)
	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatalf("failed to create test database %v", err)
	}
	testConnStr := config.Database["test"].ConnStr()

	var once sync.Once
	once.Do(func() {
		migrateConn, err := sql.Open("postgres", testConnStr)
		if err != nil {
			log.Fatalf("failed to connect to test database %v", err)
		}
		defer migrateConn.Close()

		driver, err := postgres.WithInstance(migrateConn, &postgres.Config{})
		m, err := migrate.NewWithDatabaseInstance(
			"file://"+config.Project.Dirs.Migrations,
			"postgres",
			driver,
		)
		m.Up()

	})

	testDB, err := sqlx.Open("postgres", testConnStr)
	if err != nil {
		log.Fatalf("failed to connect to test database %v", err)
	}
	removeDatabase := func() {
		testDB.Close()
		db.Exec(`DROP DATABASE library_test_database;`)
		db.Close()
	}

	return &books.SQLStore{DB: testDB}, removeDatabase
}

// Dum dums

// DummyMiddlewares noqa
var DummyMiddlewares = []func(http.Handler) http.Handler{}
