package config

import (
    "os"

    "github.com/joho/godotenv"
)

// Load loads environment variables from a .env file if present.
func Load() error {
    _ = godotenv.Load()
    return nil
}

// Get returns the environment variable or fallback if empty.
func Get(key, fallback string) string {
    v := os.Getenv(key)
    if v == "" {
        return fallback
    }
    return v
}
