package books_test

import (
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRoot(t *testing.T) {
	store := books.NewStubStore()
	server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)
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

	t.Run("can query all books", func(t *testing.T) {
		store := books.NewStubStore()
		server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)

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
		jsonStream = books.FlattenJSON(jsonStream)
		request := books.NewJSONPostRequest("/graphql", jsonStream)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		bookArr := books.GetAllBookFromGraphQLResponse(t, response.Body)
		got := len(bookArr)
		want := len(books.TestBookData())

		books.AssertStatus(t, response, http.StatusOK)
		books.AssertEqual(t, got, want)
		books.AssertBookStoreCalls(t, store.BookCalls["list"], 1)
	})

	t.Run("can query all books with a limit", func(t *testing.T) {
		store := books.NewStubStore()
		server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)

		jsonStream := []byte(`{
			"query": "{
				allBook(limit: 3) {
					title,
					pages{id}
				}
			}"
		}`)
		jsonStream = books.FlattenJSON(jsonStream)
		request := books.NewJSONPostRequest("/graphql", jsonStream)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		bookArr := books.GetAllBookFromGraphQLResponse(t, response.Body)
		got := len(bookArr)
		want := 3

		books.AssertStatus(t, response, http.StatusOK)
		books.AssertEqual(t, got, want)
		books.AssertBookStoreCalls(t, store.BookCalls["list"], 1)
	})

	t.Run("can query all books with an offset", func(t *testing.T) {
		store := books.NewStubStore()
		server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)

		jsonStream := []byte(`{
			"query": "{
				allBook(offset: 3) {
					title,
					pages{id}
				}
			}"
		}`)
		jsonStream = books.FlattenJSON(jsonStream)
		request := books.NewJSONPostRequest("/graphql", jsonStream)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		bookArr := books.GetAllBookFromGraphQLResponse(t, response.Body)
		got := bookArr[0].Title
		testBooks := books.TestBookData()
		want := testBooks[3].Title

		books.AssertStatus(t, response, http.StatusOK)
		books.AssertEqual(t, got, want)
		books.AssertBookStoreCalls(t, store.BookCalls["list"], 1)
	})

	t.Run("can query allBook filtered by author", func(t *testing.T) {
		store := books.NewStubStore()
		server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)

		jsonStream := []byte(`{
			"query": "{
				allBook(author: \"Stephen King\") {
					title,
					pages{id}
				}
			}"
		}`)
		jsonStream = books.FlattenJSON(jsonStream)
		request := books.NewJSONPostRequest("/graphql", jsonStream)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		bookArr := books.GetAllBookFromGraphQLResponse(t, response.Body)
		got := len(bookArr)
		want := 3

		books.AssertStatus(t, response, http.StatusOK)
		books.AssertEqual(t, got, want)
		books.AssertBookStoreCalls(t, store.BookCalls["list"], 1)
	})

	t.Run("can query a book by id", func(t *testing.T) {
		store := books.NewStubStore()
		server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)

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
		jsonStream = books.FlattenJSON(jsonStream)
		request := books.NewJSONPostRequest("/graphql", jsonStream)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		book := books.GetBookFromGraphQLResponse(t, response.Body)
		got := book.ID
		want := uuid.Must(uuid.FromString("de0e4051-54b1-4f37-97f2-619b5b568d7f"))

		books.AssertStatus(t, response, http.StatusOK)
		books.AssertUUIDsEqual(t, got, want)
		books.AssertBookStoreCalls(
			t,
			store.BookCalls["de0e4051-54b1-4f37-97f2-619b5b568d7f"],
			1,
		)
	})
	t.Run("can query a book by slug", func(t *testing.T) {
		store := books.NewStubStore()
		server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)

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
		jsonStream = books.FlattenJSON(jsonStream)

		request := books.NewJSONPostRequest("/graphql", jsonStream)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		book := books.GetBookFromGraphQLResponse(t, response.Body)
		got := book.ID
		want := uuid.Must(uuid.FromString("8c79ac56-39f2-4954-8a1f-cd3b058c169f"))

		books.AssertStatus(t, response, http.StatusOK)
		books.AssertUUIDsEqual(t, got, want)
		books.AssertBookStoreCalls(
			t,
			store.BookCalls["8c79ac56-39f2-4954-8a1f-cd3b058c169f"],
			1,
		)
	})
}
