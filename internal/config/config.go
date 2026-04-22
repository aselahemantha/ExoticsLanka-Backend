package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	JWTRefreshSecret string
	CloudinaryURL    string

	// SendGrid
	SendGridAPIKey string
	FromEmail      string
	FromName       string

	// Twilio
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, looking at env variables")
	}

	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/exotics_lanka?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:        getEnv("JWT_SECRET", "super-secret-key"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "super-secret-refresh-key"),
		CloudinaryURL:    getEnv("CLOUDINARY_URL", ""),

		// SendGrid
		SendGridAPIKey: getEnv("SENDGRID_API_KEY", ""),
		FromEmail:      getEnv("FROM_EMAIL", "noreply@exoticslanka.lk"),
		FromName:       getEnv("FROM_NAME", "Exotics Lanka"),

		// Twilio
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
