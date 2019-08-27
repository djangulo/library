package books_test

import (
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/books/testutils"
	"github.com/gofrs/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRoot(t *testing.T) {
	store := testutils.NewStubStore(true)
	cache := testutils.NewStubCache(nil)
	server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	t.Run("redirects to /en on /", func(t *testing.T) {

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if response.Code != 302 {
			t.Errorf("want 302 got %d", response.Code)
		}

	})

	t.Run("GET /en returns the english template", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/en", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("want 200 got %d", response.Code)
		}
	})

	t.Run("GET /es returns the spanish template", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/es", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("want 200 got %d", response.Code)
		}
	})
}

func TestGraphQLBookQueries(t *testing.T) {

	t.Run("with cache", func(t *testing.T) {

		t.Run("can query all books", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
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
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := len(testutils.TestBookData())

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Books(1000,0)"), 1)
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allBook(limit: 3) {
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
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Books(3,0)"), 1)
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 3) {
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
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := bookArr[0].Title
			testBooks := testutils.TestBookData()
			want := testBooks[3].Title

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Books(1000,3)"), 1)
		})

		t.Run("can query allBook filtered by author", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allBook(author: \"Stephen King\") {
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
			testutils.AssertCacheQueryCalls(t, cache, ("SET:BooksByAuthor(Stephen King)"), 1)
		})

		t.Run("can query a book by id", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					book(id:\"de0e4051-54b1-4f37-97f2-619b5b568d7f\") {
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
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := uuid.Must(uuid.FromString("de0e4051-54b1-4f37-97f2-619b5b568d7f"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertBookCacheCalls(t, cache, got.String(), 1)
		})
		t.Run("can query a book by slug", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					book(slug:\"the-call-of-cthulu\") {
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
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := uuid.Must(uuid.FromString("8c79ac56-39f2-4954-8a1f-cd3b058c169f"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertBookCacheCalls(t, cache, got.String(), 1)
		})
	})

	t.Run("without cache", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
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
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := len(bookArr)
			want := len(testutils.TestBookData())

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allBook(limit: 3) {
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
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 3) {
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
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			bookArr := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
			got := bookArr[0].Title
			testBooks := testutils.TestBookData()
			want := testBooks[3].Title

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})

		t.Run("can query allBook filtered by author", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allBook(author: \"Stephen King\") {
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
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})

		t.Run("can query a book by id", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					book(id:\"de0e4051-54b1-4f37-97f2-619b5b568d7f\") {
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
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := uuid.Must(uuid.FromString("de0e4051-54b1-4f37-97f2-619b5b568d7f"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, got.String(), 1)
		})
		t.Run("can query a book by slug", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					book(slug:\"the-call-of-cthulu\") {
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
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			book := testutils.GetBookFromGraphQLResponse(t, response.Body)
			got := book.ID
			want := uuid.Must(uuid.FromString("8c79ac56-39f2-4954-8a1f-cd3b058c169f"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, got.String(), 1)
		})
	})
}

func TestGraphQLPageQueries(t *testing.T) {

	t.Run("with cache", func(t *testing.T) {

		t.Run("can query all pages", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					allPage {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
			got := len(pages)
			want := len(testutils.TestPageData())

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Pages(1000,0)"), 1)
		})

		t.Run("can query all pages with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allPage(limit: 10) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
			got := len(pages)
			want := 10

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Pages(10,0)"), 1)
		})

		t.Run("can query all pages with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allPage(offset: 5, limit: 20) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
			got := pages[0].PageNumber
			testPages := testutils.TestPageData()
			want := testPages[5].PageNumber

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Pages(20,5)"), 1)
		})

		t.Run("can query a page by id", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					page(id: \"05f2dd7f-7b42-4d7c-9c25-859f1146ad68\") {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := uuid.Must(uuid.FromString("05f2dd7f-7b42-4d7c-9c25-859f1146ad68"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertPageCacheCalls(t, cache, got.String(), 1)
		})
		t.Run("can query a page by book id and page number", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					page(
							book_id:\"de0e4051-54b1-4f37-97f2-619b5b568d7f\",
							number: 1
						) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := uuid.Must(uuid.FromString("05f2dd7f-7b42-4d7c-9c25-859f1146ad68"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertPageCacheCalls(t, cache, got.String(), 1)
		})
	})

	t.Run("test with cache unavailable", func(t *testing.T) {
		t.Run("can query all pages", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					allPage {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
			got := len(pages)
			want := len(testutils.TestPageData())

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertPageStoreCalls(t, store, "list", 1)
		})

		t.Run("can query all pages with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allPage(limit: 10) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
			got := len(pages)
			want := 10

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertPageStoreCalls(t, store, "list", 1)
		})

		t.Run("can query all pages with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allPage(offset: 5, limit: 20) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			pages := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
			got := pages[0].PageNumber
			testPages := testutils.TestPageData()
			want := testPages[5].PageNumber

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertPageStoreCalls(t, store, "list", 1)
		})

		t.Run("can query a page by id", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					page(id: \"05f2dd7f-7b42-4d7c-9c25-859f1146ad68\") {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := uuid.Must(uuid.FromString("05f2dd7f-7b42-4d7c-9c25-859f1146ad68"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertPageStoreCalls(t, store, got.String(), 1)
		})
		t.Run("can query a page by book id and page number", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					page(
							book_id:\"de0e4051-54b1-4f37-97f2-619b5b568d7f\",
							number: 1
						) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := uuid.Must(uuid.FromString("05f2dd7f-7b42-4d7c-9c25-859f1146ad68"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertPageStoreCalls(t, store, got.String(), 1)
		})
	})
}

func TestGraphQLAuthorQueries(t *testing.T) {

	t.Run("with cache", func(t *testing.T) {

		t.Run("can query all authors", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					allAuthor {
						id,
						name,
						slug,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
			got := len(authors)
			want := len(testutils.TestAuthorData())

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Authors(1000,0)"), 1)
		})

		t.Run("can query all authors with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allAuthor(limit: 2) {
						id,
						name,
						slug,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
			got := len(authors)
			want := 2

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Authors(2,0)"), 1)
		})

		t.Run("can query all authors with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allAuthor(offset: 2) {
						id,
						name,
						slug,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
			got := authors[0].Name
			testAuthors := testutils.TestAuthorData()
			want := testAuthors[2].Name

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertCacheQueryCalls(t, cache, ("SET:Authors(1000,2)"), 1)
		})

		t.Run("can query an author by id", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					author(id: \"f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d\") {
						id,
						name,
						slug,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := uuid.Must(uuid.FromString("f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertAuthorCacheCalls(t, cache, got.String(), 1)
		})
		t.Run("can query an author by name using name", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					author(name:\"Herman Melville\") {
						id,
						name,
						slug,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := uuid.Must(uuid.FromString("f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertAuthorCacheCalls(t, cache, got.String(), 1)
		})
		t.Run("can query an author by name using slug", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					author(name:\"herman-melville\") {
						id,
						name,
						slug,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := uuid.Must(uuid.FromString("f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertAuthorCacheCalls(t, cache, got.String(), 1)
		})
	})

	t.Run("without cache", func(t *testing.T) {
		t.Run("can query all authors", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					allAuthor {
						id,
						name,
						slug,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
			got := len(authors)
			want := len(testutils.TestAuthorData())

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertAuthorStoreCalls(t, store, "list", 1)
		})

		t.Run("can query all auhtors with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allAuthor(limit: 2) {
						id,
						name,
						slug,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
			got := len(authors)
			want := 2

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertAuthorStoreCalls(t, store, "list", 1)
		})

		t.Run("can query all authors with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`{
				"query": "{
					allAuthor(offset: 2) {
						id,
						name,
						slug,
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
			got := authors[0].Name
			testAuthors := testutils.TestAuthorData()
			want := testAuthors[2].Name

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertAuthorStoreCalls(t, store, "list", 1)
		})

		t.Run("can query an author by id", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					author(id: \"f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d\") {
						id,
						name,
						slug,
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := uuid.Must(uuid.FromString("f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertAuthorStoreCalls(t, store, got.String(), 1)
		})
		t.Run("can query an author by name using name", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					author(name:\"Herman Melville\") {
						id,
						name,
						slug
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := uuid.Must(uuid.FromString("f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertAuthorStoreCalls(t, store, got.String(), 1)
		})
		t.Run("can query an author by name using slug", func(t *testing.T) {
			store := testutils.NewStubStore(true)
			cache := testutils.NewStubCache(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

			jsonStream := []byte(`
			{
				"query": "{
					author(name:\"herman-melville\") {
						id,
						name,
						slug
					}
				}"
			}
			`)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := uuid.Must(uuid.FromString("f32ad0c4-0c2e-4f4d-b0b8-f5ace440bd9d"))

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
			testutils.AssertAuthorStoreCalls(t, store, got.String(), 1)
		})
	})
}
