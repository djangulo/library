package main

import (
	"fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/config"
	"github.com/go-chi/chi/middleware"
	// "github.com/gofrs/uuid"
	// "github.com/go-chi/chi"
	// "encoding/json"
	"flag"
	// "io/ioutil"
	"log"
	"net/http"
	// "path/filepath"
)

var port string
var devMode bool

func init() {
	const (
		defaultPort = "9000"
		portUsage   = "port to serve the app on, default '" + defaultPort + "'"
		devDefault  = false
		devUsage    = "enable development mode (/___graphql), default false"
	)
	flag.StringVar(&port, "port", defaultPort, portUsage)
	flag.StringVar(&port, "p", defaultPort, portUsage+" (shorthand)")
	flag.BoolVar(&devMode, "dev", devDefault, devUsage)
}

var middlewares = []func(http.Handler) http.Handler{
	middleware.RequestID,
	middleware.RealIP,
	middleware.Logger,
	middleware.Recoverer,
}

func main() {
	flag.Parse()
	fmt.Println("Listening at port " + port)
	fmt.Println("Books is a-running")
	// bookspath := "/home/djangulo/go/src/github.com/djangulo/library/books/testdata/fakeBooks.json"
	// pagesPath := "/home/djangulo/go/src/github.com/djangulo/library/books/testdata/fakePages.json"

	// datBooks, _ := ioutil.ReadFile(bookspath)
	// datPages, _ := ioutil.ReadFile(pagesPath)
	// var boooks []books.Book
	// var pages []books.Page
	// json.Unmarshal(datBooks, &boooks)
	// json.Unmarshal(datPages, &pages)

	cnf := config.Get()

	store, removeStore := books.NewSQLStore(cnf.Database)
	defer removeStore()

	server, err := books.NewBookServer(store, middlewares, devMode)
	if err != nil {
		log.Fatalf("could not create server %v", err)
	}
	if err = http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}
