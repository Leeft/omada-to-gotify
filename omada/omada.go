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

/*
 * Definitions for the types of messages this code will recognise and
 * label or treat as such.
 */

type OmadaMessageType int

const (
	UnrecognisedMessage OmadaMessageType = iota
	OmadaTestMessage
	OmadaOfflineMessage
	OmadaOnlineMessage
)

var omadaMessageTypeName = map[OmadaMessageType]string{
	UnrecognisedMessage: "unrecognised",
	OmadaTestMessage:    "test",
	OmadaOfflineMessage: "offline",
	OmadaOnlineMessage:  "online",
}

// Priorities were discussed by the Gotify author at:
// https://github.com/gotify/android/issues/18#issuecomment-437403888
var messageTypeToPriority = map[OmadaMessageType]int{
	OmadaTestMessage:    0,  // Test messages are not important
	UnrecognisedMessage: 4,  // Not specifically recognised, but still make it trigger a notification
	OmadaOfflineMessage: 10, // Going offline seems important
	OmadaOnlineMessage:  7,  // Back online is important too, not _as_ important?
}

// Functions

var shardSecretRe = regexp.MustCompile(`"shardSecret":"([^"]+)"`)

func ParseOmadaMessage(body []byte) (*OmadaMessage, error) {
	// It can be helpful to log the incoming JSON data for debugging purposes
	// but should one need to share their messages with others it's not ideal
	// that it has the 'shardSecret' within, so wipe this from the string.
	sanitised := string(body)
	sanitised = shardSecretRe.ReplaceAllString(sanitised, `"shardSecret":"****"`)
	log.Printf("Processing incoming message: `%v`", sanitised)

	// Parse the JSON body data into the omadaMessage format, populating res
	res := OmadaMessage{}
	if err := json.Unmarshal(body, &res); err != nil {
		log.Printf("Error decoding the message into the OmadaMessage format structure. Error: %v", err)
		log.Printf("The message was: %v", sanitised)
		return &res, err
	}

	// Get the type of the message by comparing it against known values
	// and the line. Then set the message priority based on that type.
	messageType := ParseTypeFromMessage(&res)
	res.Priority = messageTypeToPriority[messageType]

	log.Printf("The message is detected to be of type `%v` and is given priority %v", omadaMessageTypeName[messageType], res.Priority)

	if len(res.Text) == 0 {
		// The message doesn't have a "text" but it has a "description" so populate the
		// text field with the description
		res.Text = append(res.Text, res.Description)
	}

	switch messageType {
	case OmadaTestMessage:
		res.Controller = "Omada Webhook Test"
	}

	// If there is no timestamp, add one based on the current time.
	if res.Timestamp == 0 {
		res.Timestamp = time.Now().UnixMilli()
		log.Printf("Message is missing a timestamp; inserted current msec epoch of %v", res.Timestamp)
	}

	// Convert the timestamp to human readable string and append it to the slice
	// of texts as another line of text as the timestamp given in messages is
	// microseconds since the unix epoch, not _that_ well readable for humans.
	res.Text = append(res.Text, fmt.Sprintf("Timestamp: %v", TimestampToHumanReadable(res.Timestamp)))

	return &res, nil
}

var isATestMessage = regexp.MustCompile(`webhook test message[.] Please ignore`)
var wasOnline = regexp.MustCompile(`The online detection result of \[.+\] was online`)
var wasOffline = regexp.MustCompile(`The online detection result of \[.+\] was offline`)

func ParseTypeFromMessage(msg *OmadaMessage) OmadaMessageType {
	if isATestMessage.MatchString(msg.Description) {
		return OmadaTestMessage
	}

	for _, text := range msg.Text {
		if wasOffline.MatchString(text) {
			return OmadaOfflineMessage
		}

		if wasOnline.MatchString(text) {
			return OmadaOnlineMessage
		}
	}

	// No idea what it is, so return that type
	return UnrecognisedMessage
}

// BuildMessageBody converts an OmadaMessage into a Go API Client MessageExternal.
// The Time field is expected to be in a millisecond epoch.
func BuildMessageBody(notification *OmadaMessage) *models.MessageExternal {

	stamp := time.Unix(notification.Timestamp/1000, notification.Timestamp%1000)

	return &models.MessageExternal{
		Title:    fmt.Sprintf("%v: %v", notification.Controller, notification.Site),
		Message:  strings.Join(notification.Text, "\n"),
		Date:     time.Time(stamp),
		Priority: notification.Priority,
	}
}

// How the string is formatted exactly will depend on the OS, possibly environment
// variables and whether timezone data is available.
func TimestampToHumanReadable(timestamp int64) string {
	seconds := timestamp / 1000
	return fmt.Sprintf("%v", time.Unix(seconds, 0))
}

// EOF
