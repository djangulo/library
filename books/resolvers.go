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
		var book Book
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			if err := b.Cache.BookByID(&book, &uid, fields); err != nil {
				log.Println(err)
			} else {
				return book, nil
			}
		}
		err = b.Store.BookByID(&book, &uid, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.InsertBook(&book)
			if err != nil {
				log.Println(err)
			}
		}
		return book, nil

	case slugOK:
		var book Book

		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			if err := b.Cache.BookBySlug(&book, &slug, fields); err != nil {
				log.Println(err)
			} else {
				return book, nil
			}
		}
		err := b.Store.BookBySlug(&book, &slug, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.InsertBook(&book)
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
		books := make([]*Book, 0)
		var err error
		var key string

		if lastIDOK && lastCreatedOK {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"BooksByAuthor(%s,%d,%d,%v,%v,%v)",
					author, lim, 0, uid, timestamp, fields,
				)
				if err := b.Cache.BookQuery(&books, key); err != nil {
					log.Println(err)
				}
				if len(books) > 0 {
					return books, nil
				}
			}
			var zero int
			err = b.Store.BooksByAuthor(books,
				&author, &lim, &zero, &uid, &timestamp, fields)
		} else {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"BooksByAuthor(%s,%d,%d,%v,%v,%v)",
					author, lim, off, uuid.Nil, timestamp, fields,
				)
				if err := b.Cache.BookQuery(&books, key); err != nil {
					log.Println(err)
				}
				if len(books) > 0 {
					return books, nil
				}
			}
			err = b.Store.BooksByAuthor(books,
				&author, &lim, &off, &uuid.UUID{}, &time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.SaveBookQuery(key, books)
			if err != nil {
				log.Println(err)
			}
		}
		return books, nil
	default:
		books := make([]*Book, 0)
		var err error
		var key string

		if lastIDOK && lastCreatedOK {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"Books(%d,%d,%v,%v,%v)",
					lim, 0, uid, timestamp, fields,
				)
				if err := b.Cache.BookQuery(&books, key); err != nil {
					log.Println(err)
				}
				if len(books) > 0 {
					return books, nil
				}
			}
			var zero int
			err = b.Store.Books(books,
				&lim, &zero, &uid, &timestamp, fields)
		} else {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"Books(%d,%d,%v,%v,%v)",
					lim, off, uuid.Nil, timestamp, fields,
				)
				if err := b.Cache.BookQuery(&books, key); err != nil {
					log.Println(err)
				}
				if len(books) > 0 {
					return books, nil
				}
			}
			err = b.Store.Books(books,
				&lim, &off, &uuid.UUID{}, &time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.SaveBookQuery(key, books)
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
		var page Page
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			if err := b.Cache.PageByID(&page, &uid, fields); err != nil {
				log.Println(err)
			} else {
				return page, nil
			}
		}
		err = b.Store.PageByID(&page, &uid, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.InsertPage(&page)
			if err != nil {
				log.Println(err)
			}
		}
		return page, nil

	case bookIDOK && numberOK:
		var page Page
		uid, err := uuid.FromString(bookID)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			if err := b.Cache.PageByBookAndNumber(
				&page,
				&uid,
				&number,
				fields,
			); err != nil {
				log.Println(err)
			} else {
				return page, nil
			}
		}
		err = b.Store.PageByBookAndNumber(&page, &uid, &number, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.InsertPage(&page)
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
		pages := make([]*Page, 0)
		var err error
		var key string

		if lastIDOK && lastCreatedOK {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"Pages(%d,%d,%v,%v,%v)",
					lim, 0, uid, timestamp, fields,
				)
				if err := b.Cache.PageQuery(&pages, key); err != nil {
					log.Println(err)
				}
				if len(pages) > 0 {
					return pages, nil
				}
			}
			var zero int
			err = b.Store.Pages(pages,
				&lim, &zero, &uid, &timestamp, fields)
		} else {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"Pages(%d,%d,%v,%v,%v)",
					lim, off, uuid.Nil, timestamp, fields,
				)
				if err := b.Cache.PageQuery(&pages, key); err != nil {
					log.Println(err)
				}
				if len(pages) > 0 {
					return pages, nil
				}
			}
			err = b.Store.Pages(pages,
				&lim, &off, &uuid.UUID{}, &time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.SavePageQuery(key, pages)
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
		var author Author
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			if err := b.Cache.AuthorByID(&author, &uid, fields); err != nil {
				log.Println(err)
			} else {
				return author, nil
			}
		}
		err = b.Store.AuthorByID(&author, &uid, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.InsertAuthor(&author)
			if err != nil {
				log.Println(err)
			}
		}
		return author, nil

	case nameOK:

		var author Author

		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			if err := b.Cache.AuthorBySlug(&author, &name, fields); err != nil {
				log.Println(err)
			} else {
				return author, nil
			}
		}
		err := b.Store.AuthorBySlug(&author, &name, fields)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.InsertAuthor(&author)
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

	fields := getSelectedFields([]string{"allAuthor"}, p)

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
		authors := make([]*Author, 0)
		var err error
		var key string

		if lastIDOK && lastCreatedOK {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"Authors(%d,%d,%v,%v,%v)",
					lim, 0, uid, timestamp, fields,
				)
				if err := b.Cache.AuthorQuery(&authors, key); err != nil {
					log.Println(err)
				}
				if len(authors) > 0 {
					return authors, nil
				}
			}
			var zero int
			err = b.Store.Authors(authors,
				&lim, &zero, &uid, &timestamp, fields)
		} else {
			if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
				key = fmt.Sprintf(
					"Authors(%d,%d,%v,%v,%v)",
					lim, off, uuid.Nil, timestamp, fields,
				)
				if err := b.Cache.AuthorQuery(&authors, key); err != nil {
					log.Println(err)
				}
				if len(authors) > 0 {
					return authors, nil
				}
			}
			err = b.Store.Authors(authors,
				&lim, &off, &uuid.UUID{}, &time.Time{}, fields)
		}
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		if cacheOKErr := b.Cache.IsAvailable(); cacheOKErr == nil {
			err = b.Cache.SaveAuthorQuery(key, authors)
			if err != nil {
				log.Println(err)
			}
		}
		return authors, nil
	}
}
