package webhook

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/leeft/omada-to-gotify/gotify"
	"github.com/leeft/omada-to-gotify/omada"
)

// The format of the webhook as set up by this program
type WebhookRequest struct {
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

func WebhookServer(sharedSecret, gotifyURL, applicationToken, port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if r.Header["Access_token"] == nil || r.Header["Access_token"][0] != sharedSecret {
			http.Error(w, "Not authorized", http.StatusForbidden)
			return
		}

		omada_message, err := omada.ParseOmadaMessage(body)
		if err != nil || omada_message == nil {
			log.Printf("Error parsing Omada notification message: %v", err)
			http.Error(w, "Internal message parsing error", http.StatusInternalServerError)
			return
		}

		notifyMessage := omada.BuildMessageBody(omada_message)

		err = gotify.SendToGotify(gotifyURL, applicationToken, notifyMessage)
		if err != nil {
			log.Printf("Error sending message to Gotify: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "") // or something like: "Webhook forwarded successfully" (Omada doesn't care though)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// EOF
