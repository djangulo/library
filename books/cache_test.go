package books_test

import (
	"fmt"
	"github.com/djangulo/library/books"
	"github.com/djangulo/library/books/testutils"
	config "github.com/djangulo/library/config/books"
	"github.com/gomodule/redigo/redis"
	"testing"
)

func TestInsertBook(t *testing.T) {
	testBooks := testutils.TestBookData()
	cnf := config.CacheConfig{Host: "localhost", Port: "6379"}
	cache, _ := books.NewRedisCache(cnf)

	book, _ :=cache.InsertBook(&testBooks[0])
	conn, dropConn := cache.Conn()
	values, _ := redis.ByteSlices(redis.Values(conn.Do("HGETALL", "book:"+testBooks[0].ID.String())))
	// if err != nil {
	// 	t.Errorf("%v\n", err) 
	// }

	fmt.Printf("\n\nwant: \n%+v\n\n", testBooks[0])
	fmt.Printf("\n\ngot: \n%+v\n\n", book)
}
