package books_test

import (
	// "fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/books/testutils"
	config "github.com/djangulo/library/config/books"
	// "github.com/gomodule/redigo/redis"
	"testing"
)

func TestBookCache(t *testing.T) {
	testBooks := testutils.TestBookData()
	cnf := config.Get()
	cache, err := books.NewRedisCache(cnf.Cache["test"])
	if err != nil {
		t.Errorf("could not initialize cache %v", err)
	}

	t.Run("can insert/retrieve book by ID", func(t *testing.T) {
		want := testBooks[1]
		err := cache.InsertBook(want)
		if err != nil {
			t.Errorf("received error on InsertBook: %v", err)
		}

		var got books.Book
		err = cache.BookByID(&got, &want.ID, []string{"id"})
		if err != nil {
			t.Errorf("received error on BookByID: %v", err)
		}

		testutils.AssertUUIDsEqual(t, got.ID, want.ID)
	})

	t.Run("can insert/retrieve book by slug", func(t *testing.T) {
		want := testBooks[1]
		err := cache.InsertBook(want)
		if err != nil {
			t.Errorf("received error on InsertBook: %v", err)
		}

		var got books.Book
		err = cache.BookBySlug(&got, want.Slug, []string{"id"})
		if err != nil {
			t.Errorf("received error on BookBySlug: %v", err)
		}

		testutils.AssertUUIDsEqual(t, got.ID, want.ID)
	})

}
