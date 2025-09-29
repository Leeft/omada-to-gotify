package gotify_test

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/gotify/go-api-client/v2/client"
	"github.com/gotify/go-api-client/v2/client/message"
	"github.com/leeft/omada-to-gotify/gotify"
	"github.com/leeft/omada-to-gotify/omada"
)

// Types to mock the gotify API client out from the Send() call
// so that Send() can be tested safely here.

type GotifyClientMessageMock struct {
	Calls       int
	returnError error
}

func (mock *GotifyClientMessageMock) CreateMessage(params *message.CreateMessageParams, authInfo runtime.ClientAuthInfoWriter) (*message.CreateMessageOK, error) {
	mock.Calls += 1
	return nil, mock.returnError
}

// For the TestGotifyClient_Send test the payload doesn't really matter much
// as it's the behaviour of the integration we care about (not the underlying
// implementation of the gotify client API).
func TestGotifyClient_Send(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cl      gotify.GotifyClientMessage
		payload *omada.OmadaMessage
		calls   int
		wantErr bool
	}{
		{
			name: "valid message",
			payload: &omada.OmadaMessage{
				Site:        "Test Site",
				Description: "This is a webhook message from Omada Controller",
				Text:        []string{"The controller failed to send site logs to 192.168.10.11 automatically (1 logs in total)."},
				Controller:  "Omada Controller NNNNNN",
				Timestamp:   1758579713747,
			},
			calls:   1,
			wantErr: false,
		},
		{
			name: "valid message but inducing some artificial error",
			payload: &omada.OmadaMessage{
				Site:        "Test Site",
				Description: "This is a webhook message from Omada Controller",
				Text:        []string{"The controller failed to send site logs to 192.168.10.11 automatically (1 logs in total)."},
				Controller:  "Omada Controller NNNNNN",
				Timestamp:   1758579713747,
			},
			calls:   1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var (
			buf    bytes.Buffer
			logger = log.New(&buf, "logger: ", log.Lshortfile)
		)

		t.Run(tt.name, func(t *testing.T) {

			cl := gotify.GotifyClient{
				GotifyURL: "http://localhost:8081",
				Token:     "doesnotmatter",
				Logger:    logger,
			}

			mock := &GotifyClientMessageMock{}

			if tt.wantErr {
				mock.returnError = errors.New("test induced error")
			}

			gotErr := cl.Send(mock, tt.payload)

			if mock.Calls != tt.calls {
				t.Fatalf("Expected %d calls to have been made to the mocked method but got %d", tt.calls, mock.Calls)
			} else {
				t.Logf("%d calls were made to the mocked method", mock.Calls)
			}

			// Possible TODO: check the logging took place.

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Send() failed: %v", gotErr)
				} else {
					t.Logf("Send() correctly returned an error `%v`", gotErr)
				}

				return
			}

			if tt.wantErr {
				t.Fatal("Send() succeeded unexpectedly")
			}
		})
	}
}

func TestGotifyClient_Client(t *testing.T) {

	// I know, with one test it doesn't NEED to be a loop. But who knows
	// what the code will look like months or years from now.

	tests := []struct {
		name string // description of this test case
		want *client.GotifyREST
	}{
		{
			name: "Test client creation",
		},
	}

	for _, tt := range tests {
		var (
			buf    bytes.Buffer
			logger = log.New(&buf, "logger: ", log.Lshortfile)
		)

		gcl := gotify.GotifyClient{
			GotifyURL: "http://localhost:8081",
			Token:     "doesnotmatter",
			Logger:    logger,
		}

		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			got := gcl.Client()

			if got != nil {
				t.Logf("Got a client %v", got)
			} else {
				t.Fatal("Client() returned nil")
			}
		})
	}
}

// EOF
