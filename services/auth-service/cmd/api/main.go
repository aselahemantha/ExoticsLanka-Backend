package main

import (
	"log"
	"net/http"

	"github.com/exoticsLanka/auth-service/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Auth Service is running"))
	})

	log.Printf("Starting Auth Service on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
