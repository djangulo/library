package books

import (
	"github.com/gofrs/uuid"
	"github.com/graphql-go/graphql"
	"github.com/pkg/errors"
	"log"
)

// BookResolver GraphqlResolver for `book` queries
func (b *BookServer) BookResolver(p graphql.ResolveParams) (interface{}, error) {
	id, idOK := p.Args["id"].(string)
	slug, slugOK := p.Args["slug"].(string)

	switch {
	case idOK:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			book, err := b.Cache.BookByID(uid)
			if err != nil {
				log.Println(err)
			}
			if book.ID != uuid.Nil {
				return book, nil
			}
		}
		book, err := b.Store.BookByID(uid)
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
			book, err := b.Cache.BookBySlug(slug)
			if err != nil {
				log.Println(err)
			}
			if book.ID != uuid.Nil {
				return book, nil
			}
		}
		book, err := b.Store.BookBySlug(slug)
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
	switch {
	case authorOK:
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			books, err := b.Cache.BooksByAuthor(author)
			if err != nil {
				log.Println(err)
			}
			if books != nil {
				return books, nil
			}
		}
		books, err := b.Store.BooksByAuthor(author)
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
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			books, err := b.Cache.Books(lim, off)
			if err != nil {
				log.Println(err)
			}
			if books != nil {
				return books, nil
			}
		}
		books, err := b.Store.Books(lim, off)
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

	switch {
	case idOK:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			page, err := b.Cache.PageByID(uid)
			if err != nil {
				log.Println(err)
			}
			if page.ID != uuid.Nil {
				return page, nil
			}
		}
		page, err := b.Store.PageByID(uid)
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
			page, err := b.Cache.PageByBookAndNumber(bookUUID, number)
			if err != nil {
				log.Println(err)
			}
			if page.ID != uuid.Nil {
				return page, nil
			}
		}
		page, err := b.Store.PageByBookAndNumber(bookUUID, number)
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
	switch {
	default:
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			pages, err := b.Cache.Pages(lim, off)
			if err != nil {
				log.Println(err)
			}
			if pages != nil {
				return pages, nil
			}
		}
		pages, err := b.Store.Pages(lim, off)
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

	switch {
	case idOK:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			author, err := b.Cache.AuthorByID(uid)
			if err != nil {
				log.Println(err)
			}
			if author.ID != uuid.Nil {
				return author, nil
			}
		}
		author, err := b.Store.AuthorByID(uid)
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
			author, err := b.Cache.AuthorBySlug(name)
			if err != nil {
				log.Println(err)
			}
			if author.ID != uuid.Nil {
				return author, nil
			}
		}
		author, err := b.Store.AuthorBySlug(name)
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
	switch {
	default:
		if cacheAvailableErr := b.Cache.IsAvailable(); cacheAvailableErr == nil {
			authors, err := b.Cache.Authors(lim, off)
			if err != nil {
				log.Println(err)
			}
			if authors != nil {
				return authors, nil
			}
		}
		authors, err := b.Store.Authors(lim, off)
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
