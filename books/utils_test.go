package books_test

import (
	"github.com/djangulo/library/books"
	// "github.com/djangulo/library/books/testutils"
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
		{"idempotency", veryUnique, veryUnique},
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

// func TestGutenbergMeta(t *testing.T) {
// 	cases := []struct {
// 		name string
// 		in   string
// 		want books.Book
// 	}{
// 		{
// 			"parent's assistant",
// 			"[The Parent's Assistant, by Maria Edgeworth]",
// 			books.Book{
// 				Title:  "The Parent's Assistant",
// 				Slug:   "the-parents-assistant",
// 				Source: books.NewNullString("nltk-gutenberg"),
// 			},
// 		},
// 		{
// 			"emma",
// 			"[Emma by Jane Austen 1816]",
// 			books.Book{
// 				Title:           "Emma",
// 				PublicationYear: books.NewNullInt64(1816),
// 				Slug:            "emma",
// 				Source:          books.NewNullString("nltk-gutenberg"),
// 			},
// 		},
// 		{
// 			"the king james bible",
// 			"[The King James Bible]",
// 			books.Book{
// 				Title:           "The King James Bible",
// 				PublicationYear: books.NewNullInt64(0),
// 				Slug:            "the-king-james-bible",
// 				Source:          books.NewNullString("nltk-gutenberg"),
// 			},
// 		},
// 		{
// 			"hamlet",
// 			"[The Tragedie of Hamlet by William Shakespeare 1599]",
// 			books.Book{
// 				Title:           "The Tragedie of Hamlet",
// 				PublicationYear: books.NewNullInt64(1599),
// 				Slug:            "the-tragedie-of-hamlet",
// 				Source:          books.NewNullString("nltk-gutenberg"),
// 			},
// 		},
// 		{
// 			"stories to tell children",
// 			"[Stories to Tell to Children by Sara Cone Bryant 1918] ",
// 			books.Book{
// 				Title:           "Stories to Tell to Children",
// 				PublicationYear: books.NewNullInt64(1918),
// 				Slug:            "stories-to-tell-to-children",
// 				Source:          books.NewNullString("nltk-gutenberg"),
// 			},
// 		},
// 		{
// 			"paradise lost",
// 			"[Paradise Lost by John Milton 1667] ",
// 			books.Book{
// 				Title:           "Paradise Lost",
// 				PublicationYear: books.NewNullInt64(1667),
// 				Slug:            "paradise-lost",
// 				Source:          books.NewNullString("nltk-gutenberg"),
// 			},
// 		},
// 	}
// 	for _, c := range cases {
// 		t.Run(c.name, func(t *testing.T) {
// 			author, book := books.GutenbergMeta(c.in, false)
// 			want := c.want
// 			if !reflect.DeepEqual(author, books.Author{}) {
// 				want.AuthorID = books.NewNullUUID(author.ID.String())
// 			}
// 			if !reflect.DeepEqual(book, want) {
// 				t.Errorf("\ngot:\n%v \nwant: \n%v", book, want)
// 			}
// 		})
// 	}
// }

func TestIsSubset(t *testing.T) {
	cases := []struct {
		name string
		A, B []string
		want bool
	}{
		{"subset", []string{"a", "b", "c", "d"}, []string{"b", "c"}, true},
		{"not a subset", []string{"a", "b", "c", "d"}, []string{"b", "h"}, false},
		{"end", []string{"a", "b", "c", "d"}, []string{"d"}, true},
		{"start", []string{"a", "b", "c", "d"}, []string{"a"}, true},
		{"superset (false)", []string{"a", "b", "c", "d"}, []string{"a", "b", "c", "d", "e"}, false},
		{"empty set (true)", []string{"a", "b", "c", "d"}, []string{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := books.IsSubset(c.A, c.B)
			if got != c.want {
				t.Errorf("got '%t' want '%t'", got, c.want)
			}
		})
	}
}

func TestIsSetDifference(t *testing.T) {
	cases := []struct {
		name       string
		A, B, want []string
	}{
		{"middle", []string{"a", "b", "c", "d"}, []string{"b", "c"}, []string{"a", "d"}},
		{"last", []string{"a", "b", "c", "d"}, []string{"d"}, []string{"a", "b", "c"}},
		{"first", []string{"a", "b", "c", "d"}, []string{"a"}, []string{"b", "c", "d"}},
		{"edges", []string{"a", "b", "c", "d"}, []string{"a", "d"}, []string{"b", "c"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := books.SetDifference(c.A, c.B)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}

// func TestParseFile(t *testing.T) {
// 	path := "/home/djangulo/go/src/github.com/djangulo/library/data/corpora/gutenberg/austen-emma.txt"
// 	author, book, pages := books.ParseFile(path, 60)
// 	fmt.Println(author)
// 	fmt.Println(book)
// 	for i := 0; i < 3; i++ {
// 		fmt.Printf("page %d:\n%+v\n", i, pages[i])
// 	}
// 	fmt.Println("pages: ", len(pages))
// }
