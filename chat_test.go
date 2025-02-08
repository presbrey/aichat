package aichat_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/presbrey/aichat"
)

func TestChat(t *testing.T) {
	ctx := context.Background()
	session := newTestChat()

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
	session := newTestChat()

	content := "The weather in Boston is sunny and 22°C."
	session.AddAssistantContent(content)

	assert.Equal(t, 1, len(session.Messages), "Expected 1 message")
	assert.Equal(t, content, session.Messages[0].ContentString(), "Message content mismatch")
}

// Helper functions to reduce duplication
func newTestChat() *aichat.Chat {
	return &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: newMockS3()}}
}

func assertMessage(t *testing.T, msg *aichat.Message, expectedRole string, expectedContent any) {
	t.Helper()
	if msg == nil {
		t.Fatal("Expected non-nil message")
	}
	assert.Equal(t, expectedRole, msg.Role, "Message role mismatch")
	assert.Equal(t, expectedContent, msg.Content, "Message content mismatch")
}

func TestLastMessage(t *testing.T) {
	chat := newTestChat()

	// Test empty chat
	assert.Nil(t, chat.LastMessage(), "Expected nil message for empty chat")

	// Test message additions
	testCases := []struct {
		addFunc     func()
		wantRole    string
		wantContent string
	}{
		{
			addFunc:     func() { chat.AddUserContent("Hello") },
			wantRole:    "user",
			wantContent: "Hello",
		},
		{
			addFunc:     func() { chat.AddAssistantContent("Hi there") },
			wantRole:    "assistant",
			wantContent: "Hi there",
		},
	}

	for _, tc := range testCases {
		tc.addFunc()
		msg := chat.LastMessage()
		assertMessage(t, msg, tc.wantRole, tc.wantContent)
	}
}

// Combine TestLastMessageRole with TestLastMessageByRole
func TestMessageRoleQueries(t *testing.T) {
	chat := newTestChat()

	// Test empty chat
	assert.Empty(t, chat.LastMessageRole(), "Expected empty role for empty chat")
	assert.Nil(t, chat.LastMessageByRole("user"), "Expected nil for empty chat")

	// Add test messages
	messages := []struct {
		role    string
		content string
	}{
		{"user", "Hello"},
		{"assistant", "Hi"},
		{"user", "How are you?"},
		{"assistant", "I'm good"},
	}

	for _, m := range messages {
		chat.AddRoleContent(m.role, m.content)
		assert.Equal(t, m.role, chat.LastMessageRole(), "Last message role mismatch")
	}

	// Test LastMessageByRole
	lastUser := chat.LastMessageByRole("user")
	assertMessage(t, lastUser, "user", "How are you?")

	lastAssistant := chat.LastMessageByRole("assistant")
	assertMessage(t, lastAssistant, "assistant", "I'm good")

	assert.Nil(t, chat.LastMessageByRole("nonexistent"), "Expected nil for non-existent role")
}

func TestMessageCounts(t *testing.T) {
	chat := newTestChat()

	// Test empty chat
	assert.Zero(t, chat.MessageCount(), "Expected empty chat to have 0 messages")

	messages := []struct {
		role    string
		content string
	}{
		{"user", "Hello"},
		{"assistant", "Hi"},
		{"user", "How are you?"},
		{"tool", "result"},
	}

	// Add messages and verify counts
	for i, m := range messages {
		chat.AddRoleContent(m.role, m.content)
		assert.Equal(t, i+1, chat.MessageCount(), "Total message count mismatch")
	}

	// Test counts by role
	expectedCounts := map[string]int{
		"user":      2,
		"assistant": 1,
		"tool":      1,
		"system":    0,
	}

	for role, expected := range expectedCounts {
		assert.Equal(t, expected, chat.MessageCountByRole(role),
			fmt.Sprintf("Message count mismatch for role %s", role))
	}

	// Test clear messages
	chat.ClearMessages()
	assert.Zero(t, chat.MessageCount(), "Expected empty chat after clear")
}

