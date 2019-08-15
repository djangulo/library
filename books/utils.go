package books

import (
	"archive/zip"
	"bufio"
	"fmt"
	"github.com/djangulo/library/config"
	"github.com/gofrs/uuid"
	"io"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	re := regexp.MustCompile(`^\[([a-zA-Z\s'-]+)\,? (by|By)? ([\s.a-zA-Z]+)?\s?([\d]+)?\]$`)
	loc := re.FindAllSubmatch([]byte(line), -1)
	book.Source = NewNullString("gutenberg")

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
	}

	return author, book
}

// AcquireGutenberg conditionally dowloads and seeds database with
// gutenberg data.
func AcquireGutenberg(cnf *config.Config) {
	dataFile := fp.Join(cnf.Project.RootDir, "data", "gutenberg.zip")
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {

		out, err := os.Create(dataFile)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		url := "https://raw.githubusercontent.com/nltk/nltk_data/gh-pages/packages/corpora/gutenberg.zip"
		log.Printf("Downloading Gutenberg data from %s\n", url)

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
func ParseFile(path string) (Author, Book, []Page) {
	var pages = make([]Page, 0)
	var book Book
	var author Author

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fmt.Printf("\n%+v", scanner)
	}

	return author, book, pages
}
