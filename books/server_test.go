package books_test

import (
	"bytes"
	"encoding/json"
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
	"net/http"
	"net/http/httptest"
	"reflect"
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

	t.Run("GET /es returns the english template", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/es", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if response.Code != 200 {
			t.Errorf("want 200 got %d", response.Code)
		}
	})
}

func TestGraphQL(t *testing.T) {
	store := books.NewStubStore()
	server, _ := books.NewBookServer(store, books.DummyMiddlewares, true)

	t.Run("can query all books", func(t *testing.T) {

		data := []byte("{\"query\":\"{book(id:\"d3f444af-1101-4cb2-82b9-52846bd4bce2\"){title,publication_year,slug,author,pages,id}}\"}")

		request, _ := http.NewRequest(http.MethodPost, "/graphql", bytes.NewBuffer(data))
		request.Header.Set("Content-Type", "appilcation/json")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		var got books.Book
		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Errorf("Unable to parse response from server '%v' into Book, '%v'", response.Body.String(), err)
		}

		uid := uuid.Must(uuid.FromString("de0e4051-54b1-4f37-97f2-619b5b568d7f"))

		want, err := store.BookByID(uid)
		if err != nil {
			t.Errorf("cannot access the store: %+v", err)
		}

		if response.Code != 200 {
			t.Errorf("want 200 got %d", response.Code)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})

	t.Run("can query a book by id", func(t *testing.T) {

		data := []byte(`{"query": "{book(id:"de0e4051-54b1-4f37-97f2-619b5b568d7f"){title,publication_year,slug,author,pages}}"}`)

		request, _ := http.NewRequest(http.MethodPost, "/graphql", bytes.NewBuffer(data))
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		var got books.Book
		err := json.NewDecoder(response.Body).Decode(&got)
		if err != nil {
			t.Errorf("Unable to parse response from server '%v' into Book, '%v'", response.Body.String(), err)
		}

		uid := uuid.Must(uuid.FromString("de0e4051-54b1-4f37-97f2-619b5b568d7f"))

		want, err := store.BookByID(uid)
		if err != nil {
			t.Errorf("cannot access the store: %+v", err)
		}

		if response.Code != 200 {
			t.Errorf("want 200 got %d", response.Code)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v want %v", got, want)
		}
	})
}
