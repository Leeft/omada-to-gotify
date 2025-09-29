package gotify

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
	"github.com/leeft/omada-to-gotify/omada"
)

// BuildMessageBody takes an OmadaMessage and turns it into a Gotify API Client MessageExternal.
func BuildMessageBody(msg *omada.OmadaMessage) *models.MessageExternal {
	return &models.MessageExternal{
		Title:    msg.Title(),
		Message:  msg.Body(),
		Date:     msg.Date(),
		Priority: msg.Priority(),
	}
}

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
func SendToGotify(gotifyURL, applicationToken string, notifyMessage *models.MessageExternal) error {

	myURL, _ := url.Parse(gotifyURL)

	// This code sets up a new client as per the example of the gotify supplied go client;
	// it does so every time it forwards a webhook. I'm new to Go and new to how the
	// lifetime of these variables affects network connections, I can't find a "Close"
	// method for it plus it would seem the notifications are sent sporadically anyway
	// so this should be okay.

	client := gotify.NewClient(myURL, &http.Client{})

	params := message.NewCreateMessageParams()

	params.Body = notifyMessage

	_, err := client.Message.CreateMessage(params, auth.TokenAuth(applicationToken))
	if err != nil {
		log.Printf("Could not send message to gotify: %v", err)
		return err
	}

	log.Println("Message sent to gotify")
	return nil
}

// EOF
