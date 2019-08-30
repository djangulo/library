package main

import (
	"flag"
	"github.com/djangulo/library/books"
	config "github.com/djangulo/library/config/books"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
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

	books.AcquireGutenberg(cnf, true)
	err := books.SaveJSON(cnf, true)
	if err != nil {
		log.Fatalf("Error creating json files: %v", err)
	}
	err = books.SeedFromGutenberg(cnf, "main", true)
	if err != nil {
		log.Fatalf("Error seeding database: %v", err)
	}
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

	cache, err := books.NewRedisCache(cnf.Cache["main"])
	if err != nil {
		log.Println("could not create cache %v, initalizing as unavailable", err)
	}

	server, err := books.NewBookServer(store, cache, middlewares, devMode)
	if err != nil {
		log.Fatalf("could not create server %v", err)
	}
	if err = http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}
