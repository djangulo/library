package books_test

import (
	"encoding/json"
	"github.com/djangulo/library/books"
	"testing"
	// "github.com/djangulo/library/books/testutils"
	config "github.com/djangulo/library/config/books"
)

func BenchmarkJSONUnmarshalling(b *testing.B) {
	cnf := config.Get()
	tBooks, _ := books.BookSeedData(cnf)
	pages, _ := books.PageSeedData(cnf)
	authors, _ := books.AuthorSeedData(cnf)

	b.Run("single", func(b *testing.B) {
		b.Run("author", func(b *testing.B) {
			blob, _ := json.Marshal(authors[0])
			for i := 0; i < b.N; i++ {
				var x books.Author
				json.Unmarshal(blob, &x)
			}
		})
		b.Run("book", func(b *testing.B) {
			blob, _ := json.Marshal(tBooks[0])
			for i := 0; i < b.N; i++ {
				var x books.Book
				json.Unmarshal(blob, &x)
			}
		})
		b.Run("page", func(b *testing.B) {
			blob, _ := json.Marshal(pages[0])
			for i := 0; i < b.N; i++ {
				var x books.Page
				json.Unmarshal(blob, &x)
			}
		})
	})
	b.Run("100", func(b *testing.B) {
		b.Run("author", func(b *testing.B) {
			blob, _ := json.Marshal(authors[:])
			for i := 0; i < b.N; i++ {
				var x []books.Author
				json.Unmarshal(blob, &x)
			}
		})
		b.Run("book", func(b *testing.B) {
			blob, _ := json.Marshal(tBooks[:])
			for i := 0; i < b.N; i++ {
				var x []books.Book
				json.Unmarshal(blob, &x)
			}
		})
		b.Run("page", func(b *testing.B) {
			blob, _ := json.Marshal(pages[:100])
			for i := 0; i < b.N; i++ {
				var x books.Page
				json.Unmarshal(blob, &x)
			}
		})
	})
	b.Run("1000", func(b *testing.B) {
		b.Run("page", func(b *testing.B) {
			blob, _ := json.Marshal(pages[:1000])
			for i := 0; i < b.N; i++ {
				var x books.Page
				json.Unmarshal(blob, &x)
			}
		})
	})

}

func BenchmarkJSONMarshalling(b *testing.B) {
	cnf := config.Get()
	tBooks, _ := books.BookSeedData(cnf)
	pages, _ := books.PageSeedData(cnf)
	authors, _ := books.AuthorSeedData(cnf)

	b.Run("single", func(b *testing.B) {
		b.Run("author", func(b *testing.B) {
			x := authors[0]
			for i := 0; i < b.N; i++ {
				json.Marshal(x)
			}
		})
		b.Run("book", func(b *testing.B) {
			x := tBooks[0]
			for i := 0; i < b.N; i++ {
				json.Marshal(x)
			}
		})
		b.Run("page", func(b *testing.B) {
			x := pages[0]
			for i := 0; i < b.N; i++ {
				json.Marshal(x)
			}
		})
	})
	b.Run("100", func(b *testing.B) {
		b.Run("author", func(b *testing.B) {
			x := authors
			for i := 0; i < b.N; i++ {
				json.Marshal(x)
			}
		})
		b.Run("book", func(b *testing.B) {
			x := tBooks
			for i := 0; i < b.N; i++ {
				json.Marshal(x)
			}
		})
		b.Run("page", func(b *testing.B) {
			x := pages[:100]
			for i := 0; i < b.N; i++ {
				json.Marshal(x)
			}
		})
	})
	b.Run("1000", func(b *testing.B) {
		b.Run("page", func(b *testing.B) {
			x := pages[:1000]
			for i := 0; i < b.N; i++ {
				json.Marshal(x)
			}
		})
	})
}
