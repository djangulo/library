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
)

var (
	cnf         = config.Get()
	htmlDirPath = filepath.Join(cnf.Project.Dirs.Static, "books", "html")
)

// BookServer GraphQL Server for book storage
type BookServer struct {
	Store        Store
	Cache        Store
	templatesDir string
	http.Handler
	rootQuery *RootQuery
}

// Store noqa
type Store interface {
	IsAvailable() error
	Books(int, int) ([]Book, error)
	BookByID(uuid.UUID) (Book, error)
	BookBySlug(string) (Book, error)
	BooksByAuthor(string) ([]Book, error)
	Pages(int, int) ([]Page, error)
	PageByID(uuid.UUID) (Page, error)
	PageByBookAndNumber(uuid.UUID, int) (Page, error)
	Authors(int, int) ([]Author, error)
	AuthorByID(uuid.UUID) (Author, error)
	AuthorBySlug(string) (Author, error)
	InsertBook(Book) error
	InsertPage(Page) error
	InsertAuthor(Author) error
	BulkInsertBooks([]Book) error
	BulkInsertPages([]Page) error
	BulkInsertAuthors([]Author) error
}

// Cache noqa
type Cache interface {
	IsAvailable() error
	BookByID(uuid.UUID) (Book, error)
	BookBySlug(string) (Book, error)
	GetBookQuery(string) ([]Book, error)
	SaveBookQuery(string, []Book) error
	GetPageQuery(string) ([]Page, error)
	SavePageQuery(string, []Page) error
	GetAuthorQuery(string) ([]Author, error)
	SaveAuthorQuery(string, []Author) error
	InsertBook(Book) error
	InsertPage(Page) error
	InsertAuthor(Author) error
	PageByID(uuid.UUID) (Page, error)
	PageByBookAndNumber(uuid.UUID, int) (Page, error)
	AuthorByID(uuid.UUID) (Author, error)
	AuthorBySlug(string) (Author, error)
}

// NewBookServer returns a new server instance
func NewBookServer(
	store Store,
	cache Store,
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

// type LanguageCode struct {
// 	languageCode string `json:"languageCode"`
// }

// func (l LanguageCode) String() string {
// 	return fmt.Sprintf("%v", string(l.languageCode))
// }

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
	sampleBook, _ := b.Store.Books(1, 0)
	samplePage, _ := b.Store.Pages(1, 0)
	sampleAuthor, _ := b.Store.Authors(1, 0)

	data := IndexData(
		sampleBook[0].ID.String(),
		samplePage[0].ID.String(),
		sampleAuthor[0].ID.String(),
	)
	tmpl.Execute(w, data[langCode])

}
