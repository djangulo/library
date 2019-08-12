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
}

// ProjectConfig holds configuration options specific to this project
type ProjectConfig struct {
	RootDir string
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
	pwd, _ := os.Executable()
	rootDir := fp.Dir(fp.Dir(pwd))
	pConf := ProjectConfig{RootDir: rootDir}
	dbConfig := DatabaseConfig{
		Host:     getenv("POSTGRES_HOST", "localhost"),
		Port:     getenv("POSTGRES_PORT", "5432"),
		Name:     getenv("POSTGRES_DB", "library_staging"),
		User:     getenv("POSTGRES_USER", "lygu1kqy7qqg3eccwiuh"),
		Password: getenv("POSTGRES_PASSWORD", "ECZ599EzltUH2VdS9gxiDPnkuLAs9YrUyq26JFrbbx38a9QVuKlf5kXc8KxlhZfZ"),
		SSL:      getenv("POSTGRES_SSLMODE", "disable"),
	}
	return &Config{Database: dbConfig, Project: pConf}

}
