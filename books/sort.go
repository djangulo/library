package books

import (
	"sort"
)

type BookBy func(b1, b2 *Book) bool

func (by BookBy) Sort(books []Book) {
	ps := &bookSorter{
		books: books,
		by:    by,
	}
	sort.Sort(ps)
}

type bookSorter struct {
	books []Book
	by    func(b1, b2 *Book) bool
}

func (b *bookSorter) Len() int {
	return len(b.books)
}

func (b *bookSorter) Swap(i, j int) {
	b.books[i], b.books[j] = b.books[j], b.books[i]
}

func (b *bookSorter) Less(i, j int) bool {
	return b.by(&b.books[i], &b.books[j])
}

var BookIDSortDesc = func(b1, b2 *Book) bool {
	return b1.ID.String() < b2.ID.String()
}

var BookIDSortAsc = func(b1, b2 *Book) bool {
	return BookIDSortDesc(b2, b1)
}

type PageBy func(b1, b2 *Page) bool

func (by PageBy) Sort(pages []Page) {
	ps := &pageSorter{
		pages: pages,
		by:    by,
	}
	sort.Sort(ps)
}

type pageSorter struct {
	pages []Page
	by    func(b1, b2 *Page) bool
}

func (b *pageSorter) Len() int {
	return len(b.pages)
}

func (b *pageSorter) Swap(i, j int) {
	b.pages[i], b.pages[j] = b.pages[j], b.pages[i]
}

func (b *pageSorter) Less(i, j int) bool {
	return b.by(&b.pages[i], &b.pages[j])
}

var PageIDSortDesc = func(b1, b2 *Page) bool {
	return b1.ID.String() < b2.ID.String()
}

var PageIDSortAsc = func(b1, b2 *Page) bool {
	return PageIDSortDesc(b2, b1)
}

type PtrPageBy func(b1, b2 *Page) bool

func (by PtrPageBy) Sort(pages []*Page) {
	ps := &pageSorterPtr{
		pages: pages,
		by:    by,
	}
	sort.Sort(ps)
}

type pageSorterPtr struct {
	pages []*Page
	by    func(b1, b2 *Page) bool
}

func (b *pageSorterPtr) Len() int {
	return len(b.pages)
}

func (b *pageSorterPtr) Swap(i, j int) {
	b.pages[i], b.pages[j] = b.pages[j], b.pages[i]
}

func (b *pageSorterPtr) Less(i, j int) bool {
	return b.by(b.pages[i], b.pages[j])
}
