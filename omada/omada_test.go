package omada_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/leeft/omada-to-gotify/omada"
)

type omadaMessageMethodValues struct {
	Title    string
	Date     time.Time
	Body     string
	Priority int
	Type     omada.OmadaMessageType
}

// TestTypeOmadaMessage tests the OmadaMessage struct and with the
// right test data this implicitly tests the methods on this struct.
// So at this time no separate tests are needed for this.
func TestTypeOmadaMessage(t *testing.T) {
	// Apparently this doesn't help in Windows ... it would be very nice
	// if the tests don't only pass in WSL/Linux. So, need to work around this.
	t.Setenv("TZ", "UTC")

	tests := []struct {
		name    string
		message *omada.OmadaMessage
		want    *omadaMessageMethodValues
	}{
		// The time.UnixMilli(..) statements are to help Windows deal with
		// the problem that setting the TZ environment variable doens't work
		// there. So instead we'll compare the strings as the OS gives them
		// using this code.
		{
			name: "Normal message",
			message: &omada.OmadaMessage{
				Controller: "Test Controller",
				Site:       "Test Site",
				Text:       []string{"Test message 1"},
				Timestamp:  1640995200000,
			},
			want: &omadaMessageMethodValues{
				Title:    "Test Controller: Test Site",
				Body:     fmt.Sprintf("Test message 1\nTimestamp: %v", time.UnixMilli(1640995200000)),
				Priority: 4,
				Date:     time.UnixMilli(1640995200000),
			},
		},
		{
			name: "Controller offline message",
			message: &omada.OmadaMessage{
				Controller: "Omada_Controller",
				Site:       "Offline Site",
				Text: []string{
					"[2.5G WAN1] of [gateway:98-03-8E-3A-8D-53] is down.\r",
					"[gateway:98-03-8E-3A-8D-53]: The online detection result of [2.5G WAN1] was offline.\r",
				},
				Timestamp: 1758852904877,
			},
			want: &omadaMessageMethodValues{
				Title:    "Omada_Controller: Offline Site",
				Body:     fmt.Sprintf("[2.5G WAN1] of [gateway:98-03-8E-3A-8D-53] is down.\r\n[gateway:98-03-8E-3A-8D-53]: The online detection result of [2.5G WAN1] was offline.\r\nTimestamp: %v", omada.HumanReadableTimestamp(time.UnixMilli(1758852904877))),
				Priority: 10,
				Date:     time.UnixMilli(1758852904877),
				Type:     omada.OmadaOfflineMessage,
			},
		},
		{
			name: "Controller online message",
			message: &omada.OmadaMessage{
				Controller: "Omada_Controller 2",
				Site:       "Online Site",
				Text: []string{
					"[gateway:98-03-8E-3A-8D-53]: The online detection result of [2.5G WAN1] was online.\r",
				},
				Timestamp: 1758852934790,
			},
			want: &omadaMessageMethodValues{
				Title:    "Omada_Controller 2: Online Site",
				Body:     fmt.Sprintf("[gateway:98-03-8E-3A-8D-53]: The online detection result of [2.5G WAN1] was online.\r\nTimestamp: %v", omada.HumanReadableTimestamp(time.UnixMilli(1758852934790))),
				Priority: 7,
				Date:     time.UnixMilli(1758852934790),
				Type:     omada.OmadaOnlineMessage,
			},
		},
		{
			// Test messages sent by Omada do not have a timestamp, and do not have
			// the regular message body (just the description) so this tests that the
			// timestamp is left out, and that the body is just the description.
			// Also, the priority is kept at zero.
			name: "Test message",
			message: &omada.OmadaMessage{
				Controller:  "Test Controller",
				Description: "This is a webhook test message. Please ignore this",
			},
			want: &omadaMessageMethodValues{
				Title:    "Test Controller: ",
				Body:     "This is a webhook test message. Please ignore this",
				Priority: 0,
				Type:     omada.OmadaTestMessage,
			},
		},
		{
			// This is not an actual message I've seen, but it's interesting
			// to test the behaviour is as expected from it nonetheless.
			name: "Partial message",
			message: &omada.OmadaMessage{
				Description: "Partial message",
			},
			want: &omadaMessageMethodValues{
				Title:    ": ",
				Body:     "",
				Priority: 4,
				Type:     omada.UnrecognisedMessage,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := omadaMessageMethodValues{
				Title:    tt.message.Title(),
				Body:     tt.message.Body(),
				Date:     tt.message.Date(),
				Priority: tt.message.Priority(),
				Type:     tt.message.Type(),
			}

			if got.Title != tt.want.Title {
				t.Errorf("Test %v Title() error; got `%v` but wanted `%v`", tt.name, got.Title, tt.want.Title)
				return
			} else {
				t.Logf("Test %v Title() passed; got `%v`", tt.name, got.Title)
			}

			if got.Body != tt.want.Body {
				t.Errorf("Test %v Body() error; got `%q` but wanted `%q`", tt.name, got.Body, tt.want.Body)
				return
			} else {
				t.Logf("Test %v Body() passed; got `%v`", tt.name, got.Body)
			}

			if !tt.want.Date.IsZero() && got.Date != tt.want.Date {
				t.Errorf("Test %v Date() error; got `%v` but wanted `%v`", tt.name, got.Date, tt.want.Date)
				return
			} else {
				t.Logf("Test %v Date() passed; got `%v`", tt.name, got.Date)
			}

			if got.Priority != tt.want.Priority {
				t.Errorf("Test %v Priority() error; got `%v` but wanted `%v`", tt.name, got.Priority, tt.want.Priority)
				return
			} else {
				t.Logf("Test %v Priority() passed; got `%v`", tt.name, got.Priority)
			}

			if got.Type != tt.want.Type {
				t.Errorf("Test %v Type() error; got `%v` but wanted `%v`", tt.name, got.Type, tt.want.Type)
				return
			} else {
				t.Logf("Test %v Type() passed; got `%v`", tt.name, got.Type)
			}
		})
	}
}

