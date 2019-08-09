package books

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"testing"
)

func NewStubStore() *StubStore {

	bookspath := filepath.Join(config.RootDir, "books", "testdata", "fakeBooks.json")
	pagesPath := filepath.Join(config.RootDir, "books", "testdata", "fakePages.json")

	datBooks, _ := ioutil.ReadFile(bookspath)
	datPages, _ := ioutil.ReadFile(pagesPath)
	var books []Book
	var pages []Page
	json.Unmarshal(datBooks, &boooks)
	json.Unmarshal(datPages, &pages)
	store := books.NewStubStore(boooks, pages, map[uuid.UUID]int{}, map[uuid.UUID]int{})

	return &StubStore{
		books:     initialBooks,
		pages:     initialPages,
		BookCalls: initialBookCalls,
		PageCalls: initialPageCalls,
	}
}

type StubStore struct {
	books     []Book
	pages     []Page
	BookCalls map[uuid.UUID]int
	PageCalls map[uuid.UUID]int
}

func (s *StubStore) Books() []Book {
	return s.books
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

func (s *StubStore) BooksByID(ID uuid.UUID) (Book, error) {
	for _, b := range s.books {
		if b.ID == ID {
			return b
		}
	}
	return nil, nil
}

func (s *StubStore) BooksBySlug(slug string) (Book, error) {
	for _, b := range s.books {
		if b.Slug == slug {
			return b
		}
	}
	return nil, nil
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
