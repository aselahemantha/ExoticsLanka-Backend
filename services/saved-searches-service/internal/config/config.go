package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8086"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://user:password@localhost:5432/exotics_lanka?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your_jwt_secret_key"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
