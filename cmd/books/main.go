package main

import (
	"flag"
	"github.com/djangulo/library/books"
	config "github.com/djangulo/library/config/books"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
	"sync"
)

var (
	port    string
	devMode bool
	cnf     = config.Get()
)

func init() {
	const (
		defaultPort = "9000"
		portUsage   = "port to serve the app on, default '" + defaultPort + "'"
		devDefault  = false
		devUsage    = "enable development mode (graphql playground at /___graphql), default false"
	)
	flag.StringVar(&port, "port", defaultPort, portUsage)
	flag.StringVar(&port, "p", defaultPort, portUsage+" (shorthand)")
	flag.BoolVar(&devMode, "dev", devDefault, devUsage)

	var once sync.Once
	migrationsAndSeed := func() {
		books.AcquireGutenberg(cnf)
		books.SaveJSON(cnf)
		// books.SeedFromGutenberg(cnf, "main")
	}
	once.Do(migrationsAndSeed)
}

var middlewares = []func(http.Handler) http.Handler{
	middleware.RequestID,
	middleware.RealIP,
	middleware.Logger,
	middleware.Recoverer,
}

func main() {
	flag.Parse()
	log.Println("Books listening on port " + port)
	log.Println("Books is a-running")

	store, removeStore := books.NewSQLStore(cnf.Database["main"])
	defer removeStore()

	server, err := books.NewBookServer(store, middlewares, devMode)
	if err != nil {
		log.Fatalf("could not create server %v", err)
	}
	if err = http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}
