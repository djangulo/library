package config

import (
	"archive/zip"
	"fmt"
	// "github.com/golang-migrate/migrate/v4"
	// _ "github.com/golang-migrate/migrate/v4/database/postgres"
	// _ "github.com/golang-migrate/migrate/v4/source/file"
	"io"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"strings"
	"sync"
)

// Config holds configuration options for the project
type Config struct {
	Database DatabaseConfig
	Project  ProjectConfig
}

// DatabaseConfig holds configration options for the PostgreSQL database
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSL      string
	URL      string
}

// ProjectConfig holds configuration options specific to this project
type ProjectConfig struct {
	RootDir string
	Dirs    DirConfig
}

// DirConfig holds configuration options to the directories
type DirConfig struct {
	Seed       string
	Corpora    string
	Static     string
	Migrations string
	TestData   string
}

// ConnStr returns a PostgreSQL compatible connection string
func (d DatabaseConfig) ConnStr() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.SSL,
	)
}

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// func migrateDatabase(cnf *Config) {
// 	migrations, err := migrate.New(
// 		"file://"+cnf.Project.Dirs.Migrations,
// 		cnf.Database.URL,
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := migrations.Up(); err != nil {
// 		log.Fatal(err)
// 	}
// }

func acquireGutenbergData(cnf *Config) {
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
		unzip(dataFile, cnf.Project.Dirs.Corpora)
	} else {
		log.Printf("%s exists, skipping unzip\n", fp.Join(cnf.Project.Dirs.Corpora, "gutenberg"))

	}
}

func unzip(zipFile, dest string) {
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

func init() {
	cnf := Get()
	var once sync.Once
	migrationsAndSeed := func() {
		// migrateDatabase(cnf)
		acquireGutenbergData(cnf)
	}
	once.Do(migrationsAndSeed)
}

// Get returns a config object
func Get() *Config {
	rootDir := fp.Join(os.Getenv("GOPATH"), "src", "github.com", "djangulo", "library")
	dirConf := DirConfig{
		Migrations: getenv("MIGRATIONS_DIR", fp.Join(rootDir, "migrations")),
		Seed:       getenv("SEED_DATA_DIR", fp.Join(rootDir, "data", "seed_data")),
		Corpora:    getenv("CORPORA_DIR", fp.Join(rootDir, "data", "corpora")),
		Static:     getenv("HTML_TEMPLATES_DIR", fp.Join(rootDir, "static")),
		TestData:   getenv("TESTDATA_DIR", fp.Join(rootDir, "data", "testdata")),
	}
	pConf := ProjectConfig{
		RootDir: rootDir,
		Dirs:    dirConf,
	}
	dbConfig := DatabaseConfig{
		Host:     getenv("POSTGRES_HOST", "localhost"),
		Port:     getenv("POSTGRES_PORT", "5432"),
		Name:     getenv("POSTGRES_DB", "library_staging"),
		User:     getenv("POSTGRES_USER", "lygu1kqy7qqg3eccwiuh"),
		Password: getenv("POSTGRES_PASSWORD", "ECZ599EzltUH2VdS9gxiDPnkuLAs9YrUyq26JFrbbx38a9QVuKlf5kXc8KxlhZfZ"),
		SSL:      getenv("POSTGRES_SSLMODE", "disable"),
		URL:      getenv("POSTGRES_URL", "postgres://lygu1kqy7qqg3eccwiuh:ECZ599EzltUH2VdS9gxiDPnkuLAs9YrUyq26JFrbbx38a9QVuKlf5kXc8KxlhZfZ@localhost:5432/library_staging?sslmode=disable"),
	}
	return &Config{Database: dbConfig, Project: pConf}

}
