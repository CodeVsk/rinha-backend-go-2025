package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig(path string) {
	fileName := ".env"

	if path != "" {
		fileName = path
	}

	if err := godotenv.Load(fileName); err != nil {
		log.Printf("No .env file found, using system environment variables")
	}
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
