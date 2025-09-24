package gotify

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
	"github.com/leeft/omada-to-gotify/omada"
)

// SendToGotify sends a message to the Gotify server using the provided URL and
// application token. It initializes a Gotify client, then attempts to send a JSON
// serialised message to the specified application.
//
// Parameters:
//   - gotifyURL: The base URL of the Gotify server (e.g., "https://gotify.example.com/")
//   - appToken: The token used to authenticate with the Gotify server as configured in the application
//   - notification: The message to send, which has been processed elsewhere
//
// If successful, it prints a confirmation message and returns nil. If there's
// an error during client creation or message sending, it logs the error and returns the error.
func SendToGotify(gotifyURL, applicationToken string, notification *omada.OmadaMessage) error {
	myURL, _ := url.Parse(gotifyURL)

	// This code sets up a new client as per the example of the gotify supplied go client;
	// it does so every time it forwards a webhook. I'm new to Go and new to how the
	// lifetime of these variables affects network connections, I can't find a "Close"
	// method for it plus it would seem the notifications are sent sporadically anyway
	// so this should be okay.

	client := gotify.NewClient(myURL, &http.Client{})

	params := message.NewCreateMessageParams()

	stamp := time.Unix(notification.Timestamp/1000, notification.Timestamp%1000)

	params.Body = &models.MessageExternal{
		Title:    fmt.Sprintf("%v: %v", notification.Controller, notification.Site),
		Message:  strings.Join(notification.Text, "\n"),
		Date:     time.Time(stamp),
		Priority: notification.Priority,
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
