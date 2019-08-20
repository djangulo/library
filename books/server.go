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
	store        Store
	templatesDir string
	http.Handler
	rootQuery *RootQuery
}

// Store noqa
type Store interface {
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
}

// NewBookServer returns a new server instance
func NewBookServer(store Store, middlewares []func(http.Handler) http.Handler, developmentMode bool) (*BookServer, error) {
	b := new(BookServer)

	b.templatesDir = htmlDirPath

	b.store = store
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

type key int

var langKey key = 100000001

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
	sampleBook, _ := b.store.Books(1, 0)
	samplePage, _ := b.store.Pages(1, 0)
	sampleAuthor, _ := b.store.Authors(1, 0)

	data := IndexData(
		sampleBook[0].ID.String(),
		samplePage[0].ID.String(),
		sampleAuthor[0].ID.String(),
	)
	tmpl.Execute(w, data[langCode])

}
