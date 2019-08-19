package books_test

import (
	"fmt"
	"github.com/djangulo/library/books"
	"reflect"
	"testing"
)

func TestSlugify(t *testing.T) {
	veryUnique := "the-very-unique-name"
	cases := []struct{ name, in, want string }{
		{"spaces", "The Very Unique Name", veryUnique},
		{"multiple spaces", "  tHe    VeRy     UnIqUe     nAmE   ", veryUnique},
		{"underscores", "__tHe_VeRy___UnIqUe__nAmE   ", veryUnique},
		{"dots", "...tHe..VeRy.UnIqUe....nAmE...", veryUnique},
		{"commas", "tHe,VeRy,,,UnIqUe,,,,nAmE,", veryUnique},
		{"slashes", "tHe\\VeRy\\\\UnIqUe/\\/\\/\\nAmE\\\\//\\\\//\\\\//", veryUnique},
		{"percent", `tHe%VeRy%%UnIqUe%nAmE%%`, veryUnique},
		{"octochorpe", "###tHe##VeRy#UnIqUe###nAmE####", veryUnique},
		{"dollar", "$$tHe$$$VeRy$$$UnIqUe$$nAmE$$ $", veryUnique},
		{"exclamation", "!!!tHe!VeRy!UnIqUe!!!nAmE!!!!", veryUnique},
		{"multiple dashes", "----tHe---VeRy--UnIqUe--nAmE-----", veryUnique},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := books.Slugify(c.in, "-")
			want := c.want
			if got != want {
				t.Errorf("got '%s' want '%s'", got, want)
			}
		})
	}
}

func TestGutenbergMeta(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want books.Book
	}{
		{
			"parent's assistant",
			"[The Parent's Assistant, by Maria Edgeworth]",
			books.Book{
				Title:  "The Parent's Assistant",
				Slug:   "the-parents-assistant",
				Source: books.NewNullString("gutenberg"),
			},
		},
		{
			"emma",
			"[Emma by Jane Austen 1816]",
			books.Book{
				Title:           "Emma",
				PublicationYear: books.NewNullInt64(1816),
				Slug:            "emma",
				Source:          books.NewNullString("gutenberg"),
			},
		},
		{
			"the king james bible",
			"[The King James Bible]",
			books.Book{
				Title:           "The King James Bible",
				PublicationYear: books.NewNullInt64(0),
				Slug:            "the-king-james-bible",
				Source:          books.NewNullString("gutenberg"),
			},
		},
		{
			"hamlet",
			"[The Tragedie of Hamlet by William Shakespeare 1599]",
			books.Book{
				Title:           "The Tragedie of Hamlet",
				PublicationYear: books.NewNullInt64(1599),
				Slug:            "the-tragedie-of-hamlet",
				Source:          books.NewNullString("gutenberg"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			author, book := books.GutenbergMeta(c.in)
			want := c.want
			if !reflect.DeepEqual(author, books.Author{}) {
				want.AuthorID = &author.ID
			}
			if !reflect.DeepEqual(book, want) {
				t.Errorf("\ngot:\n%v \nwant: \n%v", book, want)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	path := "/home/djangulo/go/src/github.com/djangulo/library/data/corpora/gutenberg/austen-emma.txt"
	author, book, pages := books.ParseFile(path, 60)
	fmt.Println(author)
	fmt.Println(book)
	for i := 0; i < 3; i++ {
		fmt.Printf("page %d:\n%+v\n", i, pages[i])
	}
	fmt.Println("pages: ", len(pages))
}
