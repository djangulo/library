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
	"strings"
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

func TestDatabaseIntegration(t *testing.T) {

	store, remove := testutils.NewTestSQLStore(cnf, "test")
	defer remove("test")
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "test", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}

	tPages, _ := books.PageSeedData(cnf)
	tAuthors, _ := books.AuthorSeedData(cnf)
	tBooks, _ := books.BookSeedData(cnf)

	queryOneCases := []struct {
		name   string
		query  string
		entity string
		want   interface{}
	}{
		{"book by id", fmt.Sprintf(`{"query":"{book(id:\"%s\"){%s}}"}`, tBooks[0].ID.String(), bookFields), "book", tBooks[0]},
		{"book by slug", fmt.Sprintf(`{"query":"{book(slug:\"%s\"){%s}}"}`, tBooks[0].Slug, bookFields), "book", tBooks[0]},
		{"page by id", fmt.Sprintf(`{"query":"{page(id:\"%s\"){%s}}"}`, tPages[0].ID.String(), pageFields), "page", tPages[0]},
		{
			"page by book+number",
			fmt.Sprintf(`{"query":"{page(book_id:\"%s\",number:%d){%s}}"}`, tPages[0].BookID.String(), tPages[0].PageNumber, pageFields),
			"page",
			tPages[0],
		},
		{"author by id", fmt.Sprintf(`{"query":"{author(id:\"%s\"){%s}}"}`, tAuthors[0].ID.String(), authorFields), "author", tAuthors[0]},
		{"author by name w name", fmt.Sprintf(`{"query":"{author(name:\"%s\"){%s}}"}`, tAuthors[0].Name, authorFields), "author", tAuthors[0]},
		{"author by name w slug", fmt.Sprintf(`{"query":"{author(name:\"%s\"){%s}}"}`, tAuthors[0].Slug, authorFields), "author", tAuthors[0]},
	}

	queryManyCases := []struct {
		name   string
		query  string
		entity string
		want   interface{}
	}{
		{"allBook", fmt.Sprintf(`{"query":"{allBook{%s}}"}`, bookFields), "book", len(tBooks)},
		{"allBook w limit", fmt.Sprintf(`{"query":"{allBook(limit:3){%s}}"}`, bookFields), "book", 3},
		{"allBook w offset", fmt.Sprintf(`{"query":"{allBook(offset:3){%s}}"}`, bookFields), "book", tBooks[3]},
		{"allBook by author", fmt.Sprintf(`{"query":"{allBook(author:\"%s\"){%s}}"}`, tAuthors[0].Name, bookFields), "book", 3},
		{"allPage", fmt.Sprintf(`{"query":"{allPage{%s}}"}`, pageFields), "page", 1000}, // default limit
		{"allPage limit override", fmt.Sprintf(`{"query":"{allPage(limit:1050){%s}}"}`, pageFields), "page", 1050},
		{"allPage w limit", fmt.Sprintf(`{"query":"{allPage(limit:3){%s}}"}`, pageFields), "page", 3},
		{"allPage w offset", fmt.Sprintf(`{"query":"{allPage(offset:3){%s}}"}`, pageFields), "page", tPages[3]},
		{"allAuthor", fmt.Sprintf(`{"query":"{allAuthor{%s}}"}`, authorFields), "author", len(tAuthors)},
		{"allAuthor w limit", fmt.Sprintf(`{"query":"{allAuthor(limit:3){%s}}"}`, authorFields), "author", 3},
		{"allAuthor w offset", fmt.Sprintf(`{"query":"{allAuthor(offset:3){%s}}"}`, authorFields), "author", tAuthors[3]},
	}

	t.Run("with cache", func(t *testing.T) {
		cache, dropCache := books.NewInMemoryStore("testcache", true)
		defer dropCache()
		// prepopulate cache queries
		err := cache.BulkInsertAuthors(tAuthors)
		if err != nil {
			log.Fatalf("failed to bulk insert authors: %v\n", err)
		}
		err = cache.BulkInsertBooks(tBooks)
		if err != nil {
			log.Fatalf("failed to bulk insert books: %v\n", err)
		}
		for i := 0; i < len(tPages); i += 200 {
			err = cache.BulkInsertPages(tPages[i:(i + 200)])
			if err != nil {
				log.Fatalf("failed to bulk insert pages: %v\n", err)
			}
		}
		server, _ := books.NewBookServer(store, cache, middlewares, true)

		for _, oneCase := range queryOneCases {
			t.Run(oneCase.name, func(t *testing.T) {
				stream := []byte(oneCase.query)
				request := testutils.NewJSONPostRequest("/graphql", stream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)

				var got interface{}
				switch oneCase.entity {
				case "book":
					got = testutils.GetBookFromGraphQLResponse(t, response.Body)
				case "page":
					got = testutils.GetPageFromGraphQLResponse(t, response.Body)
				case "author":
					got = testutils.GetAuthorFromGraphQLResponse(t, response.Body)
				}

				testutils.AssertStatus(t, response, http.StatusOK)
				testutils.AssertEqual(t, got, oneCase.want)
			})
		}
		for _, manyCase := range queryManyCases {
			t.Run(manyCase.name, func(t *testing.T) {
				stream := []byte(manyCase.query)
				request := testutils.NewJSONPostRequest("/graphql", stream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)

				if strings.Contains(manyCase.name, "offset") {
					var got interface{}
					switch manyCase.entity {
					case "book":
						items := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
						got = items[0]
					case "page":
						items := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						got = items[0]
					case "author":
						items := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						got = items[0]
					}
					testutils.AssertEqual(t, got, manyCase.want)
				} else {
					var got int
					switch manyCase.entity {
					case "book":
						items := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
						got = len(items)
					case "page":
						items := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						got = len(items)
					case "author":
						items := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						got = len(items)
					}
					testutils.AssertEqual(t, got, manyCase.want)
				}
				testutils.AssertStatus(t, response, http.StatusOK)
			})
		}
	})

	t.Run("no cache", func(t *testing.T) {
		cache, dropCache := books.NewInMemoryStore("nocache", false)
		defer dropCache()

		server, _ := books.NewBookServer(store, cache, middlewares, true)

		for _, oneCase := range queryOneCases {
			t.Run(oneCase.name, func(t *testing.T) {
				stream := []byte(oneCase.query)
				request := testutils.NewJSONPostRequest("/graphql", stream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)

				var got interface{}
				switch oneCase.entity {
				case "book":
					got = testutils.GetBookFromGraphQLResponse(t, response.Body)
				case "page":
					got = testutils.GetPageFromGraphQLResponse(t, response.Body)
				case "author":
					got = testutils.GetAuthorFromGraphQLResponse(t, response.Body)
				}

				testutils.AssertStatus(t, response, http.StatusOK)
				testutils.AssertEqual(t, got, oneCase.want)
			})
		}
		for _, manyCase := range queryManyCases {
			t.Run(manyCase.name, func(t *testing.T) {
				stream := []byte(manyCase.query)
				request := testutils.NewJSONPostRequest("/graphql", stream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)

				if strings.Contains(manyCase.name, "offset") {
					var got interface{}
					switch manyCase.entity {
					case "book":
						items := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
						got = items[0]
					case "page":
						items := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						got = items[0]
					case "author":
						items := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						got = items[0]
					}
					testutils.AssertEqual(t, got, manyCase.want)
				} else {
					var got int
					switch manyCase.entity {
					case "book":
						items := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
						got = len(items)
					case "page":
						items := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						got = len(items)
					case "author":
						items := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						got = len(items)
					}
					testutils.AssertEqual(t, got, manyCase.want)
				}
				testutils.AssertStatus(t, response, http.StatusOK)
			})
		}
	})

	// t.Run("can query all books", func(t *testing.T) {
	// 	// server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 	jsonStream := []byte(`
	// 			{
	// 				"query": "{
	// 					allBook {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`)
	// 	jsonStream = testutils.FlattenJSON(jsonStream)
	// 	request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 	response := httptest.NewRecorder()
	// 	server.ServeHTTP(response, request)

	// 	bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 	got := len(bookArr)
	// 	want := len(tBooks)

	// 	testutils.AssertStatus(t, response, http.StatusOK)
	// 	testutils.AssertEqual(t, got, want)
	// })

	// t.Run("Test Book queries", func(t *testing.T) {
	// 	t.Run("with cache", func(t *testing.T) {
	// 		cache, dropCache := books.NewInMemoryStore(true)
	// 		defer dropCache()
	// 		// prepopulate cache queries
	// 		err := cache.BulkInsertAuthors(tAuthors)
	// 		if err != nil {
	// 			log.Fatalf("failed to bulk insert authors: %v\n", err)
	// 		}
	// 		err = cache.BulkInsertBooks(tBooks)
	// 		if err != nil {
	// 			log.Fatalf("failed to bulk insert books: %v\n", err)
	// 		}
	// 		for i := 0; i < len(tPages); i += 200 {
	// 			err = cache.BulkInsertPages(tPages[i:(i + 200)])
	// 			if err != nil {
	// 				log.Fatalf("failed to bulk insert pages: %v\n", err)
	// 			}
	// 		}
	// 		server, _ := books.NewBookServer(store, cache, middlewares, true)

	// 		t.Run("can query all books", func(t *testing.T) {
	// 			// server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`
	// 			{
	// 				"query": "{
	// 					allBook {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			got := len(bookArr)
	// 			want := len(tBooks)

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all books with a limit", func(t *testing.T) {
	// 			// cache, dropCache := books.NewInMemoryStore(true)
	// 			// defer dropCache()
	// 			// server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allBook(limit: 3) {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			got := len(bookArr)
	// 			want := 3

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all books with an offset", func(t *testing.T) {
	// 			// cache, dropCache := books.NewInMemoryStore(true)
	// 			// defer dropCache()
	// 			// server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allBook(offset: 1) {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			if len(bookArr) == 0 {
	// 				t.Fatal("expected a result but got none")
	// 			}
	// 			got := bookArr[0].Title
	// 			want := tBooks[1].Title

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query allBook filtered by author", func(t *testing.T) {
	// 			// cache, dropCache := books.NewInMemoryStore(true)
	// 			// defer dropCache()
	// 			// server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`{
	// 				"query": "{
	// 					allBook(author: \"%s\") {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`, tAuthors[0].Name)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			got := len(bookArr)
	// 			want := 3

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query a book by id", func(t *testing.T) {
	// 			// cache, dropCache := books.NewInMemoryStore(true)
	// 			// defer dropCache()
	// 			// server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					book(id:\"%s\") {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}
	// 			`, tBooks[0].ID.String())
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
	// 			got := book.ID
	// 			want := tBooks[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query a book by slug", func(t *testing.T) {
	// 			// cache, dropCache := books.NewInMemoryStore(true)
	// 			// defer dropCache()
	// 			// server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					book(slug:\"%s\") {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}
	// 			`, tBooks[0].Slug)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
	// 			got := book.ID
	// 			want := tBooks[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 	})

	// 	t.Run("without cache", func(t *testing.T) {
	// 		t.Run("can query all books", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore("unused", false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`
	// 			{
	// 				"query": "{
	// 					allBook {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			got := len(bookArr)
	// 			want := len(tBooks)

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all books with a limit", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allBook(limit: 3) {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			got := len(bookArr)
	// 			want := 3

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all books with an offset", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allBook(offset: 1) {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			if len(bookArr) == 0 {
	// 				t.Fatal("expected a result but got none")
	// 			}
	// 			got := bookArr[0].Title
	// 			want := tBooks[1].Title

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query allBook filtered by author", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`{
	// 				"query": "{
	// 					allBook(author: \"%s\") {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`, tAuthors[0].Name)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 			got := len(bookArr)
	// 			want := 3

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query a book by id", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					book(id:\"%s\") {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}
	// 			`, tBooks[0].ID.String())
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
	// 			got := book.ID
	// 			want := tBooks[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query a book by slug", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					book(slug:\"%s\") {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}
	// 			`, tBooks[0].Slug)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
	// 			got := book.ID
	// 			want := tBooks[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 	})
	// })

	// t.Run("Test Page queries", func(t *testing.T) {
	// 	t.Run("with cache", func(t *testing.T) {
	// 		/*

	// 			Test omitted due to large number of sample pages.

	// 		*/
	// 		// t.Run("can query all pages", func(t *testing.T) {
	// 		// 	server, _, _, tPages := testServer(store, true)
	// 		// 	jsonStream := []byte(`
	// 		// 	{
	// 		// 		"query": "{
	// 		// 			allPage(limit) {
	// 		// 				id,
	// 		// 				page_number,
	// 		// 				book_id,
	// 		// 				body,
	// 		// 			}
	// 		// 		}"
	// 		// 	}`)
	// 		// 	jsonStream = testutils.FlattenJSON(jsonStream)
	// 		// 	request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 		// 	response := httptest.NewRecorder()
	// 		// 	server.ServeHTTP(response, request)

	// 		// 	pages := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 		// 	got := len(pages)
	// 		// 	want := len(tBooks)

	// 		// 	testutils.AssertStatus(t, response, http.StatusOK)
	// 		// 	testutils.AssertEqual(t, got, want)
	// 		// })

	// 		t.Run("can query all pages with a limit", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allPage(limit: 10) {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
	// 			got := len(pages)
	// 			want := 10

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all pages with an offset", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allPage(offset: 5, limit: 20) {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
	// 			var got int
	// 			if len(pages) > 0 {
	// 				got = pages[0].PageNumber
	// 			}
	// 			t.Errorf("expected a result but got none %v", pages)
	// 			want := tPages[5].PageNumber

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query a page by id", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					page(id:\"%s\") {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body
	// 					}
	// 				}"
	// 			}
	// 			`, tPages[0].ID.String())
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
	// 			got := page.ID
	// 			want := tPages[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query a page by book id and page number", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					page(
	// 							book_id:\"%s\",
	// 							number: %d
	// 						) {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body,
	// 					}
	// 				}"
	// 			}
	// 			`, tPages[0].BookID.String(), tPages[0].PageNumber)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
	// 			got := page.ID
	// 			want := tPages[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 	})

	// 	t.Run("without cache", func(t *testing.T) {
	// 		/*

	// 			Test omitted due to large number of sample pages.

	// 		*/
	// 		// t.Run("can query all pages", func(t *testing.T) {
	// 		// 	server, _, _, tPages := testServer(store, false)
	// 		// 	jsonStream := []byte(`
	// 		// 	{
	// 		// 		"query": "{
	// 		// 			allPage(limit) {
	// 		// 				id,
	// 		// 				page_number,
	// 		// 				book_id,
	// 		// 				body,
	// 		// 			}
	// 		// 		}"
	// 		// 	}`)
	// 		// 	jsonStream = testutils.FlattenJSON(jsonStream)
	// 		// 	request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 		// 	response := httptest.NewRecorder()
	// 		// 	server.ServeHTTP(response, request)

	// 		// 	pages := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
	// 		// 	got := len(pages)
	// 		// 	want := len(tBooks)

	// 		// 	testutils.AssertStatus(t, response, http.StatusOK)
	// 		// 	testutils.AssertEqual(t, got, want)
	// 		// })

	// 		t.Run("can query all pages with a limit", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allPage(limit: 10) {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
	// 			got := len(pages)
	// 			want := 10

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all pages with an offset", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allPage(offset: 5, limit: 20) {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
	// 			if len(pages) == 0 {
	// 				t.Error("expected a result but got none")
	// 			}
	// 			got := pages[0].PageNumber
	// 			want := tPages[5].PageNumber

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query a page by id", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					page(id:\"%s\") {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body
	// 					}
	// 				}"
	// 			}
	// 			`, tPages[0].ID.String())
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
	// 			got := page.ID
	// 			want := tPages[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query a page by book id and page number", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					page(
	// 						book_id:\"%s\",
	// 							number: %d
	// 						) {
	// 						id,
	// 						page_number,
	// 						book_id,
	// 						body,
	// 					}
	// 				}"
	// 			}
	// 			`, tPages[0].BookID.String(), tPages[0].PageNumber)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
	// 			got := page.ID
	// 			want := tPages[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 	})
	// })

	// t.Run("Test Author queries", func(t *testing.T) {
	// 	t.Run("with cache", func(t *testing.T) {
	// 		t.Run("can query all authors", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`
	// 			{
	// 				"query": "{
	// 					allAuthor {
	// 						id,
	// 						name,
	// 						slug,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
	// 			got := len(authors)
	// 			want := len(tAuthors)

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all authors with a limit", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allAuthor(limit: 2) {
	// 						id,
	// 						name,
	// 						slug,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
	// 			got := len(authors)
	// 			want := 2

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all authors with an offset", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allAuthor(offset: 2) {
	// 						id,
	// 						name,
	// 						slug,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
	// 			if len(authors) == 0 {
	// 				t.Fatal("expected a result but got none")
	// 			}
	// 			got := authors[0].Name
	// 			want := tAuthors[2].Name

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query an author by id", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					author(id:\"%s\") {
	// 						id,
	// 						name,
	// 						slug,
	// 					}
	// 				}"
	// 			}
	// 			`, tAuthors[0].ID.String())
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
	// 			got := author.ID
	// 			want := tAuthors[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query an author by name using name", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					author(name:\"%s\") {
	// 						id,
	// 						slug,
	// 						name
	// 					}
	// 				}"
	// 			}
	// 			`, tAuthors[0].Name)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
	// 			got := author.ID
	// 			want := tAuthors[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query an author by name using slug", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(true)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					author(name:\"%s\") {
	// 						id,
	// 						slug,
	// 						name
	// 					}
	// 				}"
	// 			}
	// 			`, tAuthors[0].Slug)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
	// 			got := author.ID
	// 			want := tAuthors[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 	})

	// 	t.Run("without cache", func(t *testing.T) {
	// 		t.Run("can query all authors", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`
	// 			{
	// 				"query": "{
	// 					allAuthor {
	// 						id,
	// 						name,
	// 						slug
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
	// 			got := len(authors)
	// 			want := len(tAuthors)

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all authors with a limit", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allAuthor(limit: 2) {
	// 						id,
	// 						name,
	// 						slug,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
	// 			got := len(authors)
	// 			want := 2

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query all authors with an offset", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			jsonStream := []byte(`{
	// 				"query": "{
	// 					allAuthor(offset: 2) {
	// 						id,
	// 						name,
	// 						slug,
	// 					}
	// 				}"
	// 			}`)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
	// 			if len(authors) == 0 {
	// 				t.Fatal("expected a result but got none")
	// 			}
	// 			got := authors[0].Name
	// 			want := tAuthors[2].Name

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertEqual(t, got, want)
	// 		})

	// 		t.Run("can query an author by id", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					author(id:\"%s\") {
	// 						id,
	// 						name,
	// 						slug,
	// 					}
	// 				}"
	// 			}
	// 			`, tAuthors[0].ID.String())
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
	// 			got := author.ID
	// 			want := tAuthors[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query an author by name using name", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					author(name:\"%s\") {
	// 						id,
	// 						slug,
	// 						name
	// 					}
	// 				}"
	// 			}
	// 			`, tAuthors[0].Name)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
	// 			got := author.ID
	// 			want := tAuthors[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 		t.Run("can query an author by name using slug", func(t *testing.T) {
	// 			cache, dropCache := books.NewInMemoryStore(false)
	// 			defer dropCache()
	// 			server, _ := books.NewBookServer(store, cache, middlewares, true)
	// 			str := fmt.Sprintf(`
	// 			{
	// 				"query": "{
	// 					author(name:\"%s\") {
	// 						id,
	// 						slug,
	// 						name
	// 					}
	// 				}"
	// 			}
	// 			`, tAuthors[0].Slug)
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)

	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)

	// 			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
	// 			got := author.ID
	// 			want := tAuthors[0].ID

	// 			testutils.AssertStatus(t, response, http.StatusOK)
	// 			testutils.AssertUUIDsEqual(t, got, want)
	// 		})
	// 	})
	// })

}

