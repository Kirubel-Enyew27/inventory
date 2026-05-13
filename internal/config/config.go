package config

import (
	"os"

	"github.com/joho/godotenv"
)

func Load() error {
	_ = godotenv.Load()
	return nil
}

func Get(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}