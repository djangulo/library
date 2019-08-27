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
	pageFields   = "id,book_id,page_number,body"
)

func TestGetRoot(t *testing.T) {
	store := testutils.NewStubStore(true)
	cache := testutils.NewStubStore(true)
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
		{"allBook by author", fmt.Sprintf(`{"query":"{allBook(author:\"%s\"){%s}}"}`, "Stephen King", bookFields), "book", 3},
		{"allPage", fmt.Sprintf(`{"query":"{allPage{%s}}"}`, pageFields), "page", len(tPages)}, // default limit
		{"allPage w limit", fmt.Sprintf(`{"query":"{allPage(limit:3){%s}}"}`, pageFields), "page", 3},
		{"allPage w offset", fmt.Sprintf(`{"query":"{allPage(offset:3){%s}}"}`, pageFields), "page", tPages[3]},
		{"allAuthor", fmt.Sprintf(`{"query":"{allAuthor{%s}}"}`, authorFields), "author", len(tAuthors)},
		{"allAuthor w limit", fmt.Sprintf(`{"query":"{allAuthor(limit:3){%s}}"}`, authorFields), "author", 3},
		{"allAuthor w offset", fmt.Sprintf(`{"query":"{allAuthor(offset:3){%s}}"}`, authorFields), "author", tAuthors[3]},
	}

	t.Run("with cache", func(t *testing.T) {
		store := testutils.NewStubStore(true)
		cache := testutils.NewStubStore(true)
		server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

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
		store := testutils.NewStubStore(true)
		cache := testutils.NewStubStore(false)
		server, _ := books.NewBookServer(store, cache, testutils.DummyMiddlewares, true)

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

}
