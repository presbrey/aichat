package aichat

import (
	"encoding/json"
	"slices"
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
	Messages []*Message `json:"messages"`
	// Created is the timestamp when the chat session was created
	Created time.Time `json:"created"`
	// LastUpdated is the timestamp of the most recent message or modification
	LastUpdated time.Time `json:"last_updated"`
	// Meta stores arbitrary session-related data
	Meta map[string]any `json:"meta,omitempty"`
	// Options contains the configuration for these chat sessions
	Options Options `json:"-"`
}

// AddMessage adds a message to the chat
func (chat *Chat) AddMessage(message *Message) {
	if message == nil {
		return
	}
	chat.Messages = append(chat.Messages, message)
	chat.LastUpdated = time.Now()
}

// AddMessageOnce adds a message to the chat (idempotent)
func (chat *Chat) AddMessageOnce(message *Message) {
	if slices.Contains(chat.Messages, message) {
		return
	}
	chat.AddMessage(message)
}

// AddRoleContent adds a role and content to the chat
func (chat *Chat) AddRoleContent(role string, content any) *Message {
	m := &Message{
		Role:    role,
		Content: content,
	}
	chat.AddMessage(m)
	return m
}

// AddUserContent adds a user message to the chat
func (chat *Chat) AddUserContent(content any) *Message {
	return chat.AddRoleContent("user", content)
}

// AddAssistantContent adds an assistant message to the chat
func (chat *Chat) AddAssistantContent(content any) *Message {
	return chat.AddRoleContent("assistant", content)
}

// AddToolRawContent adds a raw content to the chat
func (chat *Chat) AddToolRawContent(name string, toolCallID string, content any) *Message {
	m := &Message{
		Role:       "tool",
		Name:       name,
		ToolCallID: toolCallID,
		Content:    content,
	}
	chat.AddMessage(m)
	return m
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
func (chat *Chat) AddAssistantToolCall(toolCalls []ToolCall) *Message {
	m := &Message{
		Role:      "assistant",
		ToolCalls: toolCalls,
	}
	chat.AddMessage(m)
	return m
}

// ClearMessages removes all messages from the chat
func (chat *Chat) ClearMessages() {
	chat.Messages = []*Message{}
	chat.LastUpdated = time.Now()
}

// LastMessage returns the last message in the chat
func (chat *Chat) LastMessage() *Message {
	if len(chat.Messages) == 0 {
		return nil
	}
	return chat.Messages[len(chat.Messages)-1]
}

// LastMessageByRole returns the last message in the chat by role
func (chat *Chat) LastMessageByRole(role string) *Message {
	if len(chat.Messages) == 0 {
		return nil
	}
	for i := len(chat.Messages) - 1; i >= 0; i-- {
		msg := chat.Messages[i]
		if msg.Role == role {
			return msg
		}
	}
	return nil
}

// LastMessageByType returns the last message in the chat with the given content type
func (chat *Chat) LastMessageByType(contentType string) *Message {
	for i := len(chat.Messages) - 1; i >= 0; i-- {
		msg := chat.Messages[i]
		if content, ok := msg.Content.(map[string]interface{}); ok {
			if t, ok := content["type"].(string); ok && t == contentType {
				return msg
			}
		}
	}
	return nil
}

// LastMessageRole returns the role of the last message in the chat
func (chat *Chat) LastMessageRole() string {
	msg := chat.LastMessage()
	if msg == nil {
		return ""
	}
	return msg.Role
}

// MessageCount returns the total number of messages in the chat
func (chat *Chat) MessageCount() int {
	return len(chat.Messages)
}

// MessageCountByRole returns the number of messages with a specific role
func (chat *Chat) MessageCountByRole(role string) int {
	count := 0
	for _, msg := range chat.Messages {
		if msg.Role == role {
			count++
		}
	}
	return count
}

// PopMessage removes and returns the last message from the chat
func (chat *Chat) PopMessage() *Message {
	if len(chat.Messages) == 0 {
		return nil
	}
	chat.LastUpdated = time.Now()
	msg := chat.Messages[len(chat.Messages)-1]
	chat.Messages = chat.Messages[:len(chat.Messages)-1]
	return msg
}

// PopMessageIfRole removes and returns the last message from the chat if it matches the role
func (chat *Chat) PopMessageIfRole(role string) *Message {
	if len(chat.Messages) == 0 {
		return nil
	}
	msg := chat.Messages[len(chat.Messages)-1]
	if msg.Role == role {
		chat.LastUpdated = time.Now()
		chat.Messages = chat.Messages[:len(chat.Messages)-1]
		return msg
	}
	return nil
}

// Range iterates through messages
func (chat *Chat) Range(fn func(msg *Message) error) error {
	for _, msg := range chat.Messages {
		if err := fn(msg); err != nil {
			return err
		}
	}
	return nil
}

// RangeByRole iterates through messages with a specific role
func (chat *Chat) RangeByRole(role string, fn func(msg *Message) error) error {
	for _, msg := range chat.Messages {
		if msg.Role == role {
			if err := fn(msg); err != nil {
				return err
			}
		}
	}
	return nil
}

// RemoveLastMessage removes and returns the last message from the chat
func (chat *Chat) RemoveLastMessage() *Message {
	return chat.PopMessage()
}

// SetSystemContent sets or updates the system message at the beginning of the chat.
// If the first message is a system message, it updates its content.
// Otherwise, it inserts a new system message at the beginning.
func (chat *Chat) SetSystemContent(content any) *Message {
	m := &Message{
		Role:    "system",
		Content: content,
	}
	return chat.SetSystemMessage(m)
}

// SetSystemMessage sets the system message at the beginning of the chat
func (chat *Chat) SetSystemMessage(msg *Message) *Message {
	if len(chat.Messages) > 0 && chat.Messages[0].Role == "system" {
		chat.Messages[0] = msg
	} else {
		chat.UnshiftMessages(msg)
	}
	chat.LastUpdated = time.Now()
	return msg
}

// ShiftMessages shifts all messages to the left by one index
func (chat *Chat) ShiftMessages() *Message {
	if len(chat.Messages) == 0 {
		return nil
	}
	chat.LastUpdated = time.Now()
	msg := chat.Messages[0]
	chat.Messages = chat.Messages[1:]
	return msg
}

// UnshiftMessages unshifts all messages to the right by one index
func (chat *Chat) UnshiftMessages(msg *Message) {
	chat.LastUpdated = time.Now()
	if len(chat.Messages) == 0 {
		chat.Messages = []*Message{msg}
	} else {
		chat.Messages = append([]*Message{msg}, chat.Messages...)
	}
}
