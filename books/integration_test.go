package books_test

import (
	"fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/books/testutils"
	config "github.com/djangulo/library/config/books"
	"github.com/go-chi/chi/middleware"
	// // "github.com/djangulo/library/books"
	// // config "github.com/djangulo/library/config/books"
	// "github.com/gofrs/uuid"
	// // "github.com/jmoiron/sqlx"
	"net/http"
	"net/http/httptest"
	// "os"
	"log"
	"testing"
)

// func TestSQLStore(t *testing.T) {
// 	fmt.Println("smoke!")
// 	booksFixture, err := os.Open("testdata/fakeBooks.json")
// 	if err != nil {
// 		panic(err)
// 	}
// 	pagesFixture, err := os.Open("testdata/fakePages.json")
// 	if err != nil {
// 		panic(err)
// 	}
// 	// fmt.Println(booksFixture)
// 	// fmt.Println(pagesFixture)
// }

var (
	cnf = config.Get()
)

var middlewares = []func(http.Handler) http.Handler{
	middleware.RequestID,
	middleware.RealIP,
	// middleware.Logger,
	middleware.Recoverer,
}

func TestBookQueriesWithoutCache(t *testing.T) {

	store, remove := testutils.NewTestSQLStore(cnf)
	defer remove()
	books.AcquireGutenberg(cnf)
	err := books.SaveJSON(cnf)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "test")
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}
	t.Run("tests without cache", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, false)
			jsonStream := []byte(`
			{
				"query": "{
					allBook {
						title,
						publication_year,
						slug,
						author,
						id,
						pages{
							id
						}
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := len(testBooks)

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, false)
			jsonStream := []byte(`{
				"query": "{
					allBook(limit: 3) {
						title,
						pages{id}
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := 3

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, false)
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 1) {
						title,
						pages{id}
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			if len(bookArr) == 0 {
				t.Fatal("expected a result but got none")
			}
			got := bookArr[0].Title
			want := testBooks[1].Title

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query allBook filtered by author", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, false)
			str := fmt.Sprintf(`{
				"query": "{
					allBook(author: \"%s\") {
						title,
						pages{id}
					}
				}"
			}`, testAuthors[0].Name)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := 3

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query a book by id", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, false)
			str := fmt.Sprintf(`
			{
				"query": "{
					book(id:\"%s\") {
						title,
						publication_year,
						slug,
						author,
						id
					}
				}"
			}
			`, testBooks[0].ID.String())
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := testBooks[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query a book by slug", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, false)
			str := fmt.Sprintf(`
			{
				"query": "{
					book(slug:\"%s\") {
						title,
						publication_year,
						slug,
						author,
						id
					}
				}"
			}
			`, testBooks[0].Slug)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := testBooks[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
	})

	t.Run("tests with a cache", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, true)
			jsonStream := []byte(`
			{
				"query": "{
					allBook {
						title,
						publication_year,
						slug,
						author,
						id,
						pages{
							id
						}
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := len(testBooks)

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, true)
			jsonStream := []byte(`{
				"query": "{
					allBook(limit: 3) {
						title,
						pages{id}
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := 3

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, true)
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 1) {
						title,
						pages{id}
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			if len(bookArr) == 0 {
				t.Fatal("expected a result but got none")
			}
			got := bookArr[0].Title
			want := testBooks[1].Title

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query allBook filtered by author", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, true)
			str := fmt.Sprintf(`{
				"query": "{
					allBook(author: \"%s\") {
						title,
						pages{id}
					}
				}"
			}`, testAuthors[0].Name)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := 3

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query a book by id", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, true)
			str := fmt.Sprintf(`
			{
				"query": "{
					book(id:\"%s\") {
						title,
						publication_year,
						slug,
						author,
						id
					}
				}"
			}
			`, testBooks[0].ID.String())
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := testBooks[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query a book by slug", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, true)
			str := fmt.Sprintf(`
			{
				"query": "{
					book(slug:\"%s\") {
						title,
						publication_year,
						slug,
						author,
						id
					}
				}"
			}
			`, testBooks[0].Slug)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := testBooks[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
	})

}

func testServer(store books.Store, withCache bool) (*books.BookServer, []books.Author, []books.Book, []books.Page) {
	cache, err := books.NewRedisCache(cnf.Cache["test"])
	if !withCache {
		cache.Available = false
	}
	if err != nil {
		log.Fatal(err)
	}
	cache.Available = false

	server, _ := books.NewBookServer(store, cache, middlewares, true)

	// values to test against
	testBooks, _ := store.Books(-1, 0)
	testAuthors, _ := store.Authors(-1, 0)
	testPages, _ := store.Pages(20, 0)
	return server, testAuthors, testBooks, testPages
}
