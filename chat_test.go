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
	session.AddToolContent(
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

func TestLastMessage(t *testing.T) {
	s3 := newMockS3()
	chat := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	// Test empty chat
	if msg := chat.LastMessage(); msg != nil {
		t.Errorf("Expected nil message for empty chat, got %v", msg)
	}

	// Add a message and test
	chat.AddUserMessage("Hello")
	msg := chat.LastMessage()
	if msg == nil {
		t.Fatal("Expected non-nil message after adding user message")
	}
	if msg.Content != "Hello" || msg.Role != "user" {
		t.Errorf("Expected last message with content 'Hello' and role 'user', got content '%s' and role '%s'", msg.Content, msg.Role)
	}

	// Add another message and test
	chat.AddAssistantMessage("Hi there")
	msg = chat.LastMessage()
	if msg == nil {
		t.Fatal("Expected non-nil message after adding assistant message")
	}
	if msg.Content != "Hi there" || msg.Role != "assistant" {
		t.Errorf("Expected last message with content 'Hi there' and role 'assistant', got content '%s' and role '%s'", msg.Content, msg.Role)
	}
}

func TestLastMessageRole(t *testing.T) {
	s3 := newMockS3()
	chat := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	// Test empty chat
	if role := chat.LastMessageRole(); role != "" {
		t.Errorf("Expected empty role for empty chat, got %q", role)
	}

	// Test user message
	chat.AddUserMessage("Hello")
	if role := chat.LastMessageRole(); role != "user" {
		t.Errorf("Expected role 'user', got %q", role)
	}

	// Test assistant message
	chat.AddAssistantMessage("Hi there")
	if role := chat.LastMessageRole(); role != "assistant" {
		t.Errorf("Expected role 'assistant', got %q", role)
	}
}

func TestAddToolContent(t *testing.T) {
	s3 := newMockS3()
	chat := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}
	toolCallID := "test_call_id"
	toolName := "test_tool"

	testCases := []struct {
		name        string
		content     any
		wantErr     bool
		wantContent string
	}{
		{
			name:        "string content",
			content:     "test string",
			wantErr:     false,
			wantContent: "test string",
		},
		{
			name:        "byte slice content",
			content:     []byte("test bytes"),
			wantErr:     false,
			wantContent: "test bytes",
		},
		{
			name: "struct content",
			content: struct {
				Key   string `json:"key"`
				Value int    `json:"value"`
			}{
				Key:   "test",
				Value: 123,
			},
			wantErr:     false,
			wantContent: `{"key":"test","value":123}`,
		},
		{
			name:        "map content",
			content:     map[string]string{"key": "value"},
			wantErr:     false,
			wantContent: `{"key":"value"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := chat.AddToolContent(toolName, toolCallID, tc.content)
			if (err != nil) != tc.wantErr {
				t.Errorf("AddToolContent() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				lastMsg := chat.LastMessage()
				if lastMsg == nil {
					t.Fatal("Expected message to be added")
				}
				if lastMsg.Role != "tool" {
					t.Errorf("Expected role 'tool', got %q", lastMsg.Role)
				}
				if lastMsg.Name != toolName {
					t.Errorf("Expected tool name %q, got %q", toolName, lastMsg.Name)
				}
				if lastMsg.ToolCallID != toolCallID {
					t.Errorf("Expected tool call ID %q, got %q", toolCallID, lastMsg.ToolCallID)
				}
				if lastMsg.Content != tc.wantContent {
					t.Errorf("Expected content %q, got %q", tc.wantContent, lastMsg.Content)
				}
			}
		})
	}
}
