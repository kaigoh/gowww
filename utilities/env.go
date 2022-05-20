package utilities

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv attempts to populate environment variables from .env file
func LoadEnv() bool {
	if FileExists(".env") {
		err := godotenv.Overload()
		if err != nil {
			log.Fatal("Error loading .env file")
		} else {
			return true
		}
	}
	return false
}

// GetEnv attempts to fetch a value from the environment variables, returning it if found, or the passed fallback string if not
func GetEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if ok {
		return value
	}
	return fallback
}
