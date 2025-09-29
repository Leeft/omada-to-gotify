package webhook

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gotify/go-api-client/v2/client"
	"github.com/leeft/omada-to-gotify/gotify"
	"github.com/leeft/omada-to-gotify/omada"
)

type WebhookServer struct {
	GotifyClient     gotify.GotifyClient
	GotifyRESTClient *client.GotifyREST
	SharedSecret     string
	Logger           *log.Logger
}

func (ws *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if r.Header["Access_token"] == nil || r.Header["Access_token"][0] != ws.SharedSecret {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}

	omadaMessage, err := omada.ParseOmadaMessage(ws.Logger, body)
	if err != nil || omadaMessage == nil {
		ws.Logger.Printf("Error parsing Omada notification message: %v", err)
		http.Error(w, "Internal message parsing error", http.StatusInternalServerError)
		return
	}

	// This code sets up a new client as per the example of the gotify supplied go client;
	// it does so every time it forwards a webhook. I'm new to Go and new to how the
	// lifetime of these variables affects network connections, I can't find a "Close"
	// method for it plus it would seem the notifications are sent sporadically anyway
	// so this should be okay.

	err = ws.GotifyClient.Send(ws.GotifyRESTClient.Message, omadaMessage)

	if err != nil {
		ws.Logger.Printf("Error sending message to Gotify: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "") // or something like: "Webhook forwarded successfully" (Omada doesn't care though)
}

// EOF