func TestParseOmadaMessage(t *testing.T) {
	t.Setenv("TZ", "UTC")

	tests := []struct {
		name    string
		body    []byte
		want    *omada.OmadaMessage
		wantErr bool
	}{
		{
			name: "valid message with test description",
			body: []byte(`{
				"Site": "Test Site",
				"description": "This is a webhook message from Omada Controller",
				"shardSecret": "xxyyzz",
				"text": [
					"The controller failed to send site logs to 192.168.10.11 automatically (1 logs in total)."
				],
				"Controller": "Omada Controller NNNNNN",
				"timestamp": 1758579713747
			}`),
			want: &omada.OmadaMessage{
				Site:        "Test Site",
				Description: "This is a webhook message from Omada Controller",
				Text:        []string{"The controller failed to send site logs to 192.168.10.11 automatically (1 logs in total)."},
				Controller:  "Omada Controller NNNNNN",
				Timestamp:   1758579713747,
			},
			wantErr: false,
		},
		{
			name: "valid message with regular description",
			body: []byte(`{
				"Site": "Test Site",
				"description": "Regular alert message",
				"text": ["Alert occurred"],
				"Controller": "Test Controller",
				"timestamp": 1640995200000,
				"_priority": 3
			}`),
			want: &omada.OmadaMessage{
				Site:        "Test Site",
				Description: "Regular alert message",
				Text:        []string{"Alert occurred"},
				Controller:  "Test Controller",
				Timestamp:   1640995200000,
			},
			wantErr: false,
		},
		{
			name: "message with shardSecret",
			body: []byte(`{
				"Site": "Another Test Site",
				"description": "Very regular alert message",
				"text": [],
				"Controller": "Another Controller",
				"timestamp": 1758579713747,
				"_priority": 3,
				"shardSecret": "secret123"
			}`),
			want: &omada.OmadaMessage{
				Controller:  "Another Controller",
				Site:        "Another Test Site",
				Description: "Very regular alert message",
				Text:        []string{},
				Timestamp:   1758579713747,
			},
			wantErr: false,
		},
		{
			// timestamp added manually to the JSON plus supported in code as otherwise this gets really hard to test
			name: "Omada_test_message",
			body: []byte(`{"description":"This is a webhook test message. Please ignore this","shardSecret":"fef97b18-e440-45bc-8826-be957e4dc8f6"}`),
			want: &omada.OmadaMessage{
				Description: "This is a webhook test message. Please ignore this",
				Text:        nil,
			},
			wantErr: false,
		},

		{
			name:    "invalid JSON",
			body:    []byte(`{"invalid": json}`),
			want:    &omada.OmadaMessage{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var (
			buf    bytes.Buffer
			logger = log.New(&buf, "logger: ", log.Lshortfile)
		)

		t.Run(tt.name, func(t *testing.T) {
			got, err := omada.ParseOmadaMessage(logger, tt.body)

			logged := buf.String()

			if !strings.Contains(logged, `Processing incoming message:`) {
				t.Errorf("logger output does not contain incoming message: %s", logged)
			} else {
				t.Logf("logger output contains incoming message: %s", logged)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOmadaMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			diff := deep.Equal(got, tt.want)
			if diff != nil {
				t.Errorf("ParseOmadaMessage() test failed: %v", diff)
			}

			if !tt.wantErr && !strings.Contains(logged, `The message is detected to be of type`) {
				t.Errorf("logger output does not contain info about the detection: %s", logged)
			} else {
				t.Logf("logger output contains info about the detection: %s", logged)
			}
		})
	}
}

func TestSanitisation(t *testing.T) {
	// `shardSecret` (not my typo) is probably mostly harmless to be logged, but
	// we should still avoid logging it verbatim. Test that this happens in the
	// logging of ParseOmadaMessage.
	body := []byte(`{"description": "Sanitise this", "shardSecret":    "secret123"}`)

	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	// Create a test message to ensure the sanitization works correctly
	_, err := omada.ParseOmadaMessage(logger, body)
	if err != nil {
		t.Fatalf("ParseOmadaMessage failed: %v", err)
	}

	logged := buf.String()

	// Verify that shardSecret doesn't appear in the output
	if strings.Contains(logged, `shardSecret`) && !strings.Contains(logged, `"shardSecret":"****"`) {
		t.Fatalf("shardSecret` should not be logged; got `%v`", logged)
	} else {
		t.Logf("shardSecret` was sanitized correctly; got `%v`", logged)
	}
}

func TestParseTypeFromMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      *omada.OmadaMessage
		expected omada.OmadaMessageType
	}{
		{
			name: "Test message",
			msg: &omada.OmadaMessage{
				Description: "This is a webhook test message. Please ignore this",
				Text:        []string{"Some text"},
			},
			expected: omada.OmadaTestMessage,
		},
		{
			name: "Offline message",
			msg: &omada.OmadaMessage{
				Description: "Device offline",
				Text:        []string{"The online detection result of [2.5G WAN1] was offline"},
			},
			expected: omada.OmadaOfflineMessage,
		},
		{
			name: "Online message",
			msg: &omada.OmadaMessage{
				Description: "Device online",
				Text:        []string{"The online detection result of [2.5G WAN1] was online."},
			},
			expected: omada.OmadaOnlineMessage,
		},
		{
			name: "Unrecognised message",
			msg: &omada.OmadaMessage{
				Description: "Unknown message type",
				Text:        []string{"Some random text"},
			},
			expected: omada.UnrecognisedMessage,
		},
		{
			name: "Empty message",
			msg: &omada.OmadaMessage{
				Description: "",
				Text:        []string{},
			},
			expected: omada.UnrecognisedMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.Type()
			if result != tt.expected {
				t.Errorf("ParseTypeFromMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// EOF
