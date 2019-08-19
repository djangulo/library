package books

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"fmt"
	config "github.com/djangulo/library/config/books"
	"github.com/gofrs/uuid"
	_ "github.com/lib/pq" // unneded namespace
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	firstLineRe = regexp.MustCompile(`^\[([a-zA-Z\s'-]+)\,? (by|By)? ([\s.a-zA-Z]+)?\s?([\d]+)?\]$`)
)

// Slugify returns a slug-compatible version, separated by slugChar
func Slugify(str string, slugChar string) string {
	body := regexp.MustCompile(`[._\\/!?#$%, ]+`)
	bodyRemove := regexp.MustCompile(`\'+`)
	edge1 := regexp.MustCompile(`^-*`)
	edge2 := regexp.MustCompile(`-*$`)
	multiple := regexp.MustCompile(`[-]{2,}`)
	str = strings.ToLower(str)
	result := body.ReplaceAll([]byte(str), []byte(slugChar))
	result = bodyRemove.ReplaceAll(result, []byte(""))
	result = edge1.ReplaceAll(result, []byte(""))
	result = edge2.ReplaceAll(result, []byte(""))
	result = multiple.ReplaceAll(result, []byte(slugChar))
	return string(result)
}

// GutenbergMeta extract metadata from the gutenberg format
func GutenbergMeta(line string) (Author, Book) {
	var book Book
	var author Author
	loc := firstLineRe.FindAllSubmatch([]byte(line), -1)
	book.Source = NewNullString("nltk-gutenberg")
	book.ID = uuid.Must(uuid.NewV4())

	if len(loc) > 0 {
		if title := string(loc[0][1]); title != "" {
			book.Title = strings.Trim(title, " ")
			book.Slug = Slugify(book.Title, "-")
		}
		if auth := string(loc[0][3]); auth != "" {
			author.Name = strings.Trim(auth, " ")
			author.Slug = Slugify(auth, "-")
			author.ID = uuid.Must(uuid.NewV4())
			book.AuthorID = &author.ID
		} else {
			book.AuthorID = nil
		}
		if pubYear := string(loc[0][4]); pubYear != "" {
			year, err := strconv.Atoi(strings.Trim(pubYear, " "))
			if err != nil {
				book.PublicationYear = NewNullInt64(0)
			} else {
				book.PublicationYear = NewNullInt64(int64(year))
			}
		}
	} else {
		edges := regexp.MustCompile(`[\]\[]+`)
		title := edges.ReplaceAllString(line, "")
		book.Title = title
		book.Slug = Slugify(title, "-")
		book.PublicationYear = NewNullInt64(0)
		book.ID = uuid.Must(uuid.NewV4())
	}

	return author, book
}

// AcquireGutenberg conditionally dowloads and seeds database with
// gutenberg data.
func AcquireGutenberg(cnf *config.Config) {
	log.Println(cnf.Project.RootDir, cnf.Project.Dirs.Corpora, cnf.Project.Dirs.DataRoot)
	dataFile := fp.Join(cnf.Project.Dirs.DataRoot, "gutenberg.zip")
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		fmt.Println("isnotexst: ", err)

		out, err := os.Create(dataFile)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		url := "https://raw.githubusercontent.com/nltk/nltk_data/gh-pages/packages/corpora/gutenberg.zip"
		log.Printf("Downloading Gutenberg data from %s to %s\n", url, dataFile)

		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		_, err = io.Copy(out, res.Body)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("%s exists, skipping download\n", dataFile)
	}

	_, err := os.Stat(fp.Join(cnf.Project.Dirs.Corpora, "gutenberg"))
	if os.IsNotExist(err) {
		log.Printf("Unzipping %s...\n", dataFile)
		_, err := Unzip(dataFile, cnf.Project.Dirs.Corpora)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("%s exists, skipping unzip\n", fp.Join(cnf.Project.Dirs.Corpora, "gutenberg"))

	}
}

