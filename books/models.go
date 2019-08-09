package books

import (
	"github.com/gofrs/uuid"
)

type Book struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Slug            string    `json:"slug"`
	Author          uuid.UUID `json:"author"`
	PublicationYear int       `json:"publication_year"`
	PageCount       int       `json:"page_count"`
	Pages           []Page    `json:"pages"`
}

type Author struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Slug  string    `json:"slug"`
	Books []Book    `json:"books"`
}

type Page struct {
	ID         uuid.UUID `json:"id"`
	PageNumber int       `json:"page_number"`
	Body       string    `json:"body"`
	BookID     uuid.UUID `json:"book_id"`
}
