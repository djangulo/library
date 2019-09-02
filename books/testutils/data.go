package testutils

import (
	"encoding/json"
	"github.com/djangulo/library/books"
	config "github.com/djangulo/library/config/books"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

// TestBookData reads books json data and returns as a slice
func TestBookData() (books []*books.Book) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakeBooks.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &books)

	sort.SliceStable(books, func(i, j int) bool {
		return books[i].CreatedAt.After(books[j].CreatedAt)
	})
	sort.SliceStable(books, func(i, j int) bool {
		return books[j].ID.String() < books[i].ID.String()
	})

	return
}

// TestPageData reads pages json data and returns as a slice
func TestPageData() (pages []*books.Page) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakePages.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &pages)

	sort.SliceStable(pages, func(i, j int) bool {
		return pages[i].CreatedAt.After(pages[j].CreatedAt)
	})
	sort.SliceStable(pages, func(i, j int) bool {
		return pages[j].ID.String() < pages[i].ID.String()
	})

	return
}

// TestAuthorData reads pages json data and returns as a slice
func TestAuthorData() (authors []*books.Author) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakeAuthors.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &authors)

	sort.SliceStable(authors, func(i, j int) bool {
		return authors[i].CreatedAt.After(authors[j].CreatedAt)
	})
	sort.SliceStable(authors, func(i, j int) bool {
		return authors[j].ID.String() < authors[i].ID.String()
	})

	return
}
