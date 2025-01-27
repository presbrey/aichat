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

// ContentString returns the content of the message if its a simple string (not multipart)
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

// ContentParts returns the parts of the message content if its multipart content
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
