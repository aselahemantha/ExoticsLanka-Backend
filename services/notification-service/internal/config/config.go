package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	JWTSecret               string
	SendGridAPIKey          string
	EmailFrom               string
	EmailFromName           string
	TwilioAccountSID        string
	TwilioAuthToken         string
	TwilioPhoneNumber       string
	FirebaseCredentialsFile string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	return &Config{
		Port:                    getEnv("PORT", "8092"),
		DatabaseURL:             getEnv("DATABASE_URL", ""),
		JWTSecret:               getEnv("JWT_SECRET", ""),
		SendGridAPIKey:          getEnv("SENDGRID_API_KEY", ""),
		EmailFrom:               getEnv("EMAIL_FROM", "noreply@exotics.lk"),
		EmailFromName:           getEnv("EMAIL_FROM_NAME", "Exotics Lanka"),
		TwilioAccountSID:        getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:         getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber:       getEnv("TWILIO_PHONE_NUMBER", ""),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
