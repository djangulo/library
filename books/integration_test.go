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
	testBooks, _ := store.Books(2, 0)
	testAuthors, _ := store.Authors(2, 0)
	testPages, _ := store.Pages(2, 0)
	return server, testAuthors, testBooks, testPages
}
func TestBookQueriesWithoutCache(t *testing.T) {

	store, remove := testutils.NewTestSQLStore(cnf)
	defer remove()
	books.AcquireGutenberg(cnf)
	err := books.SaveJSON(cnf)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "test")
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}
	t.Run("tests without cache", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			server, _, _, _ := testServer(store, false)
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
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, false)
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
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, false)
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 1) {
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
						title,
						pages{id}
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
						title,
						publication_year,
						slug,
						author,
						id
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
						title,
						publication_year,
						slug,
						author,
						id
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

	t.Run("tests with a cache", func(t *testing.T) {
		t.Run("can query all books", func(t *testing.T) {
			server, _, _, _ := testServer(store, true)
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
		})

		t.Run("can query all books with a limit", func(t *testing.T) {
			server, _, _, _ := testServer(store, true)
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
		})

		t.Run("can query all books with an offset", func(t *testing.T) {
			server, _, testBooks, _ := testServer(store, true)
			jsonStream := []byte(`{
				"query": "{
					allBook(offset: 1) {
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
						title,
						pages{id}
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
						title,
						publication_year,
						slug,
						author,
						id
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
						title,
						publication_year,
						slug,
						author,
						id
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

// TestPostgreSQLPlayerStore integration test
// func TestSQLStoreIntegration(t *testing.T) {
// 	t.Run("get score", func(t *testing.T) {
// 		store, remove := books.NewTestSQLStore(
// 			config.DatabaseHost,
// 			config.DatabasePort,
// 			config.DatabaseUser,
// 			config.DatabaseName,
// 			config.DatabasePassword,
// 		)
// 		defer remove()
// 		server, _ := books.NewBookServer(store, books.BookSchema)

// 		server.ServeHTTP(httptest.NewRecorder(), books.NewPostWinRequest(player))

// 		response := httptest.NewRecorder()
// 		server.ServeHTTP(response, books.NewGetScoreRequest(player))

// 		books.AssertStatus(t, response.Code, http.StatusOK)
// 		books.AssertResponseBody(t, response.Body.String(), "1")
// 	})

// 	t.Run("get league", func(t *testing.T) {
// 		store, remove := newTestPostgreSQLPlayerStore(
// 			config.DatabaseHost,
// 			config.DatabasePort,
// 			config.DatabaseUser,
// 			config.DatabaseName,
// 			config.DatabasePassword,
// 		)
// 		defer remove()

// 		server, _ := books.NewBookServer(store, books.DummyGame)
// 		server.ServeHTTP(httptest.NewRecorder(), books.NewPostWinRequest(player))
// 		server.ServeHTTP(httptest.NewRecorder(), books.NewPostWinRequest(player))
// 		server.ServeHTTP(httptest.NewRecorder(), books.NewPostWinRequest(player))

// 		response := httptest.NewRecorder()

// 		server.ServeHTTP(response, books.NewLeagueRequest())
// 		books.AssertStatus(t, response.Code, http.StatusOK)

// 		got := books.GetLeagueFromResponse(t, response.Body)
// 		want := books.League{
// 			{Name: "Pepper", Wins: 3},
// 		}
// 		books.AssertLeague(t, got, want)
// 	})

// }

// func newTransactionWalledPlayerStore(t *testing.T, host, port, user, dbname, pass string) (*transactionWalledPlayerStore, func()) {
// 	connStr := fmt.Sprintf(
// 		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
// 		user,
// 		pass,
// 		host,
// 		port,
// 		dbname,
// 	)
// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		log.Fatalf("failed to connect database %v", err)
// 	}
// 	tx, err := db.Begin()
// 	if err != nil {
// 		t.Fatalf("error creating transaction: %v", err)
// 	}
// 	// Create savepoint
// 	_, err = tx.Exec(`SAVEPOINT test_savepoint;`)
// 	if err != nil {
// 		log.Printf("savepoint error: %v", err)
// 	}

// 	_, errCreate := tx.Exec(`
// 	CREATE TABLE IF NOT EXISTS players (
// 		id		serial		PRIMARY KEY,
// 		name	varchar(80)	NOT NULL UNIQUE,
// 		wins	int			DEFAULT 0
// 	);
// 	`)
// 	if errCreate != nil {
// 		log.Fatalf("failed to create table %v", errCreate)
// 	}

// 	removeDatabase := func() {
// 		_, err = tx.Exec(`ROLLBACK TO SAVEPOINT test_savepoint;`)
// 		if err != nil {
// 			log.Printf("rollback error: %v", err)
// 		}
// 		// Release savepoint
// 		_, err = tx.Exec(`RELEASE SAVEPOINT test_savepoint;`)
// 		if err != nil {
// 			log.Printf("release error: %v", err)
// 		}
// 		// Commit empty transaction
// 		tx.Rollback() // tx.Commit() had the same outcome
// 		db.Close()
// 	}

// 	return &transactionWalledPlayerStore{tx}, removeDatabase
// }

// type transactionWalledPlayerStore struct {
// 	DB *sql.Tx
// }

// func (s *transactionWalledPlayerStore) GetPlayerScore(name string) int {
// 	var wins int
// 	err := s.DB.QueryRow(`SELECT wins FROM players WHERE name = $1;`, name).Scan(&wins)
// 	if err != nil {
// 		log.Printf("error: %v", err)
// 		return 0
// 	}
// 	return wins
// }
// func (s *transactionWalledPlayerStore) RecordWin(name string) {
// 	var userID int
// 	err := s.DB.QueryRow(`SELECT id FROM players WHERE name = $1;`, name).Scan(&userID)
// 	if err != nil { // likely does not exist
// 		log.Printf("error: %v, inserting", err)
// 		s.DB.Exec(`INSERT INTO players(name, wins) VALUES($1, 1);`, name)
// 		return
// 	}
// 	s.DB.Exec(`UPDATE players SET wins = wins + 1 WHERE name = $1`, name)
// }

// func (s *transactionWalledPlayerStore) GetLeague() books.League {
// 	results, err := s.DB.Query(`
// 	SELECT name, wins FROM players ORDER BY	wins DESC,name ASC;`)
// 	if err != nil {
// 		fmt.Printf("error: %v", err)
// 		return nil
// 	}
// 	var players books.League
// 	for results.Next() {
// 		var player books.Player
// 		err := results.Scan(&player.Name, &player.Wins)
// 		if err != nil {
// 			fmt.Printf("error: %v", err)
// 		}
// 		players = append(players, player)
// 	}
// 	return players
// }

// func savepointServer(
// 	t *testing.T,
// 	store *books.PostgreSQLPlayerStore,
// ) (*books.PlayerServer, func()) {

// 	tx, err := store.DB.Begin()
// 	if err != nil {
// 		t.Fatalf("error: %v", err)
// 	}
// 	// Create savepoint
// 	_, err = tx.Exec(`SAVEPOINT test_savepoint;`)
// 	if err != nil {
// 		log.Printf("savepoint error: %v", err)
// 	}

// 	removeServer := func() {
// 		_, err = tx.Exec(`ROLLBACK TO SAVEPOINT test_savepoint;`)
// 		if err != nil {
// 			log.Printf("rollback error: %v", err)
// 		}
// 		// Release savepoint
// 		_, err = tx.Exec(`RELEASE SAVEPOINT test_savepoint;`)
// 		if err != nil {
// 			log.Printf("release error: %v", err)
// 		}
// 		// Commit empty transaction
// 		tx.Rollback() // tx.Commit() had the same outcome

// 	}
// 	return server, removeServer
// }

// TestPostgreSQLPlayerStore
// type testPostgreSQLPlayerStore struct {
// 	DB *sql.DB
// }

// func (s *testPostgreSQLPlayerStore) GetPlayerScore(name string) int {
// 	var wins int
// 	err := s.DB.QueryRow(`SELECT wins FROM players WHERE name = $1;`, name).Scan(&wins)
// 	if err != nil {
// 		log.Printf("error: %v", err)
// 		return 0
// 	}
// 	return wins
// }
// func (s *testPostgreSQLPlayerStore) RecordWin(name string) {
// 	var userID int
// 	err := s.DB.QueryRow(`SELECT id FROM players WHERE name = $1;`, name).Scan(&userID)
// 	if err != nil { // likely does not exist
// 		log.Printf("error: %v", err)
// 		s.DB.Exec(`
// 			INSERT INTO
// 				players(name, wins)
// 			VALUES($1, 1);
// 		`, name)
// 		return
// 	}
// 	s.DB.Exec(`UPDATE players SET wins = wins + 1 WHERE name = $1`, name)
// }

// func (s *testPostgreSQLPlayerStore) GetLeague() books.League {
// 	results, err := s.DB.Query(`	SELECT name, wins FROM players ORDER BY	wins DESC, name ASC;`)
// 	if err != nil {
// 		fmt.Printf("error: %v", err)
// 		return nil
// 	}
// 	var players books.League
// 	for results.Next() {
// 		var player books.Player
// 		err := results.Scan(&player.Name, &player.Wins)
// 		if err != nil {
// 			fmt.Printf("error: %v", err)
// 		}
// 		players = append(players, player)
// 	}
// 	return players
// }
