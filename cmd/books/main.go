package main

import (
	"flag"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/config"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
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
	cnf := config.Get()
	flag.Parse()
	log.Println("Books listening on port " + port)
	log.Println("Books is a-running")

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
