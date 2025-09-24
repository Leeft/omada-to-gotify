package main

import (
	"log"
	"os"

	"github.com/leeft/omada-to-gotify/webhook"
)

func main() {
	gotifyURL := os.Getenv("GOTIFY_URL")
	if gotifyURL == "" {
		log.Fatal("GOTIFY_URL environment variable is required")
	}

	applicationToken := os.Getenv("GOTIFY_APP_TOKEN")
	if applicationToken == "" {
		log.Fatal("GOTIFY_APP_TOKEN environment variable is required")
	}

	sharedSecret := os.Getenv("OMADA_SHARED_SECRET")
	if sharedSecret == "" {
		log.Fatal("OMADA_SHARED_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	webhook.WebhookServer(sharedSecret, gotifyURL, applicationToken, port)
}

// EOF
