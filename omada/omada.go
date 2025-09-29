package omada

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

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

// OmadaMessage type and methods

// The data structure for the JSON incoming from the Omada Controller webhook;
// this must be configured as the "Omada format" (Google chat format is not supported).
//
// Any incoming fields not mentioned here aren't supported at this time.
type OmadaMessage struct {
	Controller  string   `json:"Controller"`
	Site        string   `json:"Site"`
	Description string   `json:"description"`
	Text        []string `json:"text"`
	Timestamp   int64    `json:"timestamp"`
}

// The title for the message as it will be sent to Gotify. Will take the name
// of the controller and the name of the site and concatenate them as such:
//
// `Controller: Site`
//
// These fields are not always available though. An empty Site will currently
// be ignored; an empty Controller will be set to `Omada Webhook Test` if the
// message is also detected to be a test message.
func (msg OmadaMessage) Title() string {
	controller := msg.Controller

	if controller == "" && msg.Type() == OmadaTestMessage {
		controller = "Omada Webhook Test"
	}

	return fmt.Sprintf("%v: %v", controller, msg.Site)
}

// The date the message was generated, as best as possible to determine (not
// every message received gets a timestamp). Used to feed the Date: field on
// the message sent to Gotify.
func (msg OmadaMessage) Date() time.Time {
	if msg.Timestamp <= 0 {
		return time.Now()
	}
	//return time.Time(time.Unix(msg.Timestamp/1000, msg.Timestamp%1000))
	return time.Time(time.UnixMilli(msg.Timestamp))
}

func (msg OmadaMessage) Body() string {
	messages := msg.Text

	// An Omada controller initiated "test webhook" message does not get the usual
	// body text, but it does have a description. Load that into the body instead.
	if len(messages) == 0 && msg.Type() == OmadaTestMessage {
		messages = append(messages, msg.Description)
	}

	if msg.Timestamp > 0 {
		// The non-zero timestamp passed along to this message is a millisecond based epoch;
		// make it readable for humans and append it to the body text.
		t := time.UnixMilli(msg.Timestamp)
		messages = append(messages, fmt.Sprintf("Timestamp: %v", HumanReadableTimestamp(t)))
	}

	return strings.Join(messages, "\n")
}

// Get the type of the message by comparing the contents against known values.
func (msg OmadaMessage) Type() OmadaMessageType {
	return parseTypeFromMessage(&msg)
}

// Determine the priority of the message base on the detected message type.
func (msg OmadaMessage) Priority() int {
	return messageTypeToPriority[msg.Type()]
}

// Functions

var shardSecretRe = regexp.MustCompile(`"shardSecret":\s*"([^"]+)"`)

func ParseOmadaMessage(out *log.Logger, body []byte) (*OmadaMessage, error) {
	// It can be helpful to log the incoming JSON data for debugging purposes
	// but should one need to share their messages with others it's not ideal
	// that it has the 'shardSecret' within, so wipe this from the string.
	sanitised := string(body)
	sanitised = shardSecretRe.ReplaceAllString(sanitised, `"shardSecret":"****"`)

	out.Printf("Processing incoming message: `%v`", sanitised)

	// Parse the JSON body data into the omadaMessage format, populating res
	res := OmadaMessage{}
	if err := json.Unmarshal(body, &res); err != nil {
		out.Printf("Error decoding the message into the OmadaMessage format structure. Error: %v", err)
		out.Printf("The message was: %v", sanitised)
		return &res, err
	}

	out.Printf("The message is detected to be of type `%v` and is given priority %v", omadaMessageTypeName[res.Type()], res.Priority())

	return &res, nil
}

var isATestMessage = regexp.MustCompile(`webhook test message[.] Please ignore`)
var wasOnline = regexp.MustCompile(`The online detection result of \[.+\] was online`)
var wasOffline = regexp.MustCompile(`The online detection result of \[.+\] was offline`)

// Inspect the given message and return what type the message
// is expected to be based on its findings.
func parseTypeFromMessage(msg *OmadaMessage) OmadaMessageType {
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

func HumanReadableTimestamp(t time.Time) string {
	seconds := t.UnixMilli() / 1000
	return fmt.Sprintf("%v", time.Unix(seconds, 0))
}

// EOF
