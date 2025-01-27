package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// S3 represents a storage interface for sessions
type S3 interface {
	// Get retrieves data from storage
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Put stores data
	Put(ctx context.Context, key string, data io.Reader) error
	// Delete deletes data from storage
	Delete(ctx context.Context, key string) error
}

// Load loads a chat from S3 storage
func (s *Chat) Load(ctx context.Context, key string) error {
	if s.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	reader, err := s.Options.S3.Get(ctx, key)
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
func (s *Chat) Save(ctx context.Context, key string) error {
	if s.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}

	return s.Options.S3.Put(ctx, key, bytes.NewReader(data))
}

// Delete deletes the session from S3 storage
func (s *Chat) Delete(ctx context.Context, key string) error {
	if s.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	return s.Options.S3.Delete(ctx, key)
}

// ChatStorage represents a storage interface for sessions
type Storage struct {
	Options Options
}

// NewChatStorage creates a new session storage
func NewStorage(options Options) *Storage {
	return &Storage{
		Options: options,
	}
}

// Load loads a session from storage
func (s *Storage) Load(ctx context.Context, key string) (*Chat, error) {
	c := &Chat{Key: key, Options: s.Options}
	return c, c.Load(ctx, key)
}
