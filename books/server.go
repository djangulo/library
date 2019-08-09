package books

import (
	"context"
	// "encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gofrs/uuid"
	"github.com/graphql-go/graphql"
	"html/template"
	// "io/ioutil"
	"net/http"
	"path/filepath"
	// "strconv"
	"github.com/pkg/errors"
	"log"
)

const (
	htmlDirTemplatesPath = "html"
)

type BookServer struct {
	store        Store
	templatesDir string
	http.Handler
	rootQuery *RootQuery
}

type Store interface {
	Books() ([]Book, error)
	BookByID(ID uuid.UUID) (Book, error)
	BookBySlug(slug string) (Book, error)
}

func NewBookServer(store Store, schema graphql.Schema) (*BookServer, error) {
	b := new(BookServer)

	b.templatesDir = htmlDirTemplatesPath

	b.store = store
	b.rootQuery = b.NewRootQuery()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/{languageCode}", func(r chi.Router) {
		r.Use(LanguageCtx)
		r.Get("/", b.serveIndex)
	})

	r.Mount("/graphql", b.GraphQLRouter())

	b.Handler = r

	return b, nil
}

type LanguageCode string

func (l LanguageCode) String() string {
	return fmt.Sprintf("%s", string(l))
}

func LanguageCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := chi.URLParam(r, "languageCode")
		ctx := context.WithValue(r.Context(), "lang", lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (b *BookServer) serveIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langCode, ok := ctx.Value("lang").(*LanguageCode)
	fmt.Printf("\n\n%s\n\n", langCode)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	templatePath := filepath.Join(b.templatesDir, "index."+langCode.String()+".html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("problem opening %s %v", templatePath, err), 400)
	}

	tmpl.Execute(w, nil)

}

func (b *BookServer) BooksResolver(p graphql.ResolveParams) (interface{}, error) {
	id, idOk := p.Args["id"].(string)
	slug, slugOk := p.Args["slug"].(string)

	switch {
	case idOk:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}
		book, err := b.store.BookByID(uid)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		return book, nil

	case slugOk:
		book, err := b.store.BookBySlug(slug)
		if err != nil {
			return nil, errors.Wrap(err, "BookBySlug failed")
		}
		return book, nil
	default:
		return nil, nil
	}
	return nil, nil
}

func (b *BookServer) AllBooksResolver(p graphql.ResolveParams) (interface{}, error) {

}

// func (p *Book) serveGame(w http.ResponseWriter, r *http.Request) {
// 	p.template.Execute(w, nil)
// }

// func (p *PlayerServer) rootHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprint(w, "Hello, World!")
// }
// func (p *PlayerServer) leagueHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("content-type", "application/json")
// 	json.NewEncoder(w).Encode(p.store.GetLeague())
// }
// func (p *PlayerServer) playersHandler(w http.ResponseWriter, r *http.Request) {
// 	player := r.URL.Path[len("/players/"):]
// 	switch r.Method {
// 	case http.MethodPost:
// 		p.processWin(w, player)
// 	case http.MethodGet:
// 		p.showScore(w, player)
// 	}
// }

// func (p *PlayerServer) getLeagueTable() League {
// 	return League{
// 		{Name: "Denis", Wins: 20},
// 	}
// }

// func (p *PlayerServer) showScore(w http.ResponseWriter, player string) {
// 	score := p.store.GetPlayerScore(player)
// 	if score == 0 {
// 		w.WriteHeader(http.StatusNotFound)
// 	}
// 	fmt.Fprint(w, score)
// }

// func (p *PlayerServer) processWin(w http.ResponseWriter, player string) {
// 	p.store.RecordWin(player)
// 	w.WriteHeader(http.StatusAccepted)
// }

// var wsUpgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

// func (p *PlayerServer) webSocket(w http.ResponseWriter, r *http.Request) {
// 	conn, _ := wsUpgrader.Upgrade(w, r, nil)

// 	_, numberOfPlayersMsg, _ := conn.ReadMessage()
// 	numberOfPlayers, _ := strconv.Atoi(string(numberOfPlayersMsg))
// 	p.game.Start(numberOfPlayers, ioutil.Discard) //todo: Dont discard the blinds messages!

// 	_, winner, _ := conn.ReadMessage()
// 	p.game.Finish(string(winner))

// }

// func GetPlayerScore(name string) int {
// 	if name == "Pepper" {
// 		return 20
// 	}
// 	if name == "Floyd" {
// 		return 10
// 	}
// 	return 0
// }
