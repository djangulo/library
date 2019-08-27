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

func TestBookQueries(t *testing.T) {

	store, remove := testutils.NewTestSQLStore(cnf, "test")
	defer remove()
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "test", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}

	t.Run("with cache", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, true)
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
			want := len(testBooks)

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, true)
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
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, true)
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 1) {
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

	t.Run("without cache", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, false)
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
			want := len(testBooks)

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, false)
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
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, false)
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 1) {
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

func TestPageQueries(t *testing.T) {

	store, remove := testutils.NewTestSQLStore(cnf, "test")
	defer remove()
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "test", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}

	t.Run("with cache", func(t *testing.T) {
		/*

			Test omitted due to large number of sample pages.

		*/
		// t.Run("can query all pages", func(t *testing.T) {
		// 	server, _, _, testPages := testServer(store, true)
		// 	jsonStream := []byte(`
		// 	{
		// 		"query": "{
		// 			allPage(limit) {
		// 				id,
		// 				page_number,
		// 				book_id,
		// 				body,
		// 			}
		// 		}"
		// 	}`)
		// 	jsonStream = testutils.FlattenJSON(jsonStream)
		// 	request := testutils.NewJSONPostRequest("/graphql", jsonStream)
		// 	response := httptest.NewRecorder()
		// 	server.ServeHTTP(response, request)

		// 	pages := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
		// 	got := len(pages)
		// 	want := len(testBooks)

		// 	testutils.AssertStatus(t, response, http.StatusOK)
		// 	testutils.AssertEqual(t, got, want)
		// })

		t.Run("can query all pages with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, true)
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
		})

		t.Run("can query all pages with an offset", func(t *testing.T) {
			server, _, _, testPages := testServer(store, true)
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
			if len(pages) == 0 {
				t.Error("expected a result but got none")
			}
			got := pages[0].PageNumber
			want := testPages[5].PageNumber

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query a page by id", func(t *testing.T) {
			server, _, _, testPages := testServer(store, true)
			str := fmt.Sprintf(`
			{
				"query": "{
					page(id:\"%s\") {
						id,
						page_number,
						book_id,
						body
					}
				}"
			}
			`, testPages[0].ID.String())
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := testPages[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query a page by book id and page number", func(t *testing.T) {
			server, _, _, testPages := testServer(store, true)
			str := fmt.Sprintf(`
			{
				"query": "{
					page(
							book_id:\"%s\",
							number: %d
						) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}
			`, testPages[0].BookID.String(), testPages[0].PageNumber)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := testPages[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
	})

	t.Run("without cache", func(t *testing.T) {
		/*

			Test omitted due to large number of sample pages.

		*/
		// t.Run("can query all pages", func(t *testing.T) {
		// 	server, _, _, testPages := testServer(store, false)
		// 	jsonStream := []byte(`
		// 	{
		// 		"query": "{
		// 			allPage(limit) {
		// 				id,
		// 				page_number,
		// 				book_id,
		// 				body,
		// 			}
		// 		}"
		// 	}`)
		// 	jsonStream = testutils.FlattenJSON(jsonStream)
		// 	request := testutils.NewJSONPostRequest("/graphql", jsonStream)
		// 	response := httptest.NewRecorder()
		// 	server.ServeHTTP(response, request)

		// 	pages := testutils.GetAllBookFromGraphQLResponse(t, response.Body)
		// 	got := len(pages)
		// 	want := len(testBooks)

		// 	testutils.AssertStatus(t, response, http.StatusOK)
		// 	testutils.AssertEqual(t, got, want)
		// })

		t.Run("can query all pages with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, false)
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
		})

		t.Run("can query all pages with an offset", func(t *testing.T) {
			server, _, _, testPages := testServer(store, false)
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
			if len(pages) == 0 {
				t.Error("expected a result but got none")
			}
			got := pages[0].PageNumber
			want := testPages[5].PageNumber

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query a page by id", func(t *testing.T) {
			server, _, _, testPages := testServer(store, false)
			str := fmt.Sprintf(`
			{
				"query": "{
					page(id:\"%s\") {
						id,
						page_number,
						book_id,
						body
					}
				}"
			}
			`, testPages[0].ID.String())
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := testPages[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query a page by book id and page number", func(t *testing.T) {
			server, _, _, testPages := testServer(store, false)
			str := fmt.Sprintf(`
			{
				"query": "{
					page(
						book_id:\"%s\",
							number: %d
						) {
						id,
						page_number,
						book_id,
						body,
					}
				}"
			}
			`, testPages[0].BookID.String(), testPages[0].PageNumber)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			page := testutils.GetPageFromGraphQLResponse(t, response.Body)
			got := page.ID
			want := testPages[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
	})

}

func TestAuthorQueries(t *testing.T) {

	store, remove := testutils.NewTestSQLStore(cnf, "test")
	defer remove()
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "test", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}

	t.Run("with cache", func(t *testing.T) {
		t.Run("can query all authors", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, true)
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
			want := len(testAuthors)

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all authors with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, true)
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
		})

		t.Run("can query all authors with an offset", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, true)
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
			if len(authors) == 0 {
				t.Fatal("expected a result but got none")
			}
			got := authors[0].Name
			want := testAuthors[2].Name

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query an author by id", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, true)
			str := fmt.Sprintf(`
			{
				"query": "{
					author(id:\"%s\") {
						id,
						name,
						slug,
					}
				}"
			}
			`, testAuthors[0].ID.String())
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := testAuthors[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query an author by name using name", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, true)
			str := fmt.Sprintf(`
			{
				"query": "{
					author(name:\"%s\") {
						id,
						slug,
						name
					}
				}"
			}
			`, testAuthors[0].Name)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := testAuthors[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query an author by name using slug", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, true)
			str := fmt.Sprintf(`
			{
				"query": "{
					author(name:\"%s\") {
						id,
						slug,
						name
					}
				}"
			}
			`, testAuthors[0].Slug)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := testAuthors[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
	})

	t.Run("without cache", func(t *testing.T) {
		t.Run("can query all authors", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, false)
			jsonStream := []byte(`
			{
				"query": "{
					allAuthor {
						id,
						name,
						slug
					}
				}"
			}`)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			authors := testutils.GetAllAuthorFromGraphQLResponse(t, response.Body)
			got := len(authors)
			want := len(testAuthors)

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query all authors with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, false)
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
		})

		t.Run("can query all authors with an offset", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, false)
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
			if len(authors) == 0 {
				t.Fatal("expected a result but got none")
			}
			got := authors[0].Name
			want := testAuthors[2].Name

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertEqual(t, got, want)
		})

		t.Run("can query an author by id", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, false)
			str := fmt.Sprintf(`
			{
				"query": "{
					author(id:\"%s\") {
						id,
						name,
						slug,
					}
				}"
			}
			`, testAuthors[0].ID.String())
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)
			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := testAuthors[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query an author by name using name", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, false)
			str := fmt.Sprintf(`
			{
				"query": "{
					author(name:\"%s\") {
						id,
						slug,
						name
					}
				}"
			}
			`, testAuthors[0].Name)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := testAuthors[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
		t.Run("can query an author by name using slug", func(t *testing.T) {
			server, testAuthors, _, _ := testServer(store, false)
			str := fmt.Sprintf(`
			{
				"query": "{
					author(name:\"%s\") {
						id,
						slug,
						name
					}
				}"
			}
			`, testAuthors[0].Slug)
			jsonStream := []byte(str)
			jsonStream = testutils.FlattenJSON(jsonStream)

			request := testutils.NewJSONPostRequest("/graphql", jsonStream)
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)

			author := testutils.GetAuthorFromGraphQLResponse(t, response.Body)
			got := author.ID
			want := testAuthors[0].ID

			testutils.AssertStatus(t, response, http.StatusOK)
			testutils.AssertUUIDsEqual(t, got, want)
		})
	})

}