// Unzip zipFile onto dest
func Unzip(src, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := fp.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, fp.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(fp.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	log.Printf("Successfully unzipped %s", src)
	return filenames, nil
}

// ParseFile Parses a gutenberg file extracting the author, book, and pages if exist
func ParseFile(path string, linesPerPage int) (Author, Book, []Page) {
	var pages = make([]Page, 0)
	var book Book
	var author Author

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	counter := 0
	pageNumber := 1
	var body string
	var page Page
	for scanner.Scan() {
		if firstLineRe.Match([]byte(scanner.Text())) {
			author, book = GutenbergMeta(scanner.Text())
		}
		if counter < linesPerPage {
			body += scanner.Text() + "\n"
			counter++
		}
		if counter == (linesPerPage - 1) {
			page.Body = body
			page.PageNumber = pageNumber
			page.BookID = &book.ID
			page.ID = uuid.Must(uuid.NewV4())

			pages = append(pages, page)

			pageNumber++
			counter = 0
			body = ""
		}

	}

	return author, book, pages
}

// SeedFromGutenberg Seeds database with generated authors, books and pages
// from the gutenberg data. Each set of {Author, Book, []Page} is wrapped in a
// transaction, so as to prevent "pageless" books, or "bookless" pages, etc.
func SeedFromGutenberg(config *config.Config, database string) error {
	authors := make([]Author, 0)
	books := make([]Book, 0)
	pages := make([]Page, 0)
	gutenberg := fp.Join(config.Project.Dirs.Corpora, "gutenberg")
	log.Printf("Seeding database from Gutenberg data (dir: %s)...\n", gutenberg)
	db, err := sql.Open("postgres", config.Database[database].ConnStr())
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	defer db.Close()
	os.Chdir(gutenberg)
	err = fp.Walk(
		gutenberg,
		func(path string, info os.FileInfo, err error) error {
			log.Println("parsing ", info.Name(), "(", path, ")")
			if err != nil {
				log.Fatalln(err)
				return err
			}
			if info.IsDir() && info.Name() == "gutenberg" {
				log.Printf("skipping a dir %v\n", info.Name())
				return nil
			}
			if strings.Contains(info.Name(), "README") {
				log.Println("found README, skipping")
				return nil
			}
			author, book, pgs := ParseFile(path, config.Project.LinesPerPage)
			book.File = info.Name()
			for _, a := range authors {
				if author.Slug == a.Slug {
					book.AuthorID = &a.ID
					break
				} else {
					book.AuthorID = &author.ID
				}
			}
			pages = append(pages, pgs...)
			authors = append(authors, author)

			books = append(books, book)

			tx, err := db.Begin()
			if err != nil {
				return errors.Wrap(err, "could not begin transaction")
			}
			_, err = tx.Exec(
				`INSERT INTO authors (id, name, slug)
				VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;
				`,
				author.ID,
				author.Name,
				author.Slug,
			)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					log.Fatalf("seed database - authors: unable to rollback: %v", rollbackErr)
				}
				log.Fatalf("could not insert author: %v", err)
				return errors.Wrap(err, "could not insert author")
			}
			_, err = tx.Exec(
				`INSERT INTO books (
					id,
					title,
					slug,
					publication_year,
					page_count,
					file,
					author_id,
					source
				)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING;`,
				book.ID,
				book.Title,
				book.Slug,
				book.PublicationYear,
				book.PageCount,
				book.File,
				book.AuthorID,
				book.Source,
			)
			if err != nil {
				fmt.Printf("%+v\n", err)
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					log.Fatalf("seed database - books: unable to rollback: %v", rollbackErr)
				}
				log.Fatalf("could not insert book: %v", err)
				return errors.Wrap(err, "could not insert book")
			}
			for _, p := range pages {
				_, err := tx.Exec(
					`INSERT INTO pages (
						id,
						page_number,
						body,
						book_id
					)
					VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING;
					`,
					p.ID,
					p.PageNumber,
					p.Body,
					p.BookID,
				)
				if err != nil {
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						log.Fatalf("seed database - books: unable to rollback: %v", rollbackErr)
					}
					err = errors.Wrap(err, "could not insert page")
					fmt.Printf("%+v\n", p)
					log.Fatalf("could not insert page: %v", err)
					return err
				}
			}
			if commitErr := tx.Commit(); commitErr != nil {
				log.Fatalf("seed database - commit: unable to commit: %v", commitErr)
			}

			return nil

		},
	)
	if err != nil {
		return errors.Wrap(err, "could not seed database")
	}
	log.Println("Successfully seeded database!")
	return nil
}
