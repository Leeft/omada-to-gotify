package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
)

// The format of the webhook as set up by this program
type WebhookRequest struct {
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

// The data structure for the JSON incoming from the Omada Controller webhook;
// this must be configured as the "Omada format" (Google chat format is not supported).
//
// Any fields not mentioned here aren't supported at this time.
type omadaMessage struct {
	Site         string   `json:"Site"`
	Description  string   `json:"description"`
	Text         []string `json:"text"`
	Controller   string   `json:"Controller"`
	Timestamp    int64    `json:"timestamp"`
	SharedSecret string   `json:"shardSecret"` // Yeah, nice typo there TP-Link ...
}

// An actual example JSON message in "Omada format" as received through webhook.site:
// {
//   "Site": "Some Site",
//   "description": "This is a webhook message from Omada Controller",
//   "shardSecret": "xxxxxxxxxxx",
//   "text": [
//     "The controller failed to send site logs to 192.168.10.11 automatically (1 logs in total)."
//   ],
//   "Controller": "Omada Controller_ZZZZZZ",
//   "timestamp": 1758579713747
// }

var ErrForbidden = errors.New("forbidden")

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

	http.HandleFunc("/omadaToGotify", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Forward the webhook to Gotify
		err = parseAndForwardToGotify(gotifyURL, applicationToken, sharedSecret, body)
		if err != nil {
			log.Printf("Error forwarding to Gotify: %v", err)

			switch {
			case errors.Is(err, ErrForbidden):
				http.Error(w, "Forbidden", http.StatusForbidden)
			default:
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "") // or something like: "Webhook forwarded successfully"
	})

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func parseAndForwardToGotify(gotifyURL, applicationToken, sharedSecret string, body []byte) error {
	// Parse the JSON data into the omadaMessage format
	res := omadaMessage{}

	if err := json.Unmarshal(body, &res); err != nil {
		log.Printf("Error decoding the message into the omadaMessage format. Error: %v", err)
		log.Printf("The message was: %v\n", body)
		return err
	}

	if res.SharedSecret != sharedSecret {
		err := ErrForbidden
		log.Printf("Can't accept webhook request. Error: %v", err)
		return err
	}

	// Convert timestamp to human readable string and append it to the slice of texts as another line of text.
	res.Text = append(res.Text, fmt.Sprintf("Timestamp: %v", timestampToHumanReadable(res.Timestamp)))

	return sendToGotify(gotifyURL, applicationToken, res)
}

func timestampToHumanReadable(timestamp int64) string {
	seconds := timestamp / 1000
	return fmt.Sprintf("%v", time.Unix(seconds, 0))
}

// sendToGotify sends a message to the Gotify server using the provided URL and
// application token. It initializes a Gotify client, then attempts to send a JSON
// serialised message to the specified application.
//
// Parameters:
//   - gotifyURL: The base URL of the Gotify server (e.g., "https://gotify.example.com/")
//   - appToken: The token used to authenticate with the Gotify server as configured in the application
//
// If successful, it prints a confirmation message and returns nil. If there's
// an error during client creation or message sending, it logs the error and returns the error.
func sendToGotify(gotifyURL, applicationToken string, response omadaMessage) error {
	myURL, _ := url.Parse(gotifyURL)

	// This code sets up a new client as per the example of the gotify supplied go client;
	// it does so every time it forwards a webhook. I'm new to Go and new to how the
	// lifetime of these variables affects network connections, I can't find a "Close"
	// method for it plus it would seem the notifications are sent sporadically anyway
	// so this should be okay.

	client := gotify.NewClient(myURL, &http.Client{})

	params := message.NewCreateMessageParams()

	stamp := time.Unix(response.Timestamp/1000, response.Timestamp%1000)
	params.Body = &models.MessageExternal{
		Title:   fmt.Sprintf("%v: %v", response.Controller, response.Site),
		Message: strings.Join(response.Text, "\n"),
		Date:    time.Time(stamp),
		// TODO: Possible improvement to parse the message and set a priority based on what is found. Need many more example messages though.
		// Priority: 5,
	}

	_, err := client.Message.CreateMessage(params, auth.TokenAuth(applicationToken))
	if err != nil {
		log.Printf("Could not send message to gotify: %v", err)
		return err
	}

	log.Println("Message Sent!")
	return nil
}

// EOF
