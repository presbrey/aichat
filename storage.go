package aichat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// S3 represents a storage interface for sessions
type S3 interface {
	// Get retrieves data from storage
	Get(key string) (io.ReadCloser, error)
	// Put stores data
	Put(key string, data io.Reader) error
	// Delete deletes data from storage
	Delete(key string) error
}

// Load loads a session from S3 storage
func (s *Chat) Load(key string) error {
	if s.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	reader, err := s.Options.S3.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get session from storage: %v", err)
	}
	defer reader.Close()

	if err := json.NewDecoder(reader).Decode(s); err != nil {
		return fmt.Errorf("failed to decode session: %v", err)
	}

	return nil
}

// Save saves the session to S3 storage
func (s *Chat) Save(key string) error {
	if s.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}

	return s.Options.S3.Put(key, io.NopCloser(bytes.NewReader(data)))
}

// Delete deletes the session from S3 storage
func (s *Chat) Delete(key string) error {
	if s.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	return s.Options.S3.Delete(key)
}

// ChatStorage represents a storage interface for sessions
type ChatStorage struct {
	Options Options
}

// NewChatStorage creates a new session storage
func NewChatStorage(options Options) *ChatStorage {
	return &ChatStorage{
		Options: options,
	}
}

// Load loads a session from storage
func (s *ChatStorage) Load(key string) (*Chat, error) {
	c := NewChat(key, s.Options)
	return c, c.Load(key)
}
