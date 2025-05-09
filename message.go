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

	// meta is not marshaled to LLM and other tools
	meta map[string]any `json:"-"`
}

// Meta represents metadata for a message.
type Meta struct{ msg *Message }

// Meta returns a Meta struct for the message.
func (m *Message) Meta() *Meta {
	return &Meta{m}
}

// Set sets a metadata value for the message.
// It initializes the underlying map if it's nil.
func (meta *Meta) Set(key string, value any) {
	if meta.msg.meta == nil {
		meta.msg.meta = make(map[string]any)
	}
	meta.msg.meta[key] = value
}

// Get retrieves a metadata value for the message.
// Returns nil if the key does not exist.
func (meta *Meta) Get(key string) any {
	if meta.msg.meta == nil {
		return nil
	}
	return meta.msg.meta[key]
}

// Keys returns a slice of all metadata keys in the message.
func (meta *Meta) Keys() []string {
	if meta.msg.meta == nil {
		return []string{}
	}
	keys := make([]string, 0, len(meta.msg.meta))
	for key := range meta.msg.meta {
		keys = append(keys, key)
	}
	return keys
}

// MarshalJSON implements the json.Marshaler interface.
func (meta *Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(meta.msg.meta)
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
	if err == nil {
		if err = json.Unmarshal(b, &p); err == nil {
			return p, nil
		}
	}
	return nil, err
}
