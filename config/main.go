package config

import (
	"os"
	fp "path/filepath"
)

var (
	// RootDir project root
	RootDir string
	// DatabaseHost host for PostgreSQL database
	DatabaseHost string
	// DatabasePort port for PostgreSQL database
	DatabasePort string
	// DatabaseName name for PostgreSQL database
	DatabaseName string
	// DatabaseUser user for PostgreSQL database
	DatabaseUser string
	// DatabasePassword password for PostgreSQL database
	DatabasePassword string
)

func getenv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func init() {
	pwd, _ := os.Getwd()
	RootDir = fp.Dir(fp.Dir(pwd))
	DatabaseHost = os.Getenv("POSTGRES_HOST")
	DatabasePort = os.Getenv("POSTGRES_PORT")
	DatabaseName = getenv("POSTGRES_DB", fp.Join(RootDir, "library.db"))
	DatabaseUser = os.Getenv("POSTGRES_USER")
	DatabasePassword = os.Getenv("POSTGRES_PASSWORD")
}
