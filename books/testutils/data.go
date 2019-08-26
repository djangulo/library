package testutils

import (
	"encoding/json"
	"github.com/djangulo/library/books"
	config "github.com/djangulo/library/config/books"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TestBookData reads books json data and returns as a slice
func TestBookData() (books []books.Book) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakeBooks.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &books)
	return
}

// TestPageData reads pages json data and returns as a slice
func TestPageData() (pages []books.Page) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakePages.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &pages)
	return
}

// TestAuthorData reads pages json data and returns as a slice
func TestAuthorData() (authors []books.Author) {
	cnf := config.Get()
	path := filepath.Join(
		cnf.Project.Dirs.TestData,
		"fakeAuthors.json",
	)
	jsonb, _ := os.Open(path)
	defer jsonb.Close()
	byteData, _ := ioutil.ReadAll(jsonb)
	json.Unmarshal(byteData, &authors)
	return
}