func BenchmarkServer(b *testing.B) {
	store, remove := testutils.NewTestSQLStore(cnf, "bench")
	defer remove("bench")
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "bench", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}
	offset := 1000
	tPages, _ := books.PageSeedData(cnf)
	tAuthors, _ := books.AuthorSeedData(cnf)
	tBooks, _ := books.BookSeedData(cnf)

	b.Run("without cache", func(b *testing.B) {
		cache, dropCache := books.NewInMemoryStore("nocache", false)
		defer dropCache()
		server, _ := books.NewBookServer(store, cache, middlewares, true)

		pageQueries := make([]string, 0)
		for i := range []int{0, 1, 2, 3, 4} {
			str := fmt.Sprintf(`{
				"query": "{
					allPage(limit: 1000, offset: %d) {
						id,
						page_number,
						book_id
					}
				}"
			}`, int(i*offset))
			pageQueries = append(pageQueries, str)
		}
		b.Run("benchmark allPage queries", func(b *testing.B) {
			for i, q := range pageQueries {
				b.Run(fmt.Sprintf("query pages lim:1000 offset: %d", i*offset), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						jsonStream := []byte(q)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
			}
		})
		b.Run("benchmark allBook query", func(b *testing.B) {
			for k := 0; k < b.N; k++ {
				str := `{
					"query": "{
						allBook {
							id,
							title,
							slug,
							author_id,
							file,
							source,
							publication_year,
							page_count
						}
					}"
				}`
				jsonStream := []byte(str)
				jsonStream = testutils.FlattenJSON(jsonStream)
				request := testutils.NewJSONPostRequest("/graphql", jsonStream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
			}
		})
		b.Run("benchmark allAuthor queries", func(b *testing.B) {
			for k := 0; k < b.N; k++ {
				str := `{
					"query": "{
						allAuthor {
							id,
							name,
							slug
						}
					}"
				}`
				jsonStream := []byte(str)
				jsonStream = testutils.FlattenJSON(jsonStream)
				request := testutils.NewJSONPostRequest("/graphql", jsonStream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
			}
		})
		b.Run("benchmark individual book queries", func(b *testing.B) {
			for _, book := range tBooks[:2] {
				b.Run(fmt.Sprintf("BookByID(%s)", book.ID.String()), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								book(id:\"%s\") {
									id,
									title,
									slug,
									author_id,
									file,
									source,
									publication_year,
									page_count
								}
							}"
						}`, book.ID.String())
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
				b.Run(fmt.Sprintf("BookBySlug(%s)", book.Slug), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								book(slug:\"%s\") {
									id,
									title,
									slug,
									author_id,
									file,
									source,
									publication_year,
									page_count
								}
							}"
						}`, book.Slug)
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
			}
		})
		b.Run("benchmark individual author queries", func(b *testing.B) {
			for _, author := range tAuthors[:2] {
				b.Run(fmt.Sprintf("AuthorByID(%s)", author.ID.String()), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								author(id:\"%s\") {
									id,
									slug,
									name
								}
							}"
						}`, author.ID.String())
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
				b.Run(fmt.Sprintf("AuthorBySlug(%s)", author.Slug), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								author(name:\"%s\") {
									id,
									slug,
									name
								}
							}"
						}`, author.Slug)
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
			}
		})
	})

	b.Run("with sqlite in-memory cache", func(b *testing.B) {
		cache, dropCache := books.NewInMemoryStore("benchcache", false)
		defer dropCache()

		// prepopulate cache queries
		err := cache.BulkInsertAuthors(tAuthors)
		if err != nil {
			log.Fatalf("failed to bulk insert authors: %v\n", err)
		}
		err = cache.BulkInsertBooks(tBooks)
		if err != nil {
			log.Fatalf("failed to bulk insert books: %v\n", err)
		}
		for i := 0; i < len(tPages); i += 200 {
			err = cache.BulkInsertPages(tPages[i:(i + 200)])
			if err != nil {
				log.Fatalf("failed to bulk insert pages: %v\n", err)
			}
		}
		server, _ := books.NewBookServer(store, cache, middlewares, true)

		log.Println("printing le books")
		books, err := cache.Books(5, 0)
		if err != nil {
			log.Println(err)
		}
		log.Println(books)

		pageQueries := make([]string, 0)
		for i := range []int{0, 1, 2, 3, 4} {
			str := fmt.Sprintf(`{
				"query": "{
					allPage(limit: 1000, offset: %d) {
						id,
						page_number,
						book_id
					}
				}"
			}`, int(i*offset))
			pageQueries = append(pageQueries, str)
		}
		b.Run("benchmark allPage queries", func(b *testing.B) {
			for i, q := range pageQueries {
				b.Run(fmt.Sprintf("query pages lim:1000 offset: %d", i*offset), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						jsonStream := []byte(q)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
			}
		})
		b.Run("benchmark allBook query", func(b *testing.B) {
			for k := 0; k < b.N; k++ {
				str := `{
					"query": "{
						allBook {
							id,
							title,
							slug,
							author_id,
							file,
							source,
							publication_year,
							page_count
						}
					}"
				}`
				jsonStream := []byte(str)
				jsonStream = testutils.FlattenJSON(jsonStream)
				request := testutils.NewJSONPostRequest("/graphql", jsonStream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
			}
		})
		b.Run("benchmark allAuthor queries", func(b *testing.B) {
			for k := 0; k < b.N; k++ {
				str := `{
					"query": "{
						allAuthor {
							id,
							name,
							slug
						}
					}"
				}`
				jsonStream := []byte(str)
				jsonStream = testutils.FlattenJSON(jsonStream)
				request := testutils.NewJSONPostRequest("/graphql", jsonStream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
			}
		})
		b.Run("benchmark individual book queries", func(b *testing.B) {
			for _, book := range tBooks[:2] {
				b.Run(fmt.Sprintf("BookByID(%s)", book.ID.String()), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								book(id:\"%s\") {
									id,
									title,
									slug,
									author_id,
									file,
									source,
									publication_year,
									page_count
								}
							}"
						}`, book.ID.String())
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
				b.Run(fmt.Sprintf("BookBySlug(%s)", book.Slug), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								book(slug:\"%s\") {
									id,
									title,
									slug,
									author_id,
									file,
									source,
									publication_year,
									page_count
								}
							}"
						}`, book.Slug)
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
			}
		})
		b.Run("benchmark individual author queries", func(b *testing.B) {
			for _, author := range tAuthors[:2] {
				b.Run(fmt.Sprintf("AuthorByID(%s)", author.ID.String()), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								author(id:\"%s\") {
									id,
									slug,
									name
								}
							}"
						}`, author.ID.String())
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
				b.Run(fmt.Sprintf("AuthorBySlug(%s)", author.Slug), func(b *testing.B) {
					for k := 0; k < b.N; k++ {
						str := fmt.Sprintf(`{
							"query": "{
								author(name:\"%s\") {
									id,
									slug,
									name
								}
							}"
						}`, author.Slug)
						jsonStream := []byte(str)
						jsonStream = testutils.FlattenJSON(jsonStream)
						request := testutils.NewJSONPostRequest("/graphql", jsonStream)
						response := httptest.NewRecorder()
						server.ServeHTTP(response, request)
					}
				})
			}
		})
	})

	// b.Run("with go-redis cache", func(b *testing.B) {
	// 	cache, err := books.NewRedisCache(cnf.Cache["goredis_bench"])
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	// prepopulate cache queries
	// 	for i := 0; i < 5; i++ {
	// 		key := fmt.Sprintf("Pages(%d,%d)", offset, i*offset)
	// 		cache.SavePageQuery(key, tPages[(i*offset):(i*offset+offset)])
	// 	}
	// 	for i := 0; i < len(tPages); i++ {
	// 		cache.InsertPage(tPages[i])
	// 	}
	// 	cache.SaveBookQuery("Books(1000,0)", tBooks)
	// 	for i := 0; i < len(tBooks[:2]); i++ {
	// 		cache.InsertBook(tBooks[i])
	// 	}
	// 	cache.SaveAuthorQuery("Authors(1000,0)", tAuthors)
	// 	for i := 0; i < len(tAuthors[:2]); i++ {
	// 		cache.InsertAuthor(tAuthors[i])
	// 	}

	// 	server, _ := books.NewBookServer(store, cache, middlewares, true)

	// 	pageQueries := make([]string, 0)
	// 	for i := range []int{0, 1, 2, 3, 4} {
	// 		str := fmt.Sprintf(`{
	// 			"query": "{
	// 				allPage(limit: 1000, offset: %d) {
	// 					id,
	// 					page_number,
	// 					book_id
	// 				}
	// 			}"
	// 		}`, int(i*offset))
	// 		pageQueries = append(pageQueries, str)
	// 	}
	// 	b.Run("benchmark allPage queries", func(b *testing.B) {
	// 		for i, q := range pageQueries {
	// 			b.Run(fmt.Sprintf("query pages lim:1000 offset: %d", i*offset), func(b *testing.B) {
	// 				for k := 0; k < b.N; k++ {
	// 					jsonStream := []byte(q)
	// 					jsonStream = testutils.FlattenJSON(jsonStream)
	// 					request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 					response := httptest.NewRecorder()
	// 					server.ServeHTTP(response, request)
	// 				}
	// 			})
	// 		}
	// 	})
	// 	b.Run("benchmark allBook query", func(b *testing.B) {
	// 		for k := 0; k < b.N; k++ {
	// 			str := `{
	// 				"query": "{
	// 					allBook {
	// 						id,
	// 						title,
	// 						slug,
	// 						author_id,
	// 						file,
	// 						source,
	// 						publication_year,
	// 						page_count
	// 					}
	// 				}"
	// 			}`
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)
	// 		}
	// 	})
	// 	b.Run("benchmark allAuthor queries", func(b *testing.B) {
	// 		for k := 0; k < b.N; k++ {
	// 			str := `{
	// 				"query": "{
	// 					allAuthor {
	// 						id,
	// 						name,
	// 						slug
	// 					}
	// 				}"
	// 			}`
	// 			jsonStream := []byte(str)
	// 			jsonStream = testutils.FlattenJSON(jsonStream)
	// 			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 			response := httptest.NewRecorder()
	// 			server.ServeHTTP(response, request)
	// 		}
	// 	})
	// 	b.Run("benchmark individual book queries", func(b *testing.B) {
	// 		for _, book := range tBooks[:2] {
	// 			b.Run(fmt.Sprintf("BookByID(%s)", book.ID.String()), func(b *testing.B) {
	// 				for k := 0; k < b.N; k++ {
	// 					str := fmt.Sprintf(`{
	// 						"query": "{
	// 							book(id:\"%s\") {
	// 								id,
	// 								title,
	// 								slug,
	// 								author_id,
	// 								file,
	// 								source,
	// 								publication_year,
	// 								page_count
	// 							}
	// 						}"
	// 					}`, book.ID.String())
	// 					jsonStream := []byte(str)
	// 					jsonStream = testutils.FlattenJSON(jsonStream)
	// 					request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 					response := httptest.NewRecorder()
	// 					server.ServeHTTP(response, request)
	// 				}
	// 			})
	// 			b.Run(fmt.Sprintf("BookBySlug(%s)", book.Slug), func(b *testing.B) {
	// 				for k := 0; k < b.N; k++ {
	// 					str := fmt.Sprintf(`{
	// 						"query": "{
	// 							book(slug:\"%s\") {
	// 								id,
	// 								title,
	// 								slug,
	// 								author_id,
	// 								file,
	// 								source,
	// 								publication_year,
	// 								page_count
	// 							}
	// 						}"
	// 					}`, book.Slug)
	// 					jsonStream := []byte(str)
	// 					jsonStream = testutils.FlattenJSON(jsonStream)
	// 					request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 					response := httptest.NewRecorder()
	// 					server.ServeHTTP(response, request)
	// 				}
	// 			})
	// 		}
	// 	})
	// 	b.Run("benchmark individual author queries", func(b *testing.B) {
	// 		for _, author := range tAuthors[:2] {
	// 			b.Run(fmt.Sprintf("AuthorByID(%s)", author.ID.String()), func(b *testing.B) {
	// 				for k := 0; k < b.N; k++ {
	// 					str := fmt.Sprintf(`{
	// 						"query": "{
	// 							author(id:\"%s\") {
	// 								id,
	// 								slug,
	// 								name
	// 							}
	// 						}"
	// 					}`, author.ID.String())
	// 					jsonStream := []byte(str)
	// 					jsonStream = testutils.FlattenJSON(jsonStream)
	// 					request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 					response := httptest.NewRecorder()
	// 					server.ServeHTTP(response, request)
	// 				}
	// 			})
	// 			b.Run(fmt.Sprintf("AuthorBySlug(%s)", author.Slug), func(b *testing.B) {
	// 				for k := 0; k < b.N; k++ {
	// 					str := fmt.Sprintf(`{
	// 						"query": "{
	// 							author(name:\"%s\") {
	// 								id,
	// 								slug,
	// 								name
	// 							}
	// 						}"
	// 					}`, author.Slug)
	// 					jsonStream := []byte(str)
	// 					jsonStream = testutils.FlattenJSON(jsonStream)
	// 					request := testutils.NewJSONPostRequest("/graphql", jsonStream)
	// 					response := httptest.NewRecorder()
	// 					server.ServeHTTP(response, request)
	// 				}
	// 			})
	// 		}
	// 	})
	// })
}

// func testServer(store books.Store, withCache bool) (*books.BookServer, []books.Author, []books.Book, []books.Page) {
// 	cache, removeCache := books.NewInMemoryStore(true)
// 	cache, err := books.NewRedisCache(cnf.Cache["test"])
// 	if !withCache {
// 		cache.Available = false
// 	}
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	cache.Available = false

// 	server, _ := books.NewBookServer(store, cache, middlewares, true)

// 	// values to test against
// 	tBooks, _ := store.Books(-1, 0)
// 	tAuthors, _ := store.Authors(-1, 0)
// 	tPages, _ := store.Pages(20, 0)
// 	return server, tAuthors, tBooks, tPages
// }
