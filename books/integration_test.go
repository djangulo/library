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

	cases := []struct {
		name  string
		query string
		want  interface{}
	}{
		{"book by id", fmt.Sprintf(`{"query":"{book(id:\"%s\"){%s}}"}`, tBooks[0].ID.String(), bookFields), tBooks[0]},
		{"book by slug", fmt.Sprintf(`{"query":"{book(slug:\"%s\"){%s}}"}`, tBooks[0].Slug, bookFields), tBooks[0]},
		{"page by id", fmt.Sprintf(`{"query":"{page(id:\"%s\"){%s}}"}`, tPages[0].ID.String(), pageFields), tPages[0]},
		{
			"page by book+number",
			fmt.Sprintf(`{"query":"{page(book_id:\"%s\",number:%d){%s}}"}`, tPages[0].BookID.String(), tPages[0].PageNumber, pageFields),
			tPages[0],
		},
		{"author by id", fmt.Sprintf(`{"query":"{author(id:\"%s\"){%s}}"}`, tAuthors[0].ID.String(), authorFields), tAuthors[0]},
		{"author by name w name", fmt.Sprintf(`{"query":"{author(name:\"%s\"){%s}}"}`, tAuthors[0].Name, authorFields), tAuthors[0]},
		{"author by name w slug", fmt.Sprintf(`{"query":"{author(name:\"%s\"){%s}}"}`, tAuthors[0].Slug, authorFields), tAuthors[0]},
		{"allBook", fmt.Sprintf(`{"query":"{allBook{%s}}"}`, bookFields), len(tBooks)},
		{"allBook w limit", fmt.Sprintf(`{"query":"{allBook(limit:3){%s}}"}`, bookFields), 3},
		{"allBook w offset", fmt.Sprintf(`{"query":"{allBook(offset:3){%s}}"}`, bookFields), tBooks[3]},
		{"allBook by author", fmt.Sprintf(`{"query":"{allBook(author:\"%s\"){%s}}"}`, tAuthors[0].Name, bookFields), 3},
		{"allPage", fmt.Sprintf(`{"query":"{allPage{%s}}"}`, pageFields), 1000}, // default limit
		{"allPage limit override", fmt.Sprintf(`{"query":"{allPage(limit:1050){%s}}"}`, pageFields), 1050},
		{"allPage w limit", fmt.Sprintf(`{"query":"{allPage(limit:3){%s}}"}`, pageFields), 3},
		{"allPage w offset", fmt.Sprintf(`{"query":"{allPage(offset:3){%s}}"}`, pageFields), tPages[3]},
		{"allAuthor", fmt.Sprintf(`{"query":"{allAuthor{%s}}"}`, authorFields), len(tAuthors)},
		{"allAuthor w limit", fmt.Sprintf(`{"query":"{allAuthor(limit:3){%s}}"}`, authorFields), 3},
		{"allAuthor w offset", fmt.Sprintf(`{"query":"{allAuthor(offset:3){%s}}"}`, authorFields), tAuthors[3]},
	}

	t.Run("with cache", func(t *testing.T) {
		cache, dropCache := books.NewInMemoryStore("integration_test", true)
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
			var err error
			if diff := len(tPages) - i; diff < 200 {
				err = cache.BulkInsertPages(tPages[i:(i + diff)])
			} else {
				err = cache.BulkInsertPages(tPages[i:(i + 200)])
			}
			if err != nil {
				log.Fatalf("failed to bulk insert pages: %v\n", err)
			}
		}
		server, _ := books.NewBookServer(store, cache, middlewares, true)

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				stream := []byte(c.query)
				request := testutils.NewJSONPostRequest("/graphql", stream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)

				testutils.AssertStatus(t, response, http.StatusOK)

				switch want := c.want.(type) {
				case books.Book:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
					} else {
						got := testutils.GetBookFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
					}
				case books.Page:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
					} else {
						got := testutils.GetPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
					}
				case books.Author:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
					} else {
						got := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
					}
				case int:
					if strings.Contains(strings.ToLower(c.name), "book") {
						if got := testutils.GetAllBookFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "page") {
						if got := testutils.GetAllPageFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "Author") {
						if got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
						}
					}
				}

			})
		}
	})

	t.Run("no cache", func(t *testing.T) {
		cache, dropCache := books.NewInMemoryStore("unused", false)
		defer dropCache()
		server, _ := books.NewBookServer(store, cache, middlewares, true)

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				stream := []byte(c.query)
				request := testutils.NewJSONPostRequest("/graphql", stream)
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)

				testutils.AssertStatus(t, response, http.StatusOK)

				switch want := c.want.(type) {
				case books.Book:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
					} else {
						got := testutils.GetBookFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
					}
				case books.Page:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
					} else {
						got := testutils.GetPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
					}
				case books.Author:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
					} else {
						got := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
					}
				case int:
					if strings.Contains(strings.ToLower(c.name), "book") {
						if got := testutils.GetAllBookFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "page") {
						if got := testutils.GetAllPageFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "Author") {
						if got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
						}
					}
				}
			})
		}
	})

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
