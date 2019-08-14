package books_test

import (
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
				Author: books.NewNullString("Maria Edgeworth"),
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
				Author:          books.NewNullString("Jane Austen"),
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
				Author:          books.NewNullString(""),
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
				Author:          books.NewNullString("William Shakespeare"),
				Source:          books.NewNullString("gutenberg"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := books.GutenbergMeta(c.in)
			want := c.want
			if !reflect.DeepEqual(got, want) {
				t.Errorf("\ngot:\n%v \nwant: \n%v", got, want)
			}
		})
	}
}
