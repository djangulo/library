package books

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/djangulo/library/config"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// TestBookData reads books json data and returns as a slice
func TestBookData() (books []Book) {
	cnf = config.Get()
	path := filepath.Join(
		cnf.Project.RootDir,
		"books",
		"testdata",
		"fakeBooks.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &books)
	return
}

// TestPageData reads pages json data and returns as a slice
func TestPageData() (pages []Page) {
	cnf = config.Get()
	path := filepath.Join(
		cnf.Project.RootDir,
		"books",
		"testdata",
		"fakePages.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &pages)
	return
}

// NewStubStore noqa
func NewStubStore() *StubStore {

	books := TestBookData()
	pages := TestPageData()

	return &StubStore{
		books:     books,
		pages:     pages,
		BookCalls: map[string]int{},
		PageCalls: map[string]int{},
	}
}

// StubStore for testing
type StubStore struct {
	books     []Book
	pages     []Page
	BookCalls map[string]int
	PageCalls map[string]int
}

// Books noqa
func (s *StubStore) Books(limit, offset int) ([]Book, error) {
	s.BookCalls["list"]++
	if offset > len(s.books) {
		return s.books[len(s.books):], nil
	} else if offset < 0 {
		offset = 0
	}
	if limit+offset > len(s.books) {
		return s.books[offset:], nil
	}
	if limit > len(s.books) {
		limit = len(s.books)
	}
	return s.books[(0 + offset):(offset + limit)], nil
}

// Page noqa
func (s *StubStore) Page(bookID uuid.UUID, number int) Page {
	var page Page
	for _, p := range s.pages {
		if p.BookID == bookID && p.PageNumber == number {
			page = p
			break
		}
	}
	return page
}

// BookByID noqa
func (s *StubStore) BookByID(ID uuid.UUID) (Book, error) {
	bid, _ := ID.Value()
	for _, b := range s.books {
		id, _ := b.ID.Value()
		if id == bid {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return Book{}, nil
}

// BookBySlug noqa
func (s *StubStore) BookBySlug(slug string) (Book, error) {
	for _, b := range s.books {
		if b.Slug == slug {
			s.BookCalls[b.ID.String()]++
			return b, nil
		}
	}
	return Book{}, nil
}

// BooksByAuthor noqa
func (s *StubStore) BooksByAuthor(author string) ([]Book, error) {
	s.BookCalls["list"]++
	books := make([]Book, 0)
	for _, b := range s.books {
		if b.Author.Valid && strings.ToLower(b.Author.String) == strings.ToLower(author) {
			books = append(books, b)
		}
	}
	return books, nil
}

// func (s *StubStore) RoomNMessages(roomID, n, m int) []Message {
// 	messages := s.messages[roomID]
// 	return messages
// }

// func (s *StubStore) RoomLatestNMessages(roomID, n int) []Message {
// 	messages := s.messages[roomID]
// 	return messages[len(messages)-n:]
// }

// GraphQLResponse server response object
type GraphQLResponse struct {
	Data   GraphQLDataResponse      `json:"data"`
	Errors []map[string]interface{} `json:"errors"`
}

// GraphQLDataResponse noqa
type GraphQLDataResponse struct {
	Book      Book     `json:"book"`
	AllBook   []Book   `json:"allBook"`
	Page      Page     `json:"Page"`
	AllPage   []Page   `json:"allPage"`
	Author    Author   `json:"author"`
	AllAuthor []Author `json:"allAuthor"`
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
func GetBookFromGraphQLResponse(t *testing.T, body io.Reader) Book {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.Book
}

// GetAllBookFromGraphQLResponse noqa
func GetAllBookFromGraphQLResponse(t *testing.T, body io.Reader) []Book {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.AllBook
}

// Assertions

// AssertBooks noqa
func AssertBooks(t *testing.T, got, want []Book) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertPages noqa
func AssertPages(t *testing.T, got, want []Page) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertBookStoreCalls noqa
func AssertBookStoreCalls(t *testing.T, got, want int) {
	t.Helper()
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

// NewTestSQLStore Creates and returns a test database. Intended for use with integration tests.
func NewTestSQLStore(
	host,
	port,
	user,
	dbname,
	pass string,
) (*SQLStore, func()) {
	mainConnStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		user,
		pass,
		host,
		port,
		dbname,
	)
	mainDB, err := sqlx.Open("postgres", mainConnStr)
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	_, err = mainDB.Exec(`CREATE DATABASE test_database;`)
	if err != nil {
		log.Fatalf("failed to create test database %v", err)
	}

	testConnStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		user,
		pass,
		host,
		port,
		"library_test_database",
	)
	testDB, errOpenTest := sqlx.Open("postgres", testConnStr)
	if errOpenTest != nil {
		log.Fatalf("failed to connect to test database %v", errOpenTest)
	}

	// _, errCreateTable := testDB.Exec(CreateTables)
	// if errCreateTable != nil {
	// 	log.Fatalf("failed to create test DB table %v", errCreateTable)
	// }

	removeDatabase := func() {
		testDB.Close()
		mainDB.Exec(`DROP DATABASE library_test_database;`)
		mainDB.Close()
	}

	return &SQLStore{testDB}, removeDatabase
}

// Dum dums

// DummyMiddlewares noqa
var DummyMiddlewares = []func(http.Handler) http.Handler{}
