package aichat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/presbrey/aichat"
)

func TestChat(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	session := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	// Test adding user message
	session.AddUserMessage("What is the weather like in Boston?")

	// Test adding assistant tool call
	toolCalls := []aichat.ToolCall{
		{
			ID:   "call_9pw1qnYScqvGrCH58HWCvFH6",
			Type: "function",
			Function: aichat.Function{
				Name:      "get_current_weather",
				Arguments: `{"location": "Boston, MA"}`,
			},
		},
	}
	session.AddAssistantToolCall(toolCalls)

	// Test adding tool response
	session.AddToolResponse(
		"get_current_weather",
		"call_9pw1qnYScqvGrCH58HWCvFH6",
		`{"temperature": "22", "unit": "celsius", "description": "Sunny"}`,
	)

	// Verify message count
	if len(session.Messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(session.Messages))
	}

	// Test JSON marshaling
	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal session: %v", err)
	}

	// Test JSON unmarshaling
	var newSession aichat.Chat
	if err := json.Unmarshal(data, &newSession); err != nil {
		t.Fatalf("Failed to unmarshal session: %v", err)
	}

	// Verify unmarshaled data
	if len(newSession.Messages) != 3 {
		t.Errorf("Expected 3 messages after unmarshal, got %d", len(newSession.Messages))
	}

	// Verify message content
	if newSession.Messages[0].ContentString() != "What is the weather like in Boston?" {
		t.Errorf("Unexpected user message content")
	}

	// Verify tool call
	if newSession.Messages[1].ToolCalls[0].Function.Name != "get_current_weather" {
		t.Errorf("Unexpected function name in tool call")
	}

	// Verify tool response
	if newSession.Messages[2].Name != "get_current_weather" {
		t.Errorf("Unexpected tool response name")
	}

	newSession.Delete(ctx, "test-key")
}

func TestChatWithAssistantMessage(t *testing.T) {
	s3 := newMockS3()
	session := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	content := "The weather in Boston is sunny and 22°C."
	session.AddAssistantMessage(content)

	if len(session.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(session.Messages))
	}

	if session.Messages[0].ContentString() != content {
		t.Errorf("Expected content %q, got %q", content, session.Messages[0].ContentString())
	}
}
