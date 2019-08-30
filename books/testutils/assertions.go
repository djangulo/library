package testutils

import (
	"github.com/djangulo/library/books"
	"github.com/gofrs/uuid"
	"net/http/httptest"
	"reflect"
	"testing"
)

// AssertBooks noqa
func AssertBooks(t *testing.T, got, want []books.Book) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertPages noqa
func AssertPages(t *testing.T, got, want []books.Page) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertBookStoreCalls noqa
func AssertBookStoreCalls(t *testing.T, store *StubStore, id string, want int) {
	t.Helper()
	got := store.BookCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertPageStoreCalls noqa
func AssertPageStoreCalls(t *testing.T, store *StubStore, id string, want int) {
	t.Helper()
	got := store.PageCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertAuthorStoreCalls noqa
func AssertAuthorStoreCalls(t *testing.T, store *StubStore, id string, want int) {
	t.Helper()
	got := store.AuthorCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertBookInsertCalls noqa
func AssertBookInsertCalls(t *testing.T, store *StubStore, key string, want int) {
	t.Helper()
	got := store.InsertBookCalls[key]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertPageInsertCalls noqa
func AssertPageInsertCalls(t *testing.T, store *StubStore, key string, want int) {
	t.Helper()
	got := store.InsertPageCalls[key]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertAuthorInsertCalls noqa
func AssertAuthorInsertCalls(t *testing.T, store *StubStore, key string, want int) {
	t.Helper()
	got := store.InsertAuthorCalls[key]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertBookCacheCalls noqa
func AssertBookCacheCalls(t *testing.T, store *StubCache, id string, want int) {
	t.Helper()
	got := store.BookCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertPageCacheCalls noqa
func AssertPageCacheCalls(t *testing.T, store *StubCache, id string, want int) {
	t.Helper()
	got := store.PageCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertAuthorCacheCalls noqa
func AssertAuthorCacheCalls(t *testing.T, store *StubCache, id string, want int) {
	t.Helper()
	got := store.AuthorCalls[id]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertCacheQueryCalls noqa
func AssertCacheQueryCalls(t *testing.T, cache *StubCache, key string, want int) {
	t.Helper()
	got := cache.QueryCalls[key]
	if got != want {
		t.Errorf("got %d want %d calls", got, want)
	}
}

// AssertStatus noqa
func AssertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	got := response.Code
	if got != want {
		t.Errorf("response status is wrong, got %d want %d", got, want)
	}
}

// AssertError noqa
func AssertError(t *testing.T, got, want error) {
	t.Helper()
	if got == nil {
		t.Error("didn't get an error but wanted one")
	}
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertNoError noqa
func AssertNoError(t *testing.T, got error) {
	t.Helper()
	if got != nil {
		t.Error("got an error but didn't want one")
	}
}

func getVal(x interface{}) reflect.Value {
	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

// AssertEqual noqa
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	switch v := got.(type) {
	case int, string, bool:
		if v != want {
			t.Errorf("got %v want %v", got, want)
		}
		return
	case books.Book, books.Page, books.Author:
		if !reflect.DeepEqual(got, want) {
			t.Errorf("\ngot \n\t%+v\nwant\n\t%+v", got, want)
		}
		return
	default:
		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
		return
	}
	// gotVal := getVal(got)
	// wantVal := getVal(want)
	// if gotVal.Kind() != wantVal.Kind() {
	// 	t.Errorf("cannot compare type %T to %T", got, want)
	// }
	// switch gotVal.Kind() {
	// case reflect.Struct, reflect.Array:
	// 	if !reflect.DeepEqual(gotVal, wantVal) {
	// 		t.Errorf("\ngot \n\t%+v\nwant\n\t%+v", gotVal, wantVal)
	// 	}
	// default:
	// 	if got != want {
	// 		t.Errorf("\ngot \n\t%v\nwant\n\t%v", got, want)
	// 	}
	// }
}

// AssertIntsEqual noqa
func AssertIntsEqual(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got '%d' want '%d'", got, want)
	}
}

// AssertUUIDsEqual noqa
func AssertUUIDsEqual(t *testing.T, got, want uuid.UUID) {
	t.Helper()
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertBooksEqual noqa
func AssertBooksEqual(t *testing.T, got, want books.Book) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertAuthorsEqual noqa
func AssertAuthorsEqual(t *testing.T, got, want books.Author) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

// AssertPagesEqual noqa
func AssertPagesEqual(t *testing.T, got, want books.Page) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
