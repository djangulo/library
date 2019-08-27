package books_test

// import (
// 	// "fmt"
// 	"github.com/djangulo/library/books"
// 	"github.com/djangulo/library/books/testutils"
// 	config "github.com/djangulo/library/config/books"
// 	// "github.com/gomodule/redigo/redis"
// 	"reflect"
// 	"testing"
// )

// func TestBookCache(t *testing.T) {
// 	testBooks := testutils.TestBookData()
// 	cnf := config.CacheConfig{Host: "localhost", Port: "6379", DB: 0}
// 	cache, _ := books.NewRedisCache(cnf)

// 	t.Run("can insert/retrieve book by ID", func(t *testing.T) {
// 		want := testBooks[1]
// 		err := cache.InsertBook(want)
// 		if err != nil {
// 			t.Errorf("received error on InsertBook: %v", err)
// 		}

// 		got, err := cache.BookByID(want.ID)
// 		if err != nil {
// 			t.Errorf("received error on BookByID: %v", err)
// 		}

// 		if !reflect.DeepEqual(want, got) {
// 			t.Errorf("got --- %v --- want--- \n%v\n", got, want)
// 		}
// 	})

// 	t.Run("can insert/retrieve book by slug", func(t *testing.T) {
// 		want := testBooks[1]
// 		err := cache.InsertBook(want)
// 		if err != nil {
// 			t.Errorf("received error on InsertBook: %v", err)
// 		}

// 		got, err := cache.BookBySlug(want.Slug)
// 		if err != nil {
// 			t.Errorf("received error on BookBySlug: %v", err)
// 		}

// 		if !reflect.DeepEqual(want, got) {
// 			t.Errorf("got --- %v --- want--- \n%v\n", got, want)
// 		}
// 	})

// }
