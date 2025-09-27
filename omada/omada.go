package omada

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gotify/go-api-client/v2/models"
)

// The data structure for the JSON incoming from the Omada Controller webhook;
// this must be configured as the "Omada format" (Google chat format is not supported).
//
// Any fields not mentioned here aren't supported at this time.
type OmadaMessage struct {
	Site        string   `json:"Site"`
	Description string   `json:"description"`
	Text        []string `json:"text"`
	Controller  string   `json:"Controller"`
	Timestamp   int64    `json:"timestamp"`
	Priority    int      `json:"_priority"` // Assume TP-Link will not add this field
}

func ParseOmadaMessage(body []byte) (*OmadaMessage, error) {
	res := OmadaMessage{}

	// It can be helpful to log the incoming JSON data for debugging purposes.
	// But it's not ideal that it has the 'shardSecret' within, so wipe this
	// from the string. Also, a []byte is not a string yet.
	sanitised := string(body)
	re := regexp.MustCompile(`"shardSecret":"([^"]+)"`)
	sanitised = re.ReplaceAllString(sanitised, `"shardSecret":"****"`)

	// For now, we'll  always log. May have to make this configurable.
	log.Printf("Processing incoming message: `%v`", sanitised)

	// Parse the JSON data into the omadaMessage format
	if err := json.Unmarshal(body, &res); err != nil {
		log.Printf("Error decoding the message into the OmadaMessage format structure. Error: %v", err)
		log.Printf("The message was: %v", sanitised)
		return &res, err
	}

	// Special handling for the Omada test message which displays very little otherwise
	match, _ := regexp.MatchString("webhook test message[.] Please ignore", res.Description)

	if match {
		res.Controller = "Omada Webhook Test"
		if len(res.Text) == 0 {
			res.Text = append(res.Text, res.Description)
		}
		if res.Timestamp != 0 {
			res.Text = append(res.Text, fmt.Sprintf("Timestamp: %v", TimestampToHumanReadable(res.Timestamp)))
		} else {
			res.Text = append(res.Text, fmt.Sprintf("Timestamp: %v", time.Now()))
		}
		res.Priority = 0
		log.Println("Message is detected to be an Omada test webhook message, and processed as such")
	} else {
		// Convert timestamp to human readable string and append it to the slice of texts as another line of text
		// as the timestamp is microseconds since the unix epoch, not that well readable.
		res.Text = append(res.Text, fmt.Sprintf("Timestamp: %v", TimestampToHumanReadable(res.Timestamp)))

		// Priority 4 seems to be the lowest priority that pops off a notification, so I shall use
		// that as the default for every non-test message. It can be lowered or raised for other
		// messages in the future.
		res.Priority = 4
		// TODO: Change Priority level based on the contents of the notification
		// log.Println("Message is not a test message")
	}

	return &res, nil
}

// BuildMessageBody converts an OmadaMessage into a Go API Client MessageExternal.
func BuildMessageBody(notification *OmadaMessage) *models.MessageExternal {

	stamp := time.Unix(notification.Timestamp/1000, notification.Timestamp%1000)

	return &models.MessageExternal{
		Title:    fmt.Sprintf("%v: %v", notification.Controller, notification.Site),
		Message:  strings.Join(notification.Text, "\n"),
		Date:     time.Time(stamp),
		Priority: notification.Priority,
	}
}

func TimestampToHumanReadable(timestamp int64) string {
	seconds := timestamp / 1000
	return fmt.Sprintf("%v", time.Unix(seconds, 0))
}

// EOF