func BenchmarkServerWithNoCache(b *testing.B) {
	store, remove := testutils.NewTestSQLStore(cnf, "bench")
	defer remove()
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "bench", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}
	cache := new(books.RedisCache)
	// cache, err := books.NewRedisCache(cnf.Cache["test"])
	// if err != nil {
	// 	log.Fatal(err)
	// }
	cache.Available = false

	// testPages, _ := books.PageSeedData(cnf)
	testAuthors, _ := books.AuthorSeedData(cnf)
	testBooks, _ := books.BookSeedData(cnf)

	server, _ := books.NewBookServer(store, cache, middlewares, true)

	offset := 1000
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
		for _, book := range testBooks[:2] {
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
		for _, author := range testAuthors[:2] {
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
}

func BenchmarkServerWithRedigoCache(b *testing.B) {
	store, remove := testutils.NewTestSQLStore(cnf, "bench")
	defer remove()
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "bench", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}
	cache, err := books.NewRedigoCache(cnf.Cache["bench"])
	if err != nil {
		log.Fatal(err)
	}

	// prepopulate cache queries
	offset := 1000
	testPages, _ := books.PageSeedData(cnf)
	testAuthors, _ := books.AuthorSeedData(cnf)
	testBooks, _ := books.BookSeedData(cnf)
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("Pages(%d,%d)", offset, i*offset)
		cache.SavePageQuery(key, testPages[(i*offset):(i*offset+offset)])
	}
	for i := 0; i < len(testPages); i++ {
		cache.InsertPage(testPages[i])
	}
	cache.SaveBookQuery("Books(1000,0)", testBooks)
	for i := 0; i < len(testBooks[:2]); i++ {
		cache.InsertBook(testBooks[i])
	}
	cache.SaveAuthorQuery("Authors(1000,0)", testAuthors)
	for i := 0; i < len(testAuthors[:2]); i++ {
		cache.InsertAuthor(testAuthors[i])
	}

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
		for _, book := range testBooks[:2] {
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
		for _, author := range testAuthors[:2] {
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
}

func BenchmarkServerWithGoRedisCache(b *testing.B) {
	store, remove := testutils.NewTestSQLStore(cnf, "bench")
	defer remove()
	books.AcquireGutenberg(cnf, false)
	err := books.SaveJSON(cnf, false)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "bench", false)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}
	cache, err := books.NewRedisCache(cnf.Cache["bench"])
	if err != nil {
		log.Fatal(err)
	}

	// prepopulate cache queries
	offset := 1000
	testPages, _ := books.PageSeedData(cnf)
	testAuthors, _ := books.AuthorSeedData(cnf)
	testBooks, _ := books.BookSeedData(cnf)
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("Pages(%d,%d)", offset, i*offset)
		cache.SavePageQuery(key, testPages[(i*offset):(i*offset+offset)])
	}
	for i := 0; i < len(testPages); i++ {
		cache.InsertPage(testPages[i])
	}
	cache.SaveBookQuery("Books(1000,0)", testBooks)
	for i := 0; i < len(testBooks[:2]); i++ {
		cache.InsertBook(testBooks[i])
	}
	cache.SaveAuthorQuery("Authors(1000,0)", testAuthors)
	for i := 0; i < len(testAuthors[:2]); i++ {
		cache.InsertAuthor(testAuthors[i])
	}

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
		for _, book := range testBooks[:2] {
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
		for _, author := range testAuthors[:2] {
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
