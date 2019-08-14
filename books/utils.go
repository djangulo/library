package books

import (
	"archive/zip"
	"github.com/djangulo/library/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
func GutenbergMeta(line string) Book {
	var book Book
	re := regexp.MustCompile(`^\[([a-zA-Z\s'-]+)\,? (by|By)? ([\s.a-zA-Z]+)?\s?([\d]+)?\]$`)
	loc := re.FindAllSubmatch([]byte(line), -1)
	book.Source = NewNullString("gutenberg")

	if len(loc) > 0 {
		if title := string(loc[0][1]); title != "" {
			book.Title = strings.Trim(title, " ")
			book.Slug = Slugify(book.Title, "-")
		}
		if author := string(loc[0][3]); author != "" {
			book.Author = NewNullString(strings.Trim(author, " "))
		} else {
			book.Author = NewNullString("")
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
		book.Author = NewNullString("")
	}

	return book
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
		Unzip(dataFile, cnf.Project.Dirs.Corpora)
	} else {
		log.Printf("%s exists, skipping unzip\n", fp.Join(cnf.Project.Dirs.Corpora, "gutenberg"))

	}
}

// Unzip zipFile onto dest
func Unzip(zipFile, dest string) {
	err := os.Mkdir(dest, os.ModeDir)
	if err != nil {
		log.Printf("%s exists, skipping", dest)
	}
	// var filenames []string

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, file := range r.File {
		path := fp.Join(dest, file.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(path, fp.Clean(dest)+string(os.PathSeparator)) {
			log.Fatalf("%s: illegal file path", path)
		}

		// filenames = append(filenames, path)
		if file.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(fp.Dir(path), os.ModePerm); err != nil {
			log.Fatal(err)
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			log.Fatal(err)
		}

		rc, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Successfully unzipped %s", zipFile)
}

// MigrateDatabase noqa
func MigrateDatabase(cnf *config.Config) {
	migrations, err := migrate.New(
		"file://"+cnf.Project.Dirs.Migrations,
		cnf.Database.URL,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := migrations.Up(); err != nil {
		log.Fatal(err)
	}
}
