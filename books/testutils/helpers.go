package testutils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/djangulo/library/books"
	config "github.com/djangulo/library/config/books"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // noqa
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // unneded namespace
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"testing"
)

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

// GetAuthorFromGraphQLResponse noqa
func GetAuthorFromGraphQLResponse(t *testing.T, body io.Reader) books.Author {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.Author
}

// GetAllAuthorFromGraphQLResponse noqa
func GetAllAuthorFromGraphQLResponse(t *testing.T, body io.Reader) []books.Author {
	t.Helper()
	gqlRes := ParseGraphQLResponse(t, body)
	return gqlRes.Data.AllAuthor
}

// Other helpers

// NewTestSQLStore Creates and returns a test database. Intended for use with
// integration tests.
func NewTestSQLStore(config *config.Config, database string) (*books.SQLStore, func(string)) {
	db, err := sqlx.Open("postgres", config.Database["main"].ConnStr())
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	stmt := fmt.Sprintf("CREATE DATABASE %s;", config.Database[database].Name)
	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatalf("failed to create test database %v", err)
	}
	testConnStr := config.Database[database].ConnStr()

	var once sync.Once
	once.Do(func() {
		migrateConn, err := sql.Open("postgres", testConnStr)
		if err != nil {
			log.Fatalf("failed to connect to '%s' database %v", database, err)
		}
		defer migrateConn.Close()
		driver, err := postgres.WithInstance(migrateConn, &postgres.Config{})
		m, err := migrate.NewWithDatabaseInstance(
			// "file://"+config.Project.Dirs.Migrations,
			"file://../migrations",
			"postgres",
			driver,
		)
		if err != nil {
			panic(err)
		}
		m.Up()

	})

	testDB, err := sqlx.Open("postgres", testConnStr)
	if err != nil {
		log.Fatalf("failed to connect to '%s' database %v", database, err)
	}
	removeDatabase := func(database string) {
		err := testDB.Close()
		dbName := config.Database[database].Name
		if err != nil {
			log.Fatalf("error closing connection to database '%s': %v", dbName, err)
		}
		stmt := fmt.Sprintf(`DROP DATABASE %s;`, dbName)
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatalf("error dropping database %s: %v", dbName, err)
		}
		db.Close()
	}

	return &books.SQLStore{DB: testDB}, removeDatabase
}

// Dum dums

// DummyMiddlewares noqa
var DummyMiddlewares = []func(http.Handler) http.Handler{}
