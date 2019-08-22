package books

import (
	"github.com/gofrs/uuid"
	"github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

// BookResolver GraphqlResolver for `book` queries
func (b *BookServer) BookResolver(p graphql.ResolveParams) (interface{}, error) {
	id, idOK := p.Args["id"].(string)
	slug, slugOK := p.Args["slug"].(string)
	var book Book
	var err error
	var uid uuid.UUID

	switch {
	case idOK:
		uid, err = uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}

		err = b.Cache.IsAvailable()
		if err == nil {
			book, err = b.Cache.BookByID(uid)
		}
		if err != nil && book.ID == uuid.Nil {
			book, err := b.Store.BookByID(uid)
			if err != nil {
				return nil, errors.Wrap(err, "cannot get from db")
			}

			err = b.Cache.InsertBook(book)
			if err != nil {
				return nil, errors.Wrap(err, "failed to add book to cache")
			}
			return book, err
		}

		book, err = b.Store.BookByID(uid)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		return book, nil

	case slugOK:
		book, err = b.Store.BookBySlug(slug)
		if err != nil {
			return nil, errors.Wrap(err, "BookBySlug failed")
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
		books, err := b.Store.BooksByAuthor(author)
		if err != nil {
			return nil, errors.Wrap(err, "BookByAuthor failed")
		}
		return books, nil
	default:
		books, err := b.Store.Books(lim, off)
		if err != nil {
			return nil, errors.Wrap(err, "could not get Books from store")
		}
		return books, nil
	}
}

// PageResolver GraphqlResolver for `page` queries
func (b *BookServer) PageResolver(p graphql.ResolveParams) (interface{}, error) {
	id, idOK := p.Args["id"].(string)
	bookID, bookIDOK := p.Args["bookId"].(string)
	number, numberOK := p.Args["number"].(int)

	switch {
	case idOK:
		uid, err := uuid.FromString(id)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}
		page, err := b.Store.PageByID(uid)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get from db")
		}
		return page, nil

	case bookIDOK && numberOK:
		bookUUID, err := uuid.FromString(bookID)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing UUID")
		}
		page, err := b.Store.PageByBookAndNumber(bookUUID, number)
		if err != nil {
			return nil, errors.Wrap(err, "PageByBookAndNumber failed")
		}
		return page, nil

	default:
		return nil, nil
	}
}

// AllPageResolver returns all pages
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
	pages, err := b.Store.Pages(lim, off)
	if err != nil {
		return nil, errors.Wrap(err, "could not get Books from store")
	}
	return pages, nil
}
