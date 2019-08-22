package books_test

import (
	// "fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/books/testutils"
	config "github.com/djangulo/library/config/books"
	// "github.com/gomodule/redigo/redis"
	"reflect"
	"testing"
)

func TestInsertBook(t *testing.T) {
	testBooks := testutils.TestBookData()
	cnf := config.CacheConfig{Host: "127.0.0.1", Port: "6379"}
	cache, _ := books.NewRedisCache(cnf)

	want, err := cache.InsertBook(&testBooks[0])
	if err != nil {
		t.Errorf("received error on InsertBook: %v", err)
	}

	got, err := cache.BookByID(want.ID)
	if err != nil {
		t.Errorf("received error on BookByID: %v", err)
	}

	if !reflect.DeepEqual(*want, got) {
		t.Errorf("got --- %v --- want--- \n%v\n", got, want)
	}
}
