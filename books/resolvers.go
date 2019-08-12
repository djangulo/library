package books

import (
	"github.com/gofrs/uuid"
	"github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

// BooksResolver GraphqlResolver for `book` queries
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
}

// AllBooksResolver returns all books
func (b *BookServer) AllBooksResolver(p graphql.ResolveParams) (interface{}, error) {
	limit, limitOk := p.Args["limit"].(int)
	offset, offsetOk := p.Args["offset"].(int)
	author, authorOk := p.Args["author"].(string)

	var lim int
	if limitOk {
		lim = limit
	} else {
		lim = -1
	}

	var off int
	if offsetOk {
		off = offset
	} else {
		off = 0
	}
	switch {
	case authorOk:
		books, err := b.store.BooksByAuthor(author)
		if err != nil {
			return nil, errors.Wrap(err, "BookByAuthor failed")
		}
		return books, nil

	default:
		books, err := b.store.Books(lim, off)
		if err != nil {
			return nil, errors.Wrap(err, "could not get Books from store")
		}
		return books, nil
	}
}