func TestAddToolContent(t *testing.T) {
	chat := newTestChat()
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
	chat := newTestChat()

	// Create a struct that will fail JSON marshaling
	badContent := make(chan int)

	err := chat.AddToolContent("test", "test-id", badContent)
	assert.Error(t, err, "Expected error when marshaling invalid content")
}

func TestUnmarshalJSONError(t *testing.T) {
	chat := newTestChat()

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

func TestLastMessageByType(t *testing.T) {
	chat := newTestChat()

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

func TestRangeByRole(t *testing.T) {
	chat := newTestChat()

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
	chat := newTestChat()

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
	chat := newTestChat()
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

func TestChat_SetSystemContent(t *testing.T) {
	updatedMessage := "updated system message"
	initialChat := newTestChat()
	initialChat.Messages = []*aichat.Message{
		{Role: "system", Content: "old system message"},
		{Role: "user", Content: "user message"},
	}
	tests := []struct {
		name          string
		initialChat   *aichat.Chat
		content       any
		expectedFirst string
		expectedCount int
		expectedValue any
	}{
		{
			name:          "empty chat adds system message",
			initialChat:   newTestChat(),
			content:       "new system message",
			expectedFirst: "system",
			expectedCount: 1,
			expectedValue: "new system message",
		},
		{
			name:          "updates existing system message",
			initialChat:   initialChat,
			content:       updatedMessage,
			expectedFirst: "system",
			expectedCount: 2,
			expectedValue: updatedMessage,
		},
		{
			name:          "updates existing system message again",
			initialChat:   initialChat,
			content:       updatedMessage,
			expectedFirst: "system",
			expectedCount: 2,
			expectedValue: updatedMessage,
		},
		{
			name: "prepends system message to existing messages",
			initialChat: &aichat.Chat{
				Messages: []*aichat.Message{
					{Role: "user", Content: "first message"},
					{Role: "assistant", Content: "response"},
				},
			},
			content:       "new system message",
			expectedFirst: "system",
			expectedCount: 3,
			expectedValue: "new system message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat := tt.initialChat
			beforeTime := chat.LastUpdated

			chat.SetSystemContent(tt.content)

			if len(chat.Messages) != tt.expectedCount {
				t.Errorf("expected %d messages, got %d", tt.expectedCount, len(chat.Messages))
			}

			if chat.Messages[0].Role != tt.expectedFirst {
				t.Errorf("expected first message role %s, got %s", tt.expectedFirst, chat.Messages[0].Role)
			}

			if chat.Messages[0].Content != tt.expectedValue {
				t.Errorf("expected content %v, got %v", tt.expectedValue, chat.Messages[0].Content)
			}

			if !strings.Contains(tt.name, "again") && !chat.LastUpdated.After(beforeTime) {
				t.Error("LastUpdated timestamp was not updated")
			}
		})
	}
}

func TestChat_ShiftMessages(t *testing.T) {
	tests := []struct {
		name          string
		initialChat   *aichat.Chat
		expectedMsg   *aichat.Message
		expectedCount int
	}{
		{
			name: "shifts message from non-empty chat",
			initialChat: func() *aichat.Chat {
				chat := newTestChat()
				chat.Messages = []*aichat.Message{
					{Role: "system", Content: "system message"},
					{Role: "user", Content: "user message"},
				}
				return chat
			}(),
			expectedMsg: &aichat.Message{
				Role:    "system",
				Content: "system message",
			},
			expectedCount: 1,
		},
		{
			name: "shifts message from single message chat",
			initialChat: &aichat.Chat{
				Messages: []*aichat.Message{
					{Role: "user", Content: "only message"},
				},
			},
			expectedMsg: &aichat.Message{
				Role:    "user",
				Content: "only message",
			},
			expectedCount: 0,
		},
		{
			name: "shifts message from empty chat",
			initialChat: &aichat.Chat{
				Messages: []*aichat.Message{},
			},
			expectedMsg:   nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat := tt.initialChat

			msg := chat.ShiftMessages()

			if msg == nil && tt.expectedMsg != nil {
				t.Error("expected non-nil message, got nil")
			} else if msg != nil && tt.expectedMsg == nil {
				t.Error("expected nil message, got non-nil")
			} else if msg != nil && tt.expectedMsg != nil {
				if msg.Role != tt.expectedMsg.Role {
					t.Errorf("expected role %s, got %s", tt.expectedMsg.Role, msg.Role)
				}
				if msg.Content != tt.expectedMsg.Content {
					t.Errorf("expected content %v, got %v", tt.expectedMsg.Content, msg.Content)
				}
			}

			if len(chat.Messages) != tt.expectedCount {
				t.Errorf("expected %d messages remaining, got %d", tt.expectedCount, len(chat.Messages))
			}
		})
	}
}

func TestChat_PopMessageIfRole(t *testing.T) {
	tests := []struct {
		name          string
		initialChat   *aichat.Chat
		role          string
		expectedMsg   *aichat.Message
		expectedCount int
		shouldUpdate  bool
	}{
		{
			name: "pops matching role from chat",
			initialChat: func() *aichat.Chat {
				chat := newTestChat()
				chat.Messages = []*aichat.Message{
					{Role: "user", Content: "first"},
					{Role: "assistant", Content: "last message"},
				}
				return chat
			}(),
			role: "assistant",
			expectedMsg: &aichat.Message{
				Role:    "assistant",
				Content: "last message",
			},
			expectedCount: 1,
			shouldUpdate:  true,
		},
		{
			name: "does not pop non-matching role",
			initialChat: func() *aichat.Chat {
				chat := newTestChat()
				chat.Messages = []*aichat.Message{
					{Role: "user", Content: "first"},
					{Role: "assistant", Content: "last message"},
				}
				return chat
			}(),
			role:          "user",
			expectedMsg:   nil,
			expectedCount: 2,
			shouldUpdate:  false,
		},
		{
			name:          "returns nil for empty chat",
			initialChat:   newTestChat(),
			role:          "user",
			expectedMsg:   nil,
			expectedCount: 0,
			shouldUpdate:  false,
		},
		{
			name: "pops last message when multiple matching roles exist",
			initialChat: func() *aichat.Chat {
				chat := newTestChat()
				chat.Messages = []*aichat.Message{
					{Role: "user", Content: "first"},
					{Role: "assistant", Content: "middle"},
					{Role: "assistant", Content: "last message"},
				}
				return chat
			}(),
			role: "assistant",
			expectedMsg: &aichat.Message{
				Role:    "assistant",
				Content: "last message",
			},
			expectedCount: 2,
			shouldUpdate:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chat := tt.initialChat
			beforeTime := chat.LastUpdated

			msg := chat.PopMessageIfRole(tt.role)

			// Check returned message
			if msg == nil && tt.expectedMsg != nil {
				t.Error("expected non-nil message, got nil")
			} else if msg != nil && tt.expectedMsg == nil {
				t.Error("expected nil message, got non-nil")
			} else if msg != nil && tt.expectedMsg != nil {
				if msg.Role != tt.expectedMsg.Role {
					t.Errorf("expected role %s, got %s", tt.expectedMsg.Role, msg.Role)
				}
				if msg.Content != tt.expectedMsg.Content {
					t.Errorf("expected content %v, got %v", tt.expectedMsg.Content, msg.Content)
				}
			}

			// Check remaining message count
			if len(chat.Messages) != tt.expectedCount {
				t.Errorf("expected %d messages remaining, got %d", tt.expectedCount, len(chat.Messages))
			}

			// Check LastUpdated timestamp
			if tt.shouldUpdate && !chat.LastUpdated.After(beforeTime) {
				t.Error("LastUpdated timestamp was not updated when it should have been")
			} else if !tt.shouldUpdate && chat.LastUpdated != beforeTime {
				t.Error("LastUpdated timestamp was updated when it should not have been")
			}
		})
	}
}
