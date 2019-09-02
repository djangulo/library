package books

import (
	"context"
	"fmt"
	config "github.com/djangulo/library/config/books"
	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

var (
	cnf         = config.Get()
	htmlDirPath = filepath.Join(cnf.Project.Dirs.Static, "books", "html")
)

// BookServer GraphQL Server for book storage
type BookServer struct {
	Store        Storer
	Cache        Cacher
	templatesDir string
	http.Handler
	rootQuery *RootQuery
}

// Storer noqa
type Storer interface {
	AuthorByID(*Author, *uuid.UUID, []string) error
	AuthorBySlug(*Author, string, []string) error
	Authors([]*Author, int, int, *uuid.UUID, *time.Time, []string) error
	BookByID(*Book, *uuid.UUID, []string) error
	BookBySlug(*Book, string, []string) error
	Books([]*Book, int, int, *uuid.UUID, *time.Time, []string) error
	BooksByAuthor([]*Book, string, int, int, *uuid.UUID, *time.Time, []string) error
	BulkInsertAuthors([]*Author) error
	BulkInsertBooks([]*Book) error
	BulkInsertPages([]*Page) error
	InsertAuthor(*Author) error
	InsertBook(*Book) error
	InsertPage(*Page) error
	PageByBookAndNumber(*Page, *uuid.UUID, int, []string) error
	PageByID(*Page, *uuid.UUID, []string) error
	Pages([]*Page, int, int, *uuid.UUID, *time.Time, []string) error
}

// Cacher noqa
type Cacher interface {
	AuthorByID(*Author, *uuid.UUID, []string) error
	AuthorBySlug(*Author, string, []string) error
	AuthorQuery(*[]*Author, string) error
	BookByID(*Book, *uuid.UUID, []string) error
	BookBySlug(*Book, string, []string) error
	BookQuery(*[]*Book, string) error
	InsertAuthor(*Author) error
	InsertBook(*Book) error
	InsertPage(*Page) error
	IsAvailable() error
	PageByBookAndNumber(*Page, *uuid.UUID, int, []string) error
	PageByID(*Page, *uuid.UUID, []string) error
	PageQuery(*[]*Page, string) error
	SaveAuthorQuery(string, []*Author) error
	SaveBookQuery(string, []*Book) error
	SavePageQuery(string, []*Page) error
}

// NewBookServer returns a new server instance
func NewBookServer(
	store Storer,
	cache Cacher,
	middlewares []func(http.Handler) http.Handler,
	developmentMode bool,
) (*BookServer, error) {
	b := new(BookServer)

	b.templatesDir = htmlDirPath

	b.Store = store
	b.Cache = cache
	b.rootQuery = b.NewRootQuery()
	r := chi.NewRouter()

	// middlewares
	for _, m := range middlewares {
		r.Use(m)
	}

	if developmentMode {
		// fmt.Println("Development mode enabled, graphqlendpoint at /___graphql")
		gqlPlayground := filepath.Join(b.templatesDir, "graphqlPlayground.html")
		r.Get("/___graphql", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, gqlPlayground)
		})
	}

	r.Route("/{languageCode}", func(r chi.Router) {
		r.Use(LanguageCtx)
		r.Get("/", b.serveIndex)
	})
	r.Get("/", b.redirectRoot)

	r.Mount("/graphql", b.GraphQLRouter())

	b.Handler = r

	return b, nil
}

type languageKey int

var langKey languageKey = 100000001

// LanguageCtx reads language code from url (/en/) and assigns a system language
func LanguageCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		langCode := chi.URLParam(r, "languageCode")
		ctx := context.WithValue(r.Context(), langKey, langCode)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (b *BookServer) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/en", 302)
}

func (b *BookServer) serveIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langCode := ctx.Value(langKey).(string)
	if langCode == "" {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	templatePath := filepath.Join(b.templatesDir, "index.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("problem opening %s %v", templatePath, err), 400)
	}

	var sampleBook Book
	var sampleAuthor Author
	var samplePage Page
	herman := "herman-melville"
	moby := "moby-dick"
	if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
		b.Cache.AuthorBySlug(&sampleAuthor, herman, []string{"id"})
		b.Cache.BookBySlug(&sampleBook, moby, []string{"id"})
		b.Cache.PageByBookAndNumber(&samplePage, &sampleBook.ID, 1, []string{"id"})
	}
	if sampleAuthor.ID == uuid.Nil {
		b.Store.AuthorBySlug(&sampleAuthor, herman, []string{"id"})
	}
	if sampleBook.ID == uuid.Nil {
		b.Store.BookBySlug(&sampleBook, moby, []string{"id"})
	}
	if samplePage.ID == uuid.Nil {
		b.Store.PageByBookAndNumber(&samplePage, &sampleBook.ID, 1, []string{"id"})
	}

	data := IndexData(
		sampleBook.ID.String(),
		"samplePage.ID.String()",
		sampleAuthor.ID.String(),
	)
	tmpl.Execute(w, data[langCode])

}
