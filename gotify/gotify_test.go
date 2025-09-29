package gotify_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/leeft/omada-to-gotify/gotify"
	"github.com/leeft/omada-to-gotify/omada"
)

func TestBuildMessageBody(t *testing.T) {
	t.Setenv("TZ", "UTC")

	// Test case 1: Normal message

	normalMessage := &omada.OmadaMessage{
		Controller: "Test Controller",
		Site:       "Test Site",
		Text:       []string{"Test message 1", "Test message 2"},
		Timestamp:  1609459200000, // 2021-01-01 00:00:00 UTC in milliseconds
	}

	result := gotify.BuildMessageBody(normalMessage)
	expectedTitle := "Test Controller: Test Site"
	expectedMessage := fmt.Sprintf("Test message 1\nTest message 2\nTimestamp: %v", time.UnixMilli(1609459200000))
	expectedPriority := 4

	if result.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, result.Title)
	}

	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Priority != expectedPriority {
		t.Errorf("Expected priority %d, got %d", expectedPriority, result.Priority)
	}

	// Test case 2: "Test" message sent from the UI; it doesn't have a timestamp attached.
	testMessage := &omada.OmadaMessage{
		Controller:  "Omada Webhook Test",
		Site:        "Test Site",
		Description: "webhook test message. Please ignore",
		Text:        []string{"Test message 1"},
	}

	result2 := gotify.BuildMessageBody(testMessage)
	expectedTitle2 := "Omada Webhook Test: Test Site"
	expectedMessage2 := "Test message 1"
	expectedPriority2 := 0

	if result2.Title != expectedTitle2 {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle2, result2.Title)
	}
	if result2.Message != expectedMessage2 {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage2, result2.Message)
	}
	if result2.Priority != expectedPriority2 {
		t.Errorf("Expected priority %d, got %d", expectedPriority2, result2.Priority)
	}
}
