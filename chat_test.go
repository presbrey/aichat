package aichat_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/presbrey/aichat"
)

func TestChat(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	session := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	// Test adding user message
	session.AddUserContent("What is the weather like in Boston?")

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
	assert.Equal(t, 3, len(session.Messages), "Expected 3 messages")

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
	assert.Equal(t, 3, len(newSession.Messages), "Expected 3 messages after unmarshal")

	// Verify message content
	assert.Equal(t, "What is the weather like in Boston?", newSession.Messages[0].ContentString(), "Unexpected user message content")

	// Verify tool call
	assert.Equal(t, "get_current_weather", newSession.Messages[1].ToolCalls[0].Function.Name, "Unexpected function name in tool call")

	// Verify tool response
	assert.Equal(t, "get_current_weather", newSession.Messages[2].Name, "Unexpected tool response name")

	newSession.Delete(ctx, "test-key")
}

func TestChatWithAssistantMessage(t *testing.T) {
	s3 := newMockS3()
	session := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	content := "The weather in Boston is sunny and 22°C."
	session.AddAssistantContent(content)

	assert.Equal(t, 1, len(session.Messages), "Expected 1 message")
	assert.Equal(t, content, session.Messages[0].ContentString(), "Message content mismatch")
}

func TestLastMessage(t *testing.T) {
	s3 := newMockS3()
	chat := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	// Test empty chat
	assert.Nil(t, chat.LastMessage(), "Expected nil message for empty chat")

	// Add a message and test
	chat.AddUserContent("Hello")
	msg := chat.LastMessage()
	if msg == nil {
		t.Fatal("Expected non-nil message after adding user message")
	}
	assert.Equal(t, "Hello", msg.Content, "Expected last message content 'Hello'")
	assert.Equal(t, "user", msg.Role, "Expected last message role 'user'")

	// Add another message and test
	chat.AddAssistantContent("Hi there")
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
	chat.AddUserContent("Hello")
	if role := chat.LastMessageRole(); role != "user" {
		t.Errorf("Expected role 'user', got %q", role)
	}

	// Test assistant message
	chat.AddAssistantContent("Hi there")
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

func TestAddToolContentError(t *testing.T) {
	chat := &aichat.Chat{}

	// Create a struct that will fail JSON marshaling
	badContent := make(chan int)

	err := chat.AddToolContent("test", "test-id", badContent)
	assert.Error(t, err, "Expected error when marshaling invalid content")
}

func TestUnmarshalJSONError(t *testing.T) {
	chat := &aichat.Chat{}

	// Invalid JSON that will cause an unmarshal error
	invalidJSON := []byte(`{"messages": [{"role": "user", "content": invalid}]}`)

	err := chat.UnmarshalJSON(invalidJSON)
	assert.Error(t, err, "Expected error when unmarshaling invalid JSON")
}

func TestContentPartsError(t *testing.T) {
	msg := &aichat.Message{
		Role: "user",
		// Content that will fail JSON marshaling
		Content: []interface{}{make(chan int)},
	}

	parts, err := msg.ContentParts()
	if err == nil {
		t.Error("Expected error when processing invalid content parts, got nil")
	}
	if parts != nil {
		t.Error("Expected nil parts when error occurs")
	}
}

func TestLastMessageByRole(t *testing.T) {
	// Test empty chat
	chat := &aichat.Chat{}
	if msg := chat.LastMessageByRole("user"); msg != nil {
		t.Error("Expected nil for empty chat")
	}

	// Add messages with different roles
	chat.Messages = []*aichat.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "How are you?"},
		{Role: "assistant", Content: "I'm good"},
	}

	// Test finding last user message
	lastUser := chat.LastMessageByRole("user")
	if lastUser == nil {
		t.Error("Expected to find last user message")
	}
	if lastUser.Content != "How are you?" {
		t.Errorf("Expected 'How are you?', got %q", lastUser.Content)
	}

	// Test finding last assistant message
	lastAssistant := chat.LastMessageByRole("assistant")
	if lastAssistant == nil {
		t.Error("Expected to find last assistant message")
	}
	if lastAssistant.Content != "I'm good" {
		t.Errorf("Expected 'I'm good', got %q", lastAssistant.Content)
	}

	// Test non-existent role
	if msg := chat.LastMessageByRole("nonexistent"); msg != nil {
		t.Error("Expected nil for non-existent role")
	}
}

