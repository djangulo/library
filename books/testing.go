package books

import (
	"database/sql"
	"encoding/json"
	"fmt"
	// "github.com/djangulo/library/config"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

func NewStubStore() *StubStore {
	// cnf := config.Get()

	jsonBooks, _ := os.Open("/home/djangulo/go/src/github.com/djangulo/library/books/testdata/fakeBooks.json")
	jsonPages, _ := os.Open("/home/djangulo/go/src/github.com/djangulo/library/books/testdata/fakePages.json")

	defer jsonBooks.Close()
	defer jsonPages.Close()

	// bookspath := filepath.Join(cnf.Project.RootDir, "books", "testdata", "fakeBooks.json")
	// pagesPath := filepath.Join(cnf.Project.RootDir, "books", "testdata", "fakePages.json")

	datBooks, _ := ioutil.ReadAll(jsonBooks)
	datPages, _ := ioutil.ReadAll(jsonPages)
	var tmpBooks []map[string]interface{}
	var tmpPages []map[string]interface{}
	json.Unmarshal(datBooks, &tmpBooks)
	json.Unmarshal(datPages, &tmpPages)

	books := make([]Book, 0)
	for _, b := range tmpBooks {
		book := Book{
			ID:              uuid.Must(uuid.FromString(b["id"].(string))),
			Title:           b["title"].(string),
			Slug:            b["slug"].(string),
			Author:          sql.NullString{Valid: true, String: b["author"].(string)},
			PublicationYear: sql.NullInt64{Valid: true, Int64: int64(b["publication_year"].(float64))},
			PageCount:       int(b["page_count"].(float64)),
			Pages:           []Page{},
		}
		books = append(books, book)
	}

	pages := make([]Page, 0)
	for _, p := range tmpPages {
		page := Page{
			ID:         uuid.Must(uuid.FromString(p["id"].(string))),
			BookID:     uuid.Must(uuid.FromString(p["book_id"].(string))),
			PageNumber: int(p["page_number"].(float64)),
			Body:       p["body"].(string),
		}
		pages = append(pages, page)
	}

	return &StubStore{
		books:     books,
		pages:     pages,
		BookCalls: map[uuid.UUID]int{},
		PageCalls: map[uuid.UUID]int{},
	}
}

type StubStore struct {
	books     []Book
	pages     []Page
	BookCalls map[uuid.UUID]int
	PageCalls map[uuid.UUID]int
}

func (s *StubStore) Books(limit int) ([]Book, error) {
	return s.books, nil
}

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

func (s *StubStore) BookByID(ID uuid.UUID) (Book, error) {
	bid, _ := ID.Value()
	fmt.Printf("\n\n%+v\n\n", bid)
	for _, b := range s.books {
		id, _ := b.ID.Value()
		if id == bid {
			return b, nil
		}
	}
	return Book{}, nil
}

func (s *StubStore) BookBySlug(slug string) (Book, error) {
	for _, b := range s.books {
		if b.Slug == slug {
			return b, nil
		}
	}
	return Book{}, nil
}

func (s *StubStore) BooksByAuthor(author string) ([]Book, error) {
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

// Assertions

func AssertBooks(t *testing.T, got, want []Book) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func AssertPages(t *testing.T, got, want []Page) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func AssertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("response status is wrong, got %d want %d", got, want)
	}
}

// Other helpers

/*
ExtractMessagesFromResponse :
Extracts  Books from a JSON response
*/
func ExtractBooksFromResponse(
	t *testing.T,
	body io.Reader,
) (books []Book) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&books)
	if err != nil {
		t.Fatalf("Unable to parse response from server '%s' into []Book, '%v'", body, err)
	}
	return
}

/*
NewTestSQLStore :
Creates and returns a test database. Intended for use with integration tests.
*/
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

var DummyMiddlewares = []func(http.Handler) http.Handler{}
