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
	// Metadata stores arbitrary session-related data
	Metadata map[string]any `json:"metadata,omitempty"`
	// Options contains the configuration for these chat sessions
	Options Options `json:"-"`
}

// AddRoleContent adds a role and content to the session
func (chat *Chat) AddRoleContent(role string, content any) {
	chat.Messages = append(chat.Messages, Message{
		Role:    role,
		Content: content,
	})
	chat.LastUpdated = time.Now()
}

// AddUserMessage adds a user message to the session
func (chat *Chat) AddUserMessage(content any) {
	chat.AddRoleContent("user", content)
}

// AddAssistantMessage adds an assistant message to the session
func (chat *Chat) AddAssistantMessage(content any) {
	chat.AddRoleContent("assistant", content)
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

// AddToolResponse adds a tool response message
func (chat *Chat) AddToolResponse(name, toolCallID, content string) {
	chat.Messages = append(chat.Messages, Message{
		Role:       "tool",
		Name:       name,
		ToolCallID: toolCallID,
		Content:    content,
	})
	chat.LastUpdated = time.Now()
}

// MarshalJSON implements custom JSON marshaling for the session
func (chat *Chat) MarshalJSON() ([]byte, error) {
	type Alias Chat
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(chat),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for the session
func (chat *Chat) UnmarshalJSON(data []byte) error {
	type Alias Chat
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(chat),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal session: %v", err)
	}
	return nil
}
