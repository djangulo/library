package books_test

import (
	"github.com/djangulo/library/books"
	// "github.com/djangulo/library/books/testutils"
	config "github.com/djangulo/library/config/books"
	"sort"
	"testing"
)

func BenchmarkSortMethods(b *testing.B) {
	cnf := config.Get()
	tBooks, _ := books.BookSeedData(cnf)
	pages, _ := books.PageSeedData(cnf)

	ptrPagesSlice := make([]*books.Page, len(pages), len(pages))
	for i, p := range pages {
		ptrPagesSlice[i] = &p
	}
	ptrBooksSlice := make([]*books.Book, len(tBooks), len(tBooks))
	for i, b := range tBooks {
		ptrBooksSlice[i] = &b
	}

	b.Run("books programmable sort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				books.BookBy(books.BookIDSortDesc).Sort(tBooks)
			} else {
				books.BookBy(books.BookIDSortAsc).Sort(tBooks)
			}
		}
	})
	b.Run("books slice stable (sort.SliceStable)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sort.SliceStable(tBooks, func(i, j int) bool { return tBooks[i].ID.String() < tBooks[j].ID.String() })
			} else {
				sort.SliceStable(tBooks, func(i, j int) bool { return tBooks[j].ID.String() < tBooks[i].ID.String() })
			}
		}
	})
	b.Run("books ptr slice stable (sort.SliceStable)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sort.SliceStable(ptrBooksSlice, func(i, j int) bool { return ptrBooksSlice[i].ID.String() < ptrBooksSlice[j].ID.String() })
			} else {
				sort.SliceStable(ptrBooksSlice, func(i, j int) bool { return ptrBooksSlice[j].ID.String() < ptrBooksSlice[i].ID.String() })
			}
		}
	})
	b.Run("books slice unstable (sort.Slice)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sort.Slice(tBooks, func(i, j int) bool { return tBooks[i].ID.String() < tBooks[j].ID.String() })
			} else {
				sort.Slice(tBooks, func(i, j int) bool { return tBooks[j].ID.String() < tBooks[i].ID.String() })
			}
		}
	})
	b.Run("pages all programmable sort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				books.PageBy(books.PageIDSortDesc).Sort(pages)
			} else {
				books.PageBy(books.PageIDSortAsc).Sort(pages)
			}
		}
	})
	b.Run("pages all slice stable (sort.SliceStable)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sort.SliceStable(pages, func(i, j int) bool { return pages[i].ID.String() < pages[j].ID.String() })
			} else {
				sort.SliceStable(pages, func(i, j int) bool { return pages[j].ID.String() < pages[i].ID.String() })
			}
		}
	})
	b.Run("pages all slice unstable (sort.Slice)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sort.Slice(pages, func(i, j int) bool { return pages[i].ID.String() < pages[j].ID.String() })
			} else {
				sort.Slice(pages, func(i, j int) bool { return pages[j].ID.String() < pages[i].ID.String() })
			}
		}
	})

	b.Run("pages ptr programmable sort", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				books.PtrPageBy(books.PageIDSortDesc).Sort(ptrPagesSlice)
			} else {
				books.PtrPageBy(books.PageIDSortAsc).Sort(ptrPagesSlice)
			}
		}
	})
	b.Run("pages ptr slice stable (sort.SliceStable)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sort.SliceStable(pages, func(i, j int) bool { return ptrPagesSlice[i].ID.String() < ptrPagesSlice[j].ID.String() })
			} else {
				sort.SliceStable(pages, func(i, j int) bool { return ptrPagesSlice[j].ID.String() < ptrPagesSlice[i].ID.String() })
			}
		}
	})
	b.Run("pages ptr slice unstable (sort.Slice)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				sort.Slice(ptrPagesSlice, func(i, j int) bool { return ptrPagesSlice[i].ID.String() < ptrPagesSlice[j].ID.String() })
			} else {
				sort.Slice(ptrPagesSlice, func(i, j int) bool { return ptrPagesSlice[j].ID.String() < ptrPagesSlice[i].ID.String() })
			}
		}
	})
}
