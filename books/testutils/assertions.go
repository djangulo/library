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
	gotVal := getVal(got)
	wantVal := getVal(want)
	if gotVal.Kind() != wantVal.Kind() {
		t.Errorf("cannot compare type %T to %T", got, want)
	}
	switch gotVal.Kind() {
	case reflect.Struct, reflect.Array:
		if !reflect.DeepEqual(gotVal, wantVal) {
			t.Errorf("got %v want %v", gotVal, wantVal)
		}
	default:
		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	}
}

// AssertUUIDsEqual noqa
func AssertUUIDsEqual(t *testing.T, got, want uuid.UUID) {
	t.Helper()
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}
