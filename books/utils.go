package books

import (
	"archive/zip"
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	config "github.com/djangulo/library/config/books"
	"github.com/gofrs/uuid"
	_ "github.com/lib/pq" // unneded namespace
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	firstLineRegex = regexp.MustCompile(`^\[.*(\] ?)$`)
	metadataRegex  = regexp.MustCompile(`^\[([a-zA-Z\s'-]+)\,? (by|By)? ([\s.a-zA-Z]+)?\s?([\d]+)?(\] ?)$`)
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
func GutenbergMeta(line string, assignID bool) (Author, Book) {
	var book Book
	var author Author
	loc := metadataRegex.FindAllSubmatch([]byte(line), -1)
	book.Source = NewNullString("nltk-gutenberg")
	if assignID {
		book.ID = uuid.Must(uuid.NewV4())
	}

	if len(loc) > 0 {
		if title := string(loc[0][1]); title != "" {
			book.Title = strings.Trim(title, " ")
			book.Slug = Slugify(book.Title, "-")
		}
		if auth := string(loc[0][3]); auth != "" {
			author.Name = strings.Trim(auth, " ")
			author.Slug = Slugify(auth, "-")
			author.ID = uuid.Must(uuid.NewV4())
			book.AuthorID = NewNullUUID(author.ID.String())
		} else {
			book.AuthorID = NewNullUUID("")
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
		edges := regexp.MustCompile(`[\[\]]+`)
		title := edges.ReplaceAllString(line, "")
		book.Title = title
		book.Slug = Slugify(title, "-")
		book.PublicationYear = NewNullInt64(0)
		book.AuthorID = NewNullUUID("")
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
	line := 1
	for scanner.Scan() {
		if line == 1 {
			author, book = GutenbergMeta(scanner.Text(), true)
		}
		line++
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
	if len(pages) > 0 {
		book.PageCount = len(pages)
	}

	return author, book, pages
}

// SaveJSON Parses gutenberg data into json files, which the app uses to seed
// the database on initialization.
func SaveJSON(config *config.Config) error {
	authors := make([]Author, 0)
	books := make([]Book, 0)
	pages := make([]Page, 0)
	gutenbergSeed := fp.Join(config.Project.Dirs.Seed, "gutenberg")
	if _, err := os.Stat(gutenbergSeed); os.IsNotExist(err) {
		gutenberg := fp.Join(config.Project.Dirs.Corpora, "gutenberg")
		log.Printf("Reading data from database from Gutenberg data (dir: %s)...\n", gutenberg)
		err := os.MkdirAll(gutenbergSeed, os.ModeDir)
		if err != nil {
			log.Fatalf("error creating directory %v: %v", gutenbergSeed, err)
		}
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
				book.File = NewNullString(info.Name())
				for _, a := range authors {
					if author.Slug == a.Slug {
						book.AuthorID = NewNullUUID(a.ID.String())
						break
					} else {
						book.AuthorID = NewNullUUID(author.ID.String())
					}
				}
				pages = append(pages, pgs...)
				authorExists := false
				for _, a := range authors {
					if a.Slug == author.Slug {
						authorExists = true
					}
				}
				if !authorExists && author.Slug != "" {
					authors = append(authors, author)
				}
				books = append(books, book)
				return nil
			},
		)
		if err != nil {
			return errors.Wrap(err, "could not read files")
		}

		out, err := os.Create(fp.Join(gutenbergSeed, "books.json"))
		if err != nil {
			log.Fatalln(err)
		}
		w := bufio.NewWriter(out)
		_, err = w.Write([]byte("["))
		if err != nil {
			log.Fatalln(err)
		}
		for i, b := range books {
			f, _ := json.MarshalIndent(b, "", "  ")
			w.Write(f)
			if i != len(books)-1 {
				w.Write([]byte(","))
			}
		}
		_, err = w.Write([]byte("]"))
		w.Flush()
		out.Close()

		out, err = os.Create(fp.Join(gutenbergSeed, "authors.json"))
		if err != nil {
			log.Fatalln(err)
		}
		w = bufio.NewWriter(out)
		_, err = w.Write([]byte("["))
		if err != nil {
			log.Fatalln(err)
		}
		for i, b := range authors {
			f, _ := json.MarshalIndent(b, "", "  ")
			w.Write(f)
			if i != len(authors)-1 {
				w.Write([]byte(","))
			}
		}
		_, err = w.Write([]byte("]"))
		w.Flush()
		out.Close()

		out, err = os.Create(fp.Join(gutenbergSeed, "pages.json"))
		if err != nil {
			log.Fatalln(err)
		}
		w = bufio.NewWriter(out)
		_, err = w.Write([]byte("["))
		if err != nil {
			log.Fatalln(err)
		}
		for i, b := range pages {
			f, _ := json.MarshalIndent(b, "", "  ")
			w.Write(f)
			if i != len(pages)-1 {
				w.Write([]byte(","))
			}
		}
		_, err = w.Write([]byte("]"))
		w.Flush()
		out.Close()

		log.Println("Successfully created JSON files")
	} else {
		log.Printf("%v exists, skipping...\n", config.Project.Dirs.Seed)
	}
	return nil
}

// SeedFromGutenberg Seeds database with generated authors, books and pages
// from the gutenberg data.
func SeedFromGutenberg(config *config.Config, database string) error {
	gutenbergSeed := fp.Join(config.Project.Dirs.Seed, "gutenberg")
	if _, err := os.Stat(gutenbergSeed); os.IsNotExist(err) {
		return errors.Wrap(err, "Seed directory not found, create json files with `SaveJSON` first.")
	}
	var authors []Author
	var books []Book
	var pages []Page

	log.Printf("Seeding database from Gutenberg data (dir: %s)...\n", gutenbergSeed)
	db, err := sql.Open("postgres", config.Database[database].ConnStr())
	if err != nil {
		log.Fatalf("failed to connect database %v", err)
	}
	defer db.Close()

	authorsJSON, err := os.Open(fp.Join(gutenbergSeed, "authors.json"))
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				"could not open %s",
				fp.Join(gutenbergSeed, "authors.json"),
			),
		)
	}
	byteAuthors, err := ioutil.ReadAll(authorsJSON)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				"could not read %s",
				fp.Join(gutenbergSeed, "authors.json"),
			),
		)
	}

	err = json.Unmarshal(byteAuthors, &authors)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal authors.json")
	}
	// insert all authors first
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "could not begin transaction")
	}
	_, err = tx.Exec(`SET CLIENT_ENCODING TO 'LATIN2';`)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(
				err,
				"seed database - authors, set encoding: unable to rollback",
			)
		}
		return errors.Wrap(err, "unable to set encoding")
	}
	for _, a := range authors {
		_, err = tx.Exec(
			`INSERT INTO authors (id, name, slug)
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;
			`,
			a.ID,
			a.Name,
			a.Slug,
		)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return errors.Wrap(
					err,
					"seed database - authors: unable to rollback",
				)
			}
			return errors.Wrap(err, fmt.Sprintf("could not insert author %v", a))
		}
	}
	_, err = tx.Exec(`RESET CLIENT_ENCODING;`)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return errors.Wrap(
				err,
				"seed database - reset encoding, unable to rollback",
			)
		}
		return errors.Wrap(
			err,
			"seed database - reset encoding, unable to reset",
		)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return errors.Wrap(err, "unable to commit")
	}

	// insert books and its pages, each on a transaction
	booksJSON, err := os.Open(fp.Join(gutenbergSeed, "books.json"))
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				"could not open %s",
				fp.Join(gutenbergSeed, "books.json"),
			),
		)
	}
	byteBooks, err := ioutil.ReadAll(booksJSON)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				"could not read %s",
				fp.Join(gutenbergSeed, "books.json"),
			),
		)
	}
	err = json.Unmarshal(byteBooks, &books)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal books.json")
	}

	pagesJSON, err := os.Open(fp.Join(gutenbergSeed, "pages.json"))
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				"could not open %s",
				fp.Join(gutenbergSeed, "pages.json"),
			),
		)
	}
	bytePages, err := ioutil.ReadAll(pagesJSON)
	if err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				"could not read %s",
				fp.Join(gutenbergSeed, "pages.json"),
			),
		)
	}
	err = json.Unmarshal(bytePages, &pages)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal pages.json")
	}

	for _, b := range books {
		tx, err := db.Begin()
		if err != nil {
			return errors.Wrap(err, "could not begin transaction")
		}

		_, err = tx.Exec(`SET CLIENT_ENCODING TO 'LATIN2';`)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return errors.Wrap(
					err,
					"seed database - set encoding: unable to rollback",
				)
			}
			return errors.Wrap(err, "unable to set encoding")
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
			b.ID,
			b.Title,
			b.Slug,
			b.PublicationYear,
			b.PageCount,
			b.File,
			b.AuthorID,
			b.Source,
		)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return errors.Wrap(
					err,
					fmt.Sprintf(
						"seed database - books, unable to rollback on book: %+v",
						b,
					),
				)
			}
			return errors.Wrap(
				err,
				fmt.Sprintf(
					"seed database - books, could not insert book: %+v",
					b,
				),
			)
		}

		for _, p := range pages {
			if p.BookID.String() == b.ID.String() {
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
					rollbackErr := tx.Rollback()
					if rollbackErr != nil {
						return errors.Wrap(
							err,
							fmt.Sprintf(
								"seed database - pages, unable to rollback on page %+v of book %+v",
								p,
								b,
							),
						)
					}
					return errors.Wrap(
						err,
						fmt.Sprintf(
							"seed database - pages, could not insert page %+v of book %+v",
							p,
							b,
						),
					)
				}
			}
		}
		_, err = tx.Exec(`RESET CLIENT_ENCODING`)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return errors.Wrap(
					err,
					"seed database - reset encoding, unable to rollback",
				)
			}
			return errors.Wrap(
				err,
				"seed database - reset encoding, unable to reset",
			)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return errors.Wrap(err, "unable to commit")
		}
	}
	log.Println("Successfully seeded database!")
	return nil
}

func mustOpen(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("unable to open file %s: %v", path, err)
	}
	return file
}

func mustRead(file *os.File) []byte {
	byteData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("unable to read file %v: %v", file, err)
	}
	return byteData
}

func mustOpenAndRead(path string) []byte {
	file := mustOpen(path)
	byteData := mustRead(file)
	return byteData
}
