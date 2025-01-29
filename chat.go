package aichat

import (
	"encoding/json"
	"fmt"
	"time"
)

// Options contains configuration options for Chat sessions.
// S3 provides storage capabilities for persisting chat sessions.
type Options struct {
	S3 S3
}

// Chat represents a chat session with message history
type Chat struct {
	// ID is the unique identifier for the chat session
	ID string `json:"id,omitempty"`
	// Key is the storage key used for persistence
	Key string `json:"key,omitempty"`
	// Messages contains the chronological history of chat messages
	Messages []Message `json:"messages"`
	// Created is the timestamp when the chat session was created
	Created time.Time `json:"created"`
	// LastUpdated is the timestamp of the most recent message or modification
	LastUpdated time.Time `json:"last_updated"`
	// Meta stores arbitrary session-related data
	Meta map[string]any `json:"meta,omitempty"`
	// Options contains the configuration for these chat sessions
	Options Options `json:"-"`
}

// AddRoleContent adds a role and content to the chat
func (chat *Chat) AddRoleContent(role string, content any) {
	chat.Messages = append(chat.Messages, Message{
		Role:    role,
		Content: content,
	})
	chat.LastUpdated = time.Now()
}

// AddUserContent adds a user message to the chat
func (chat *Chat) AddUserContent(content any) {
	chat.AddRoleContent("user", content)
}

// AddAssistantContent adds an assistant message to the chat
func (chat *Chat) AddAssistantContent(content any) {
	chat.AddRoleContent("assistant", content)
}

// AddToolRawContent adds a raw content to the chat
func (chat *Chat) AddToolRawContent(name string, toolCallID string, content any) {
	chat.Messages = append(chat.Messages, Message{
		Role:       "tool",
		Name:       name,
		ToolCallID: toolCallID,
		Content:    content,
	})
	chat.LastUpdated = time.Now()
}

// AddToolContent adds a tool content to the chat
func (chat *Chat) AddToolContent(name string, toolCallID string, content any) error {
	switch contentT := content.(type) {
	case []byte:
		content = string(contentT)
	case string:
	default:
		b, err := json.Marshal(contentT)
		if err != nil {
			return err
		}
		content = string(b)
	}
	chat.AddToolRawContent(name, toolCallID, content)
	return nil
}

// AddAssistantToolCall adds an assistant message with tool calls
func (chat *Chat) AddAssistantToolCall(toolCalls []ToolCall) {
	chat.Messages = append(chat.Messages, Message{
		Role:      "assistant",
		Content:   nil,
		ToolCalls: toolCalls,
	})
	chat.LastUpdated = time.Now()
}

// LastMessage returns the last message in the chat
func (chat *Chat) LastMessage() *Message {
	if len(chat.Messages) == 0 {
		return nil
	}
	return &chat.Messages[len(chat.Messages)-1]
}

// LastMessageRole returns the role of the last message in the chat
func (chat *Chat) LastMessageRole() string {
	msg := chat.LastMessage()
	if msg == nil {
		return ""
	}
	return msg.Role
}

// Range iterates through messages
func (chat *Chat) Range(fn func(msg Message) error) error {
	for _, msg := range chat.Messages {
		if err := fn(msg); err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for the chat
func (chat *Chat) MarshalJSON() ([]byte, error) {
	type Alias Chat
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(chat),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for the chat
func (chat *Chat) UnmarshalJSON(data []byte) error {
	type Alias Chat
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(chat),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal chat: %v", err)
	}
	return nil
}
