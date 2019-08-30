package books

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/pkg/errors"
	"log"
	"time"
)

func getSelectedFields(selectionPath []string,
	resolveParams graphql.ResolveParams) []string {
	fields := resolveParams.Info.FieldASTs
	for _, propName := range selectionPath {
		found := false
		for _, field := range fields {
			if field.Name.Value == propName {
				selections := field.SelectionSet.Selections
				fields = make([]*ast.Field, 0)
				for _, selection := range selections {
					fields = append(fields, selection.(*ast.Field))
				}
				found = true
				break
			}
		}
		if !found {
			return []string{}
		}
	}
	collect := make([]string, 0)
	for _, field := range fields {
		collect = append(collect, field.Name.Value)
	}
	return collect
}

// BookResolver GraphqlResolver for `book` queries
func (b *BookServer) BookResolver(p graphql.ResolveParams) (interface{}, error) {
	id, idOK := p.Args["id"].(string)
	slug, slugOK := p.Args["slug"].(string)

	fields := getSelectedFields([]string{"book"}, p)

	switch {
	case idOK:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			book, err := b.Cache.BookByID(uid, fields)
			if err != nil {
				log.Println(err)
			}
			if book.ID != uuid.Nil {
				return book, nil
			}
		}
		book, err := b.Store.BookByID(uid, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.InsertBook(book)
			if err != nil {
				log.Println(err)
			}
		}
		return book, nil

	case slugOK:

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			book, err := b.Cache.BookBySlug(slug, fields)
			if err != nil {
				log.Println(err)
			}
			if book.ID != uuid.Nil {
				return book, nil
			}
		}
		book, err := b.Store.BookBySlug(slug, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.InsertBook(book)
			if err != nil {
				log.Println(err)
			}
		}
		return book, nil

	default:
		return nil, nil
	}
}

// AllBookResolver returns all books
func (b *BookServer) AllBookResolver(p graphql.ResolveParams) (interface{}, error) {
	limit, limitOK := p.Args["limit"].(int)
	offset, offsetOK := p.Args["offset"].(int)
	author, authorOK := p.Args["author"].(string)
	lastID, lastIDOK := p.Args["last_id"].(string)
	lastCreated, lastCreatedOK := p.Args["last_created_at"].(string)

	fields := getSelectedFields([]string{"allBook"}, p)

	var lim int
	if limitOK {
		lim = limit
	} else {
		lim = -1
	}

	var off int
	if offsetOK {
		off = offset
	} else {
		off = 0
	}

	var err error

	var uid uuid.UUID
	if lastIDOK {
		uid, err = uuid.FromString(lastID)
		if err != nil {
			return nil, errors.Wrap(
				err,
				fmt.Sprintf("error parsing UUID (last_id): %v", lastID),
			)
		}
	}

	var timestamp time.Time
	if lastCreatedOK {
		timestamp, err = time.Parse(time.RFC3339, lastCreated)
		if err != nil {
			return nil, errors.Wrap(
				err,
				fmt.Sprintf(
					"error parsing datetime (last_created_at): %v",
					lastCreated,
				),
			)
		}
	}

	switch {
	case authorOK:
		var books []Book
		var err error

		// query cache disabled
		// if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
		// 	if lastIDOK && lastCreatedOK {
		// 		books, err = b.Cache.BooksByAuthor(
		// 			author, lim, 0, uid, timestamp, fields)
		// 	} else {
		// 		books, err = b.Cache.BooksByAuthor(author,
		// 			lim, off, uuid.Nil, time.Time{}, fields)
		// 	}
		// 	if err != nil {
		// 		log.Println(err)
		// 	}
		// 	// cache return only if it meets expectations
		// 	if len(books) == lim {
		// 		return books, nil
		// 	}
		// }
		if lastIDOK && lastCreatedOK {
			books, err = b.Store.BooksByAuthor(
				author, lim, 0, uid, timestamp, fields)
		} else {
			books, err = b.Store.BooksByAuthor(author,
				lim, off, uuid.Nil, time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.BulkInsertBooks(books)
			if err != nil {
				log.Println(err)
			}
		}
		return books, nil
	default:
		var books []Book
		var err error
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			if lastIDOK && lastCreatedOK {
				books, err = b.Cache.Books(lim, 0, uid, timestamp, fields)
			} else {
				books, err = b.Cache.Books(lim, off, uuid.Nil, time.Time{}, fields)
			}
			if err != nil {
				log.Println(err)
			}
			if books != nil {
				return books, nil
			}
		}
		if lastIDOK && lastCreatedOK {
			books, err = b.Store.Books(lim, 0, uid, timestamp, fields)
		} else {
			books, err = b.Store.Books(lim, off, uuid.Nil, time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.BulkInsertBooks(books)
			if err != nil {
				log.Println(err)
			}
		}
		return books, nil
	}
}

// PageResolver GraphqlResolver for `page` queries
func (b *BookServer) PageResolver(p graphql.ResolveParams) (interface{}, error) {
	id, idOK := p.Args["id"].(string)
	bookID, bookIDOK := p.Args["book_id"].(string)
	number, numberOK := p.Args["number"].(int)

	fields := getSelectedFields([]string{"page"}, p)

	switch {
	case idOK:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			page, err := b.Cache.PageByID(uid, fields)
			if err != nil {
				log.Println(err)
			}
			if page.ID != uuid.Nil {
				return page, nil
			}
		}
		page, err := b.Store.PageByID(uid, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.InsertPage(page)
			if err != nil {
				log.Println(err)
			}
		}
		return page, nil

	case bookIDOK && numberOK:
		bookUUID, err := uuid.FromString(bookID)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			page, err := b.Cache.PageByBookAndNumber(bookUUID, number, fields)
			if err != nil {
				log.Println(err)
			}
			if page.ID != uuid.Nil {
				return page, nil
			}
		}
		page, err := b.Store.PageByBookAndNumber(bookUUID, number, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.InsertPage(page)
			if err != nil {
				log.Println(err)
			}
		}
		return page, nil

	default:
		return nil, nil
	}
}

