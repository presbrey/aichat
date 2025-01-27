package aichat

import (
	"encoding/json"
)

// Message represents a chat message in the session
type Message struct {
	Role       string     `json:"role"`
	Content    any        `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	Name       string     `json:"name,omitempty"`         // For tool responses
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool responses
}

// ContentString returns the content of the message as a string if the content is a simple string.
// Returns an empty string if the content is not a string type (e.g., multipart content).
func (m *Message) ContentString() string {
	v, _ := m.Content.(string)
	return v
}

// Part represents a part of a message content that can be either text or image
type Part struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL struct {
		URL    string `json:"url"`
		Detail string `json:"detail,omitempty"`
	} `json:"image_url,omitempty"`
}

// ContentParts returns the parts of a multipart message content (text and images).
// Returns nil for both parts and error if the content is not multipart.
// Returns error if the content cannot be properly marshaled/unmarshaled.
func (m *Message) ContentParts() ([]*Part, error) {
	d, ok := m.Content.([]any)
	if !ok {
		return nil, nil
	}
	p := []*Part{}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
