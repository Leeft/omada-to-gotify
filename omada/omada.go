package omada

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"
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

	// Parse the JSON data into the omadaMessage format
	if err := json.Unmarshal(body, &res); err != nil {
		log.Printf("Error decoding the message into the omadaMessage format. Error: %v", err)
		log.Printf("The message was: %v\n", body)
		return &res, err
	}

	// Special handling for the Omada test message which displays very little otherwise
	match, _ := regexp.MatchString("webhook test message[.] Please ignore", res.Description)

	if match {
		res.Controller = "Omada Webhook Test"
		if len(res.Text) == 0 {
			res.Text = append(res.Text, res.Description)
		}
		res.Text = append(res.Text, fmt.Sprintf("Timestamp: %v", time.Now()))
		res.Priority = 0
	} else {
		// Convert timestamp to human readable string and append it to the slice of texts as another line of text
		// as the timestamp is microseconds since the unix epoch, not that well readable.
		res.Text = append(res.Text, fmt.Sprintf("Timestamp: %v", timestampToHumanReadable(res.Timestamp)))

		// Priority 4 seems to be the lowest priority that pops off a notification, so I shall use
		// that as the default for every non-test message. It can be lowered or raised for other
		// messages in the future.
		res.Priority = 4
		// TODO: Change Priority level based on the contents of the notification
	}

	return &res, nil
}

func timestampToHumanReadable(timestamp int64) string {
	seconds := timestamp / 1000
	return fmt.Sprintf("%v", time.Unix(seconds, 0))
}

// EOF
