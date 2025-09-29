package gotify

import (
	"log"
	"net/http"
	"net/url"

	"github.com/go-openapi/runtime"
	"github.com/gotify/go-api-client/v2/auth"
	"github.com/gotify/go-api-client/v2/client"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/gotify/go-api-client/v2/gotify"
	"github.com/gotify/go-api-client/v2/models"
	"github.com/leeft/omada-to-gotify/omada"
)

// Interface used in .Send() with which the implementation can be mocked with DI
type GotifyClientMessage interface {
	CreateMessage(params *message.CreateMessageParams, authInfo runtime.ClientAuthInfoWriter) (*message.CreateMessageOK, error)
}

type GotifyClient struct {
	GotifyURL string
	Token     string
	Logger    *log.Logger
}

// Private method to turn an `OmadaMessage` into a `CreateMessageParams` that the gotify client
// code can work with.
func (msg GotifyClient) parameters(payload *omada.OmadaMessage) *message.CreateMessageParams {
	params := message.NewCreateMessageParams()
	params.Body = &models.MessageExternal{
		Title:    payload.Title(),
		Message:  payload.Body(),
		Date:     payload.Date(),
		Priority: payload.Priority(),
	}
	return params
}

// Public method to build and return a `GotifyREST` client, which is passed to the Send method.
// With this separation it's MUCH easier to mock and test the Send method.
func (msg GotifyClient) Client() *client.GotifyREST {
	myURL, _ := url.Parse(msg.GotifyURL)
	client := gotify.NewClient(myURL, &http.Client{})
	return client
}

// SendToGotify sends a message to the Gotify server using the provided client and
// application token. It attempts to encode and send a JSON serialised message to
// the specified application.
//
// If successful, it prints a confirmation message and returns nil. If there's
// an error during client creation or message sending, it logs the error and returns the error.
func (msg GotifyClient) Send(cl GotifyClientMessage, payload *omada.OmadaMessage) error {
	_, err := cl.CreateMessage(msg.parameters(payload), auth.TokenAuth(msg.Token))

	if err != nil {
		msg.Logger.Printf("Could not send message to gotify: %v", err)
		return err
	}

	msg.Logger.Println("Message sent to gotify")
	return nil
}

// EOF
