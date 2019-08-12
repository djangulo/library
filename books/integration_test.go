package books_test

import (
// "fmt"
// // "github.com/djangulo/library/books"
// // "github.com/djangulo/library/config"
// // "github.com/gofrs/uuid"
// // "github.com/jmoiron/sqlx"
// // "net/http"
// // "net/http/httptest"
// "os"
// "testing"
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
