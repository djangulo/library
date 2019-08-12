package books

import (
	// "database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/graphql-go/graphql"
	// "github.com/graphql-go/graphql/language/ast"
	"github.com/pkg/errors"
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
		fmt.Printf("%+v\n", errors.Wrap(err, "error creating graphql schema"))
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
						Args: graphql.FieldConfigArgument{
							"limit": &graphql.ArgumentConfig{
								Type:         graphql.Int,
								DefaultValue: 1000,
								Description:  "Limit query size",
							},
							"author": &graphql.ArgumentConfig{
								Type:        graphql.String,
								Description: "Filter by author",
							},
						},
					},
				},
			},
		),
	}
	return &root
}

// type NullString struct {
// 	graphql.Scalar
// }

// func (s NullString) String() string {
// 	var str string
// 	if s.Valid {
// 		str = s.Str
// 	}
// 	return str
// }

// func NewNullString(v sql.NullString) *NullString {
// 	return &NullString{Str: v, Valid: true}
// }

// func SerializeNullString(value interface{}) interface{} {
// 	switch value := value.(type) {
// 	case NullString:
// 		return value.Str
// 	case *NullString:
// 		v := *value
// 		return value.Str
// 	case sql.NullString:
// 		return value.String
// 	case *sql.NullString:
// 		v := *value
// 		return value.String
// 	default:
// 		return nil
// 	}

// }

// var NullableString = graphql.NewScalar(graphql.ScalarConfig{
// 	Name:        "NullableString",
// 	Description: "The `NullableString` type repesents a nullable SQL string.",
// 	Serialize: func(value interface{}) interface{} {
// 		switch value := value.(type) {
// 		case NullString:
// 			return value.String()
// 		case *NullString:
// 			v := *value
// 			return v.String()
// 		default:
// 			return nil
// 		}
// 	},
// 	ParseValue: func(value interface{}) interface{} {
// 		switch vaule := value.(type) {
// 		case string:
// 			return NewNullString(value)
// 		case *string:
// 			return NewNullString(*value)
// 		default:
// 			return nil
// 		}
// 	},
// 	ParseLiteral: func(valueAST ast.Value) interface{} {
// 		switch valueAST := valueAST.(type) {
// 		case *ast.StringValue:
// 			return NewNullString(valueAST.Value)
// 		default:
// 			return nil
// 		}
// 	},
// })

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
