package books

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/graphql-go/graphql"
	"net/http"
)

// GraphQLRouter router for the /graphql endpoint
func (b *BookServer) GraphQLRouter() http.Handler {
	r := chi.NewRouter()
	schema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: b.rootQuery.Query,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating schema %v", err))
	}
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "Must provide the graphql query in request body", 400)
			return
		}

		var body requestBody
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing JSON request body: %v", err), 400)
		}

		result := executeQuery(body.Query, schema)

		render.JSON(w, r, result)
	})
	return r
}

// BookSchema Basic schema for graphql server
// var BookSchema, _ = graphql.NewSchema(
// 	graphql.SchemaConfig{
// 		Query: queryType,
// 	},
// )

var authorType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Author",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"slug": &graphql.Field{
				Type: graphql.String,
			},
			"books": &graphql.Field{
				Type: graphql.NewList(bookType),
			},
		},
	},
)

var bookType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Book",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"author": &graphql.Field{
				Type: graphql.String,
			},
			"slug": &graphql.Field{
				Type: graphql.String,
			},
			"publication_year": &graphql.Field{
				Type: graphql.Int,
			},
			"pages": &graphql.Field{
				Type: graphql.NewList(pageType),
			},
		},
	},
)

var pageType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Page",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"page_number": &graphql.Field{
				Type: graphql.String,
			},
			"body": &graphql.Field{
				Type: graphql.String,
			},
			"book_id": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

type RootQuery struct {
	Query *graphql.Object
}

func (b *BookServer) NewRootQuery() *RootQuery {
	root := RootQuery{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.Fields{
					"book": &graphql.Field{
						Type:        bookType,
						Description: "Get Books by ID or Slug.",
						Args: graphql.FieldConfigArgument{
							"id": &graphql.ArgumentConfig{
								Type: graphql.String,
							},
							"slug": &graphql.ArgumentConfig{
								Type: graphql.String,
							},
						},
						Resolve: b.BooksResolver,
					},
					"allBook": &graphql.Field{
						Type:        graphql.NewList(bookType),
						Description: "Get all books.",
						Resolve:     b.AllBooksResolver,
					},
				},
			},
		),
	}
	return &root
}

// var queryType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "Query",
// 		Fields: graphql.Fields{
// 			"book": &graphql.Field{
// 				Type: bookType,
// 				Args: graphql.FieldConfigArgument{
// 					"id": &graphql.ArgumentConfig{
// 						Type: graphql.String,
// 					},
// 				},
// 				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
// 					idQuery, ok := p.Args["id"].(string)
// 					if ok {
// 						uid, err := uuid.FromString(idQuery)
// 						if err != nil {
// 							log.Fatalf("failed to parse UUID %q: %v", uid, err)
// 						}
// 						fmt.Printf("\n\nReceived ID: %s\n\n", uid)
// 						return uid, nil
// 					}
// 					return nil, nil
// 				},
// 			},
// 		},
// 	},
// )

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

type requestBody struct {
	Query string `json:"query"`
}
