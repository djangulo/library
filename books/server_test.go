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
	store := testutils.NewStubStore(nil)
	cache := testutils.NewStubStore(nil)
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

	t.Run("test with cache available", func(t *testing.T) {

		t.Run("can query all books", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
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
			want := len(testutils.TestBookData())
	
			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query all books with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
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
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query all books with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 3) {
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
			got := bookArr[0].Title
			testBooks := testutils.TestBookData()
			want := testBooks[3].Title
	
			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query allBook filtered by author", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`{
				"query": "{
					allBook(author: \"Stephen King\") {
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
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query a book by id", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`
			{
				"query": "{
					book(id:\"de0e4051-54b1-4f37-97f2-619b5b568d7f\") {
						title,
						publication_year,
						slug,
						author,
						id
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
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(nil)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`
			{
				"query": "{
					book(slug:\"the-call-of-cthulu\") {
						title,
						publication_year,
						slug,
						author,
						id
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

	t.Run("test with cache unavailable", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
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
			want := len(testutils.TestBookData())
	
			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query all books with a limit", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
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
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query all books with an offset", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 3) {
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
			got := bookArr[0].Title
			testBooks := testutils.TestBookData()
			want := testBooks[3].Title
	
			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query allBook filtered by author", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`{
				"query": "{
					allBook(author: \"Stephen King\") {
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
			testutils.AssertBookStoreCalls(t, store, "list", 1)
		})
	
		t.Run("can query a book by id", func(t *testing.T) {
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`
			{
				"query": "{
					book(id:\"de0e4051-54b1-4f37-97f2-619b5b568d7f\") {
						title,
						publication_year,
						slug,
						author,
						id
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
			store := testutils.NewStubStore(nil)
			cache := testutils.NewStubStore(books.ErrCacheUnavailable)
			server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)
	
			jsonStream := []byte(`
			{
				"query": "{
					book(slug:\"the-call-of-cthulu\") {
						title,
						publication_year,
						slug,
						author,
						id
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

	t.Run("can query all pages", func(t *testing.T) {
		store := testutils.NewStubStore(nil)
		cache := testutils.NewStubStore(nil)
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
		store := testutils.NewStubStore(nil)
		cache := testutils.NewStubStore(nil)
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
		store := testutils.NewStubStore(nil)
		cache := testutils.NewStubStore(nil)
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
		store := testutils.NewStubStore(nil)
		cache := testutils.NewStubStore(nil)
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
		store := testutils.NewStubStore(nil)
		cache := testutils.NewStubStore(nil)
		server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

		jsonStream := []byte(`
		{
			"query": "{
				page(
						bookId:\"de0e4051-54b1-4f37-97f2-619b5b568d7f\",
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
}
