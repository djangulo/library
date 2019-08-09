package main

import (
	"fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/config"
	"github.com/gofrs/uuid"
	// "github.com/go-chi/chi"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

var port string

func init() {
	const (
		defaultPort = "9000"
		portUsage   = "port to serve the app on, default '" + defaultPort + "'"
	)
	flag.StringVar(&port, "port", defaultPort, portUsage)
	flag.StringVar(&port, "p", defaultPort, portUsage+" (shorthand)")
}

func main() {
	flag.Parse()
	fmt.Println("Listening at port " + port)
	fmt.Println("Books is a-running")
	bookspath := filepath.Join(config.RootDir, "books", "testdata", "fakeBooks.json")
	pagesPath := filepath.Join(config.RootDir, "books", "testdata", "fakePages.json")

	datBooks, _ := ioutil.ReadFile(bookspath)
	datPages, _ := ioutil.ReadFile(pagesPath)
	var boooks []books.Book
	var pages []books.Page
	json.Unmarshal(datBooks, &boooks)
	json.Unmarshal(datPages, &pages)
	store := books.NewStubStore(boooks, pages, map[uuid.UUID]int{}, map[uuid.UUID]int{})

	server, err := books.NewBookServer(store, books.BookSchema)
	if err != nil {
		log.Fatalf("could not create server %v", err)
	}
	if err = http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}