func TestLastMessageByType(t *testing.T) {
	chat := new(aichat.Chat)

	// Add messages with different content types
	chat.AddRoleContent("user", map[string]interface{}{
		"type": "text",
		"text": "Hello",
	})
	chat.AddRoleContent("assistant", map[string]interface{}{
		"type": "image",
		"url":  "test.jpg",
	})
	chat.AddRoleContent("user", map[string]interface{}{
		"type": "text",
		"text": "World",
	})

	// Test finding last message of each type
	textMsg := chat.LastMessageByType("text")
	if textMsg == nil || textMsg.Content.(map[string]interface{})["text"] != "World" {
		t.Error("Expected last text message to be 'World'")
	}

	imageMsg := chat.LastMessageByType("image")
	if imageMsg == nil || imageMsg.Content.(map[string]interface{})["url"] != "test.jpg" {
		t.Error("Expected last image message to have URL 'test.jpg'")
	}

	// Test non-existent type
	audioMsg := chat.LastMessageByType("audio")
	if audioMsg != nil {
		t.Error("Expected no message for non-existent type")
	}
}

func TestMessageCount(t *testing.T) {
	chat := new(aichat.Chat)

	if chat.MessageCount() != 0 {
		t.Error("Expected empty chat to have 0 messages")
	}

	chat.AddUserContent("Hello")
	chat.AddAssistantContent("Hi")
	chat.AddUserContent("How are you?")

	if chat.MessageCount() != 3 {
		t.Errorf("Expected 3 messages, got %d", chat.MessageCount())
	}

	chat.ClearMessages()

	if chat.MessageCount() != 0 {
		t.Error("Expected empty chat to have 0 messages")
	}
}

func TestMessageCountByRole(t *testing.T) {
	chat := new(aichat.Chat)

	chat.AddUserContent("Hello")
	chat.AddAssistantContent("Hi")
	chat.AddUserContent("How are you?")
	chat.AddToolRawContent("test-tool", "123", "result")

	tests := []struct {
		role     string
		expected int
	}{
		{"user", 2},
		{"assistant", 1},
		{"tool", 1},
		{"system", 0},
	}

	for _, test := range tests {
		count := chat.MessageCountByRole(test.role)
		if count != test.expected {
			t.Errorf("Expected %d messages for role '%s', got %d", test.expected, test.role, count)
		}
	}
}

func TestRangeByRole(t *testing.T) {
	chat := new(aichat.Chat)

	// Add test messages
	chat.AddUserContent("U1")
	chat.AddAssistantContent("A1")
	chat.AddUserContent("U2")
	chat.AddAssistantContent("A2")

	// Test ranging over user messages
	userMsgs := []string{}
	err := chat.RangeByRole("user", func(msg *aichat.Message) error {
		content, ok := msg.Content.(string)
		if !ok {
			return fmt.Errorf("expected string content")
		}
		userMsgs = append(userMsgs, content)
		return nil
	})

	assert.NoError(t, err, "Unexpected error during RangeByRole")
	assert.Equal(t, []string{"U1", "U2"}, userMsgs, "User messages do not match")

	// Test ranging with error
	expectedErr := errors.New("test error")
	err = chat.RangeByRole("assistant", func(msg *aichat.Message) error {
		return expectedErr
	})
	if err != expectedErr {
		t.Error("Expected error to be propagated")
	}
}

func TestRemoveLastMessage(t *testing.T) {
	chat := new(aichat.Chat)

	// Test removing from empty chat
	if msg := chat.RemoveLastMessage(); msg != nil {
		t.Error("Expected nil when removing from empty chat")
	}

	// Add and remove messages
	chat.AddUserContent("First")
	chat.AddAssistantContent("Second")
	chat.AddUserContent("Third")

	initialCount := chat.MessageCount()
	lastMsg := chat.RemoveLastMessage()

	if lastMsg == nil || lastMsg.Content != "Third" {
		t.Error("Expected last message content to be 'Third'")
	}
	if chat.MessageCount() != initialCount-1 {
		t.Error("Expected message count to decrease by 1")
	}
	if last := chat.LastMessage(); last == nil || last.Content != "Second" {
		t.Error("Expected new last message to be 'Second'")
	}
}

func TestAddMessage(t *testing.T) {
	chat := &aichat.Chat{}
	msg := &aichat.Message{Role: "user", Content: "test message"}

	t.Run("add new message", func(t *testing.T) {
		chat.AddMessage(msg)
		if len(chat.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(chat.Messages))
		}
		if chat.Messages[0] != msg {
			t.Error("message not added correctly")
		}
	})

	t.Run("add same message twice", func(t *testing.T) {
		chat.AddMessage(msg)
		if len(chat.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(chat.Messages))
		}
	})

	t.Run("add nil message", func(t *testing.T) {
		originalLen := len(chat.Messages)
		chat.AddMessage(nil)
		if len(chat.Messages) != originalLen {
			t.Errorf("expected %d messages, got %d", originalLen, len(chat.Messages))
		}
	})
}
