package main

import (
	"log"
	"net/http"
	"os"

	"github.com/leeft/omada-to-gotify/gotify"
	"github.com/leeft/omada-to-gotify/webhook"
)

var version = "development"

func main() {
	logger := log.Default()

	gotifyURL := os.Getenv("GOTIFY_URL")
	if gotifyURL == "" {
		logger.Fatal("GOTIFY_URL environment variable is required")
	}

	applicationToken := os.Getenv("GOTIFY_APP_TOKEN")
	if applicationToken == "" {
		logger.Fatal("GOTIFY_APP_TOKEN environment variable is required")
	}

	sharedSecret := os.Getenv("OMADA_SHARED_SECRET")
	if sharedSecret == "" {
		logger.Fatal("OMADA_SHARED_SECRET environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gotifyClient := gotify.GotifyClient{
		GotifyURL: gotifyURL,
		Token:     applicationToken,
		Logger:    logger,
	}

	server := &webhook.WebhookServer{
		GotifyClient:        gotifyClient,
		GotifyClientMessage: gotifyClient.Client().Message,
		SharedSecret:        sharedSecret,
		Logger:              logger,
	}

	logger.Printf("omada-to-gotify %s server starting on port %s ...", version, port)

	logger.Fatal(http.ListenAndServe(":"+port, server))
}

// EOF
