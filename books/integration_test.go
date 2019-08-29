package books_test

import (
	"fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/books/testutils"
	config "github.com/djangulo/library/config/books"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
		err := books.SeedSQLite(cache, tAuthors, tBooks, tPages)
		if err != nil {
			log.Fatalf("failed to seed sqlite cache: %v\n", err)
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

	marks := []struct {
		name  string
		query string
	}{
		{"book by id", fmt.Sprintf(`{"query":"{book(id:\"%s\"){%s}}"}`, tBooks[0].ID.String(), bookFields)},
		{"book by slug", fmt.Sprintf(`{"query":"{book(slug:\"%s\"){%s}}"}`, tBooks[0].Slug, bookFields)},
		{"page by id", fmt.Sprintf(`{"query":"{page(id:\"%s\"){%s}}"}`, tPages[0].ID.String(), pageFields)},
		{
			"page by book+number",
			fmt.Sprintf(`{"query":"{page(book_id:\"%s\",number:%d){%s}}"}`, tPages[0].BookID.String(), tPages[0].PageNumber, pageFields),
		},
		{"author by id", fmt.Sprintf(`{"query":"{author(id:\"%s\"){%s}}"}`, tAuthors[0].ID.String(), authorFields)},
		{"author by name w name", fmt.Sprintf(`{"query":"{author(name:\"%s\"){%s}}"}`, tAuthors[0].Name, authorFields)},
		{"author by name w slug", fmt.Sprintf(`{"query":"{author(name:\"%s\"){%s}}"}`, tAuthors[0].Slug, authorFields)},
		{"allBook", fmt.Sprintf(`{"query":"{allBook{%s}}"}`, bookFields)},
		{"allBook w limit", fmt.Sprintf(`{"query":"{allBook(limit:3){%s}}"}`, bookFields)},
		{"allBook w offset", fmt.Sprintf(`{"query":"{allBook(offset:3){%s}}"}`, bookFields)},
		{"allBook by author", fmt.Sprintf(`{"query":"{allBook(author:\"%s\"){%s}}"}`, tAuthors[0].Name, bookFields)},
		{"allPage 1", fmt.Sprintf(`{"query":"{allPage{%s}}"}`, pageFields)},
		{"allPage 2", fmt.Sprintf(`{"query":"{allPage(offset:1000){%s}}"}`, pageFields)},
		{"allPage 3", fmt.Sprintf(`{"query":"{allPage(offset:2000){%s}}"}`, pageFields)},
		{"allPage 4", fmt.Sprintf(`{"query":"{allPage(offset:3000){%s}}"}`, pageFields)},
		{"allPage 5", fmt.Sprintf(`{"query":"{allPage(offset:4000){%s}}"}`, pageFields)},
		{"allPage limit override", fmt.Sprintf(`{"query":"{allPage(limit:1050){%s}}"}`, pageFields)},
		{"allPage w limit", fmt.Sprintf(`{"query":"{allPage(limit:3){%s}}"}`, pageFields)},
		{"allPage w offset", fmt.Sprintf(`{"query":"{allPage(offset:3){%s}}"}`, pageFields)},
		{"allAuthor", fmt.Sprintf(`{"query":"{allAuthor{%s}}"}`, authorFields)},
		{"allAuthor w limit", fmt.Sprintf(`{"query":"{allAuthor(limit:3){%s}}"}`, authorFields)},
		{"allAuthor w offset", fmt.Sprintf(`{"query":"{allAuthor(offset:3){%s}}"}`, authorFields)},
	}

	b.Run("with cache", func(b *testing.B) {
		cache, dropCache := books.NewInMemoryStore("integration_benchmark", true)
		defer dropCache()
		// prepopulate cache queries
		err := books.SeedSQLite(cache, tAuthors, tBooks, tPages)
		if err != nil {
			log.Fatalf("failed to seed sqlite cache: %v\n", err)
		}
		server, _ := books.NewBookServer(store, cache, middlewares, true)

		for _, c := range marks {
			b.Run(c.name, func(b *testing.B) {
				stream := []byte(c.query)
				for i := 0; i < b.N; i++ {
					request := testutils.NewJSONPostRequest("/graphql", stream)
					response := httptest.NewRecorder()
					server.ServeHTTP(response, request)
				}

			})
		}
	})
	b.Run("no cache", func(b *testing.B) {
		cache, dropCache := books.NewInMemoryStore("unused", false)
		defer dropCache()
		server, _ := books.NewBookServer(store, cache, middlewares, true)

		for _, c := range marks {
			b.Run(c.name, func(b *testing.B) {
				stream := []byte(c.query)
				for i := 0; i < b.N; i++ {
					request := testutils.NewJSONPostRequest("/graphql", stream)
					response := httptest.NewRecorder()
					server.ServeHTTP(response, request)
				}

			})
		}
	})

}
