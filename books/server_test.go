package books_test

import (
	"fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/books/testutils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	bookFields   = "id,title,slug,author_id,file,source,publication_year,page_count"
	authorFields = "id,name,slug"
	pageFields   = "id,book_id,page_number,body,created_at,updated_at, deleted_at"
)

func TestGetRoot(t *testing.T) {
	store := testutils.NewStubStore(testutils.Available, testutils.PrepopulateStore)
	cache := testutils.NewStubCache(nil, true)
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

func TestGraphQLQueries(t *testing.T) {

	tBooks := testutils.TestBookData()
	tPages := testutils.TestPageData()
	tAuthors := testutils.TestAuthorData()

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
		{"allBook by author", fmt.Sprintf(`{"query":"{allBook(author:\"%s\"){%s}}"}`, "Stephen King", bookFields), 3},
		{"allPage", fmt.Sprintf(`{"query":"{allPage{%s}}"}`, pageFields), len(tPages)}, // default limit
		{"allPage w limit", fmt.Sprintf(`{"query":"{allPage(limit:3){%s}}"}`, pageFields), 3},
		{"allPage w offset", fmt.Sprintf(`{"query":"{allPage(offset:3){%s}}"}`, pageFields), tPages[3]},
		{"allAuthor", fmt.Sprintf(`{"query":"{allAuthor{%s}}"}`, authorFields), len(tAuthors)},
		{"allAuthor w limit", fmt.Sprintf(`{"query":"{allAuthor(limit:3){%s}}"}`, authorFields), 3},
		{"allAuthor w offset", fmt.Sprintf(`{"query":"{allAuthor(offset:3){%s}}"}`, authorFields), tAuthors[3]},
	}

	t.Run("with prepopulated cache", func(t *testing.T) {
		// Cache is  available and prepopulated:
		// 		assert store is not called (cache takes over)
		// 		assert cache is called
		// 		assert no insertions are run
		// 		assert correctness

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				store := testutils.NewStubStore(testutils.Available, testutils.PrepopulateStore)
				cache := testutils.NewStubCache(nil, true)

				server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

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
						testutils.AssertBookStoreCalls(t, store, "list", 0)
						testutils.AssertBookCacheCalls(t, cache, "list", 1)
						testutils.AssertStoreBookInsertCalls(t, store, "bulk", 0)
					} else {
						got := testutils.GetBookFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertBookStoreCalls(t, store, want.ID.String(), 0)
						testutils.AssertBookCacheCalls(t, cache, want.ID.String(), 1)
						testutils.AssertStoreBookInsertCalls(t, store, want.ID.String(), 0)
					}
				case books.Page:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
						testutils.AssertPageStoreCalls(t, store, "list", 0)
						testutils.AssertPageCacheCalls(t, cache, "list", 1)
					} else {
						got := testutils.GetPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertPageStoreCalls(t, store, want.ID.String(), 0)
						testutils.AssertPageCacheCalls(t, cache, want.ID.String(), 1)
					}
				case books.Author:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
						testutils.AssertAuthorStoreCalls(t, store, "list", 0)
						testutils.AssertAuthorCacheCalls(t, cache, "list", 1)
					} else {
						got := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertAuthorStoreCalls(t, store, want.ID.String(), 0)
						testutils.AssertAuthorCacheCalls(t, cache, want.ID.String(), 1)
					}
				case int:
					if strings.Contains(strings.ToLower(c.name), "book") {
						if got := testutils.GetAllBookFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertBookStoreCalls(t, store, "list", 0)
							testutils.AssertBookCacheCalls(t, cache, "list", 1)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "page") {
						if got := testutils.GetAllPageFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertPageStoreCalls(t, store, "list", 0)
							testutils.AssertPageCacheCalls(t, cache, "list", 1)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "Author") {
						if got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertAuthorStoreCalls(t, store, "list", 0)
							testutils.AssertAuthorCacheCalls(t, cache, "list", 1)
						}
					}
				}

			})
		}
	})

	t.Run("with non-prepopulated cache", func(t *testing.T) {
		// Cache is available but unpopulated:
		// 		assert store is called
		// 		assert cache is NOT called
		// 		assert cache insertions are run
		// 		assert correctness

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				store := testutils.NewStubStore(testutils.Available, testutils.PrepopulateStore)
				cache := testutils.NewStubCache(nil, testutils.NOPrepopulateStore)

				server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

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
						testutils.AssertBookStoreCalls(t, store, "list", 1)
						testutils.AssertBookCacheCalls(t, cache, "list", 0)
					} else {
						got := testutils.GetBookFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertBookStoreCalls(t, store, want.ID.String(), 1)
						testutils.AssertBookCacheCalls(t, cache, want.ID.String(), 0)
					}
				case books.Page:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
						testutils.AssertPageStoreCalls(t, store, "list", 1)
						testutils.AssertPageCacheCalls(t, cache, "list", 0)
					} else {
						got := testutils.GetPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertPageStoreCalls(t, store, want.ID.String(), 1)
						testutils.AssertPageCacheCalls(t, cache, want.ID.String(), 0)
					}
				case books.Author:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
						testutils.AssertAuthorStoreCalls(t, store, "list", 1)
						testutils.AssertAuthorCacheCalls(t, cache, "list", 0)
					} else {
						got := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertAuthorStoreCalls(t, store, want.ID.String(), 1)
						testutils.AssertAuthorCacheCalls(t, cache, want.ID.String(), 0)
					}
				case int:
					if strings.Contains(strings.ToLower(c.name), "book") {
						if got := testutils.GetAllBookFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertBookStoreCalls(t, store, "list", 1)
							testutils.AssertBookCacheCalls(t, cache, "list", 0)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "page") {
						if got := testutils.GetAllPageFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertPageStoreCalls(t, store, "list", 1)
							testutils.AssertPageCacheCalls(t, cache, "list", 0)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "Author") {
						if got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertAuthorStoreCalls(t, store, "list", 1)
							testutils.AssertAuthorCacheCalls(t, cache, "list", 0)
						}
					}
				}

			})
		}
	})

	t.Run("no cache", func(t *testing.T) {
		// Cache is unavailable:
		// 		assert store is called
		// 		assert cache is NOT called
		// 		assert NO cache insertions are run
		// 		assert correctness

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				store := testutils.NewStubStore(testutils.Available, testutils.PrepopulateStore)
				cache := testutils.NewStubCache(books.ErrCacheUnavailable, testutils.NOPrepopulateStore)

				server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

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
						testutils.AssertBookStoreCalls(t, store, "list", 1)
						testutils.AssertBookCacheCalls(t, cache, "list", 0)
					} else {
						got := testutils.GetBookFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertBookStoreCalls(t, store, want.ID.String(), 1)
						testutils.AssertBookCacheCalls(t, cache, want.ID.String(), 0)
					}
				case books.Page:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
						testutils.AssertPageStoreCalls(t, store, "list", 1)
						testutils.AssertPageCacheCalls(t, cache, "list", 0)
					} else {
						got := testutils.GetPageFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertPageStoreCalls(t, store, want.ID.String(), 1)
						testutils.AssertPageCacheCalls(t, cache, want.ID.String(), 0)
					}
				case books.Author:
					if strings.Contains(strings.ToLower(c.name), "offset") {
						got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got[0].ID, want.ID)
						testutils.AssertAuthorStoreCalls(t, store, "list", 1)
						testutils.AssertAuthorCacheCalls(t, cache, "list", 0)
					} else {
						got := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
						testutils.AssertUUIDsEqual(t, got.ID, want.ID)
						testutils.AssertAuthorStoreCalls(t, store, want.ID.String(), 1)
						testutils.AssertAuthorCacheCalls(t, cache, want.ID.String(), 0)
					}
				case int:
					if strings.Contains(strings.ToLower(c.name), "book") {
						if got := testutils.GetAllBookFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertBookStoreCalls(t, store, "list", 1)
							testutils.AssertBookCacheCalls(t, cache, "list", 0)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "page") {
						if got := testutils.GetAllPageFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertPageStoreCalls(t, store, "list", 1)
							testutils.AssertPageCacheCalls(t, cache, "list", 0)
						}
					}
					if strings.Contains(strings.ToLower(c.name), "Author") {
						if got := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body); len(got) > 0 {
							testutils.AssertIntsEqual(t, len(got), want)
							testutils.AssertAuthorStoreCalls(t, store, "list", 1)
							testutils.AssertAuthorCacheCalls(t, cache, "list", 0)
						}
					}
				}

			})
		}
	})

}
