package config

import (
	"fmt"
	"os"
	fp "path/filepath"
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
	DataRoot   string
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

// Get returns a config object
func Get() *Config {
	rootDir := getenv("ROOT_DIR", fp.Join(os.Getenv("GOPATH"), "src", "github.com", "djangulo", "library"))
	dirConf := DirConfig{
		Migrations: getenv("MIGRATIONS_DIR", fp.Join(rootDir, "migrations")),
		DataRoot:   getenv("DATA_ROOT", fp.Join(rootDir, "data")),
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
