package booksconfig

import (
	"fmt"
	"log"
	"os"
	fp "path/filepath"
	"strconv"
)

// Config holds configuration options for the project
type Config struct {
	Database map[string]DatabaseConfig
	Cache    map[string]CacheConfig
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

// CacheConfig holds configuration options for the Redis (or other) cache
type CacheConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// ProjectConfig holds configuration options specific to this project
type ProjectConfig struct {
	Name         string
	RootDir      string
	LinesPerPage int
	Dirs         DirConfig
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

// ConnStr returns a Redis cache connection address
func (c CacheConfig) ConnStr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func getenvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	inted, err := strconv.Atoi(value)
	if err != nil {
		log.Panicln(err)
	}
	return inted
}

// Get returns a config object
func Get() *Config {
	rootDir := getenv("ROOT_DIR", fp.Join(os.Getenv("GOPATH"), "src", "github.com", "djangulo", "library"))
	dirConf := DirConfig{
		Migrations: getenv("MIGRATIONS_DIR", fp.Join(rootDir, "migrations")),
		DataRoot:   getenv("DATA_ROOT", fp.Join(rootDir, "data")),
		Seed:       getenv("SEED_DATA_DIR", fp.Join(rootDir, "data", "seed")),
		Corpora:    getenv("CORPORA_DIR", fp.Join(rootDir, "data", "corpora")),
		Static:     getenv("STATIC_DIR", fp.Join(rootDir, "static")),
		TestData:   getenv("TESTDATA_DIR", fp.Join(rootDir, "data", "testdata")),
	}
	pConf := ProjectConfig{
		Name:         "library_books",
		RootDir:      rootDir,
		LinesPerPage: 60,
		Dirs:         dirConf,
	}
	cacheConfig := CacheConfig{
		Host:     getenv("REDIS_HOST", "localhost"),
		Port:     getenv("REDIS_PORT", "6739"),
		Password: getenv("REDIS_PASSWORD", ""),
		DB:       getenvInt("REDIS_DB", 0),
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
	testDbConfig := DatabaseConfig{
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		Name:     pConf.Name + "_test_database",
		User:     dbConfig.User,
		Password: dbConfig.Password,
		SSL:      "disable",
	}
	return &Config{
		Database: map[string]DatabaseConfig{
			"main": dbConfig,
			"test": testDbConfig,
		},
		Cache: map[string]CacheConfig{
			"main": cacheConfig,
		},
		Project: pConf,
	}

}
