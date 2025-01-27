package aichat

import (
	"encoding/json"
	"fmt"
	"time"
)

// Options contains configuration options for Chat
type Options struct {
	S3 S3
}

// Chat represents a chat session with message history
type Chat struct {
	ID          string         `json:"id,omitempty"`
	Key         string         `json:"key,omitempty"`
	Messages    []Message      `json:"messages"`
	Created     time.Time      `json:"created"`
	LastUpdated time.Time      `json:"last_updated"`
	Metadata    map[string]any `json:"metadata,omitempty"`

	Options Options `json:"-"`
}

// AddRoleContent adds a role and content to the session
func (s *Chat) AddRoleContent(role string, content any) {
	s.Messages = append(s.Messages, Message{
		Role:    role,
		Content: content,
	})
	s.LastUpdated = time.Now()
}

// AddUserMessage adds a user message to the session
func (s *Chat) AddUserMessage(content any) {
	s.AddRoleContent("user", content)
}

// AddAssistantMessage adds an assistant message to the session
func (s *Chat) AddAssistantMessage(content any) {
	s.AddRoleContent("assistant", content)
}

// AddAssistantToolCall adds an assistant message with tool calls
func (s *Chat) AddAssistantToolCall(toolCalls []ToolCall) {
	s.Messages = append(s.Messages, Message{
		Role:      "assistant",
		Content:   nil,
		ToolCalls: toolCalls,
	})
	s.LastUpdated = time.Now()
}

// AddToolResponse adds a tool response message
func (s *Chat) AddToolResponse(name, toolCallID, content string) {
	s.Messages = append(s.Messages, Message{
		Role:       "tool",
		Name:       name,
		ToolCallID: toolCallID,
		Content:    content,
	})
	s.LastUpdated = time.Now()
}

// MarshalJSON implements custom JSON marshaling for the session
func (s *Chat) MarshalJSON() ([]byte, error) {
	type Alias Chat
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for the session
func (s *Chat) UnmarshalJSON(data []byte) error {
	type Alias Chat
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal session: %v", err)
	}
	return nil
}