// AllPageResolver returns all books
func (b *BookServer) AllPageResolver(p graphql.ResolveParams) (interface{}, error) {
	limit, limitOK := p.Args["limit"].(int)
	offset, offsetOK := p.Args["offset"].(int)
	lastID, lastIDOK := p.Args["last_id"].(string)
	lastCreated, lastCreatedOK := p.Args["last_created_at"].(string)

	fields := getSelectedFields([]string{"allPage"}, p)

	var lim int
	if limitOK {
		lim = limit
	} else {
		lim = -1
	}

	var off int
	if offsetOK {
		off = offset
	} else {
		off = 0
	}

	var err error

	var uid uuid.UUID
	if lastIDOK {
		uid, err = uuid.FromString(lastID)
		if err != nil {
			return nil, errors.Wrap(
				err,
				fmt.Sprintf("error parsing UUID (last_id): %v", lastID),
			)
		}
	}

	var timestamp time.Time
	if lastCreatedOK {
		timestamp, err = time.Parse(time.RFC3339, lastCreated)
		if err != nil {
			return nil, errors.Wrap(
				err,
				fmt.Sprintf(
					"error parsing datetime (last_created_at): %v",
					lastCreated,
				),
			)
		}
	}

	switch {
	default:
		var pages []Page
		var err error
		// if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
		// 	if lastIDOK && lastCreatedOK {
		// 		pages, err = b.Cache.Pages(lim, 0, uid, timestamp, fields)
		// 	} else {
		// 		pages, err = b.Cache.Pages(lim, off, uuid.Nil, time.Time{}, fields)
		// 	}
		// 	if err != nil {
		// 		log.Println(err)
		// 	}
		// 	if pages != nil {
		// 		return pages, nil
		// 	}
		// }
		if lastIDOK && lastCreatedOK {
			pages, err = b.Store.Pages(lim, 0, uid, timestamp, fields)
		} else {
			pages, err = b.Store.Pages(lim, off, uuid.Nil, time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.BulkInsertPages(pages)
			if err != nil {
				log.Println(err)
			}
		}
		return pages, nil
	}
}

// AuthorResolver GraphqlResolver for `book` queries
func (b *BookServer) AuthorResolver(p graphql.ResolveParams) (interface{}, error) {
	id, idOK := p.Args["id"].(string)
	name, nameOK := p.Args["name"].(string)

	fields := getSelectedFields([]string{"author"}, p)

	switch {
	case idOK:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			author, err := b.Cache.AuthorByID(uid, fields)
			if err != nil {
				log.Println(err)
			}
			if author.ID != uuid.Nil {
				return author, nil
			}
		}
		author, err := b.Store.AuthorByID(uid, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.InsertAuthor(author)
			if err != nil {
				log.Println(err)
			}
		}
		return author, nil

	case nameOK:

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			author, err := b.Cache.AuthorBySlug(name, fields)
			if err != nil {
				log.Println(err)
			}
			if author.ID != uuid.Nil {
				return author, nil
			}
		}
		author, err := b.Store.AuthorBySlug(name, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.InsertAuthor(author)
			if err != nil {
				log.Println(err)
			}
		}
		return author, nil

	default:
		return nil, nil
	}
}

// AllAuthorResolver returns all books
func (b *BookServer) AllAuthorResolver(p graphql.ResolveParams) (interface{}, error) {
	limit, limitOK := p.Args["limit"].(int)
	offset, offsetOK := p.Args["offset"].(int)
	lastID, lastIDOK := p.Args["last_id"].(string)
	lastCreated, lastCreatedOK := p.Args["last_created_at"].(string)

	fields := getSelectedFields([]string{"allPage"}, p)

	var lim int
	if limitOK {
		lim = limit
	} else {
		lim = -1
	}

	var off int
	if offsetOK {
		off = offset
	} else {
		off = 0
	}

	var err error

	var uid uuid.UUID
	if lastIDOK {
		uid, err = uuid.FromString(lastID)
		if err != nil {
			return nil, errors.Wrap(
				err,
				fmt.Sprintf("error parsing UUID (last_id): %v", lastID),
			)
		}
	}

	var timestamp time.Time
	if lastCreatedOK {
		timestamp, err = time.Parse(time.RFC3339, lastCreated)
		if err != nil {
			return nil, errors.Wrap(
				err,
				fmt.Sprintf(
					"error parsing datetime (last_created_at): %v",
					lastCreated,
				),
			)
		}
	}

	switch {
	default:
		var authors []Author
		var err error
		// if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
		// 	if lastIDOK && lastCreatedOK {
		// 		authors, err = b.Cache.Authors(lim, 0, uid, timestamp, fields)
		// 	} else {
		// 		authors, err = b.Cache.Authors(lim, off, uuid.Nil, time.Time{}, fields)
		// 	}
		// 	if err != nil {
		// 		log.Println(err)
		// 	}
		// 	if authors != nil {
		// 		return authors, nil
		// 	}
		// }
		if lastIDOK && lastCreatedOK {
			authors, err = b.Store.Authors(lim, 0, uid, timestamp, fields)
		} else {
			authors, err = b.Store.Authors(lim, off, uuid.Nil, time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			err = b.Cache.BulkInsertAuthors(authors)
			if err != nil {
				log.Println(err)
			}
		}
		return authors, nil
	}
}
