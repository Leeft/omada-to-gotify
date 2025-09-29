package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/leeft/omada-to-gotify/gotify"
	"github.com/leeft/omada-to-gotify/webhook"
)

var version = "development"

func main() {
	logger := log.Default()

	_, server, port, err := InitMain(logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Printf("omada-to-gotify %s server starting on port %s ...", version, port)

	logger.Fatal(http.ListenAndServe(":"+port, server))
}

func InitMain(logger *log.Logger) (gc gotify.GotifyClient, s *webhook.WebhookServer, p string, err error) {
	gotifyURL := os.Getenv("GOTIFY_URL")
	if gotifyURL == "" {
		return gotify.GotifyClient{}, nil, "", errors.New("GOTIFY_URL environment variable is required")
	}

	applicationToken := os.Getenv("GOTIFY_APP_TOKEN")
	if applicationToken == "" {
		return gotify.GotifyClient{}, nil, "", errors.New("GOTIFY_APP_TOKEN environment variable is required")
	}

	sharedSecret := os.Getenv("OMADA_SHARED_SECRET")
	if sharedSecret == "" {
		return gotify.GotifyClient{}, nil, "", errors.New("OMADA_SHARED_SECRET environment variable is required")
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

	return gotifyClient, server, port, nil
}

// EOF
