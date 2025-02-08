package aichat_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/presbrey/aichat"
)

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

func TestBasicMessageOperations(t *testing.T) {
	emptyChat := new(aichat.Chat)
	assert.Error(t, emptyChat.Delete(context.TODO(), "test-id"))
	assert.Nil(t, emptyChat.PopMessageIfRole("user"))

	t.Run("add and verify messages", func(t *testing.T) {
		chat := newTestChat()

		// Test empty state
		assert.Zero(t, chat.MessageCount(), "Expected empty chat to have 0 messages")
		assert.Nil(t, chat.LastMessage(), "Expected nil message for empty chat")
		assert.Empty(t, chat.LastMessageRole(), "Expected empty role for empty chat")

		// Test adding different message types
		testMessages := []struct {
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
			{
				addFunc: func() {
					chat.AddAssistantToolCall([]aichat.ToolCall{{
						ID:   "test-id",
						Type: "function",
						Function: aichat.Function{
							Name:      "test-tool",
							Arguments: "tool result",
						},
					}})
					chat.LastMessage().Content = "tool call"
				},
				wantRole:    "assistant",
				wantContent: "tool call",
			},
			{
				addFunc:     func() { chat.AddToolContent("test-tool", "test-id", "tool result") },
				wantRole:    "tool",
				wantContent: "tool result",
			},
			{
				addFunc:     func() { chat.AddToolContent("test-tool", "test-id", []byte("tool result")) },
				wantRole:    "tool",
				wantContent: "tool result",
			},
			{
				addFunc:     func() { chat.AddToolContent("test-tool", "test-id", map[string]interface{}{"key": "value"}) },
				wantRole:    "tool",
				wantContent: "{\"key\":\"value\"}",
			},
		}

		for i, tm := range testMessages {
			tm.addFunc()
			assert.Equal(t, i+1, chat.MessageCount(), "Unexpected message count")
			msg := chat.LastMessage()
			assertMessage(t, msg, tm.wantRole, tm.wantContent)
			assert.Equal(t, tm.wantRole, chat.LastMessageRole(), "Unexpected last message role")
		}
	})

	t.Run("message removal operations", func(t *testing.T) {
		chat := newTestChat()

		// Test removing from empty chat
		assert.Nil(t, chat.RemoveLastMessage(), "Expected nil when removing from empty chat")
		assert.Nil(t, chat.ShiftMessages(), "Expected nil when shifting from empty chat")

		// Add messages and test removal
		messages := []string{"first", "second", "third"}
		for _, msg := range messages {
			chat.AddUserContent(msg)
		}

		// Test PopMessageIfRole
		assert.Nil(t, chat.PopMessageIfRole("assistant"))
		assert.Equal(t, "third", chat.PopMessageIfRole("user").Content, "Unexpected removed message content")
		assert.Equal(t, 2, chat.MessageCount(), "Unexpected message count after removal")

		// Test RemoveLastMessage
		lastMsg := chat.RemoveLastMessage()
		assert.Equal(t, "second", lastMsg.Content, "Unexpected removed message content")
		assert.Equal(t, 1, chat.MessageCount(), "Unexpected message count after removal")

		// Test ShiftMessages
		firstMsg := chat.ShiftMessages()
		assert.Equal(t, "first", firstMsg.Content, "Unexpected shifted message content")
		assert.Equal(t, 0, chat.MessageCount(), "Unexpected message count after shift")
	})

	t.Run("message query operations", func(t *testing.T) {
		chat := newTestChat()
		assert.Nil(t, chat.LastMessageByRole("system"))

		// Add messages of different types
		chat.AddRoleContent("user", map[string]interface{}{
			"type": "text",
			"text": "Hello",
		})
		chat.AddAssistantContent(map[string]interface{}{
			"type": "image",
			"url":  "test.jpg",
		})
		chat.AddUserContent(map[string]interface{}{
			"type": "text",
			"text": "World",
		})

		// Test LastMessageByType
		textMsg := chat.LastMessageByType("text")
		assert.NotNil(t, textMsg, "Expected to find text message")
		assert.Equal(t, "World", textMsg.Content.(map[string]interface{})["text"])

		imageMsg := chat.LastMessageByType("image")
		assert.NotNil(t, imageMsg, "Expected to find image message")
		assert.Equal(t, "test.jpg", imageMsg.Content.(map[string]interface{})["url"])

		// Test LastMessageByRole
		userMsg := chat.LastMessageByRole("user")
		assert.NotNil(t, userMsg, "Expected to find user message")
		assistantMsg := chat.LastMessageByRole("assistant")
		assert.NotNil(t, assistantMsg, "Expected to find assistant message")

		// Test non-existent queries
		assert.Nil(t, chat.LastMessageByType("audio"), "Expected nil for non-existent type")
		assert.Nil(t, chat.LastMessageByRole("system"), "Expected nil for non-existent role")
	})

	t.Run("message counting", func(t *testing.T) {
		chat := newTestChat()

		messages := []struct {
			role    string
			content string
		}{
			{"user", "Hello"},
			{"assistant", "Hi"},
			{"user", "How are you?"},
			{"tool", "result"},
		}

		for i, m := range messages {
			chat.AddRoleContent(m.role, m.content)
			assert.Equal(t, i+1, chat.MessageCount(), "Total message count mismatch")
		}

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

		chat.ClearMessages()
		assert.Zero(t, chat.MessageCount(), "Expected empty chat after clear")
	})

	t.Run("direct message addition", func(t *testing.T) {
		chat := newTestChat()
		chat.SetSystemContent("You are a helpful assistant.")
		chat.SetSystemContent("You are a helpful assistant.")
		assert.Equal(t, "You are a helpful assistant.", chat.Messages[0].ContentString())

		// Test adding new message
		msg := &aichat.Message{Role: "user", Content: "test message"}
		chat.AddMessage(msg)
		assert.Equal(t, 2, len(chat.Messages), "Expected 2 messages")
		assert.Same(t, msg, chat.Messages[1], "Message not added correctly")

		chat.ShiftMessages()
		chat.SetSystemContent("You are an unhelpful assistant.")

		// Test adding same message twice
		chat.AddMessage(msg)
		assert.Equal(t, 2, len(chat.Messages), "Expected no duplicate messages")

		// Test adding nil message
		originalLen := len(chat.Messages)
		chat.AddMessage(nil)
		assert.Equal(t, originalLen, len(chat.Messages), "Expected no change when adding nil message")
	})
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
