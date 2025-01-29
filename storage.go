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
func (chat *Chat) Load(ctx context.Context, key string) error {
	if chat.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	reader, err := chat.Options.S3.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get session from storage: %v", err)
	}
	defer reader.Close()

	if err := json.NewDecoder(reader).Decode(chat); err != nil {
		return fmt.Errorf("failed to decode chat data: %v", err)
	}

	return nil
}

// Save saves the session to S3 storage
func (chat *Chat) Save(ctx context.Context, key string) error {
	if chat.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	data, err := json.Marshal(chat)
	if err != nil {
		return fmt.Errorf("failed to marshal chat data: %v", err)
	}

	return chat.Options.S3.Put(ctx, key, bytes.NewReader(data))
}

// Delete deletes the session from S3 storage
func (chat *Chat) Delete(ctx context.Context, key string) error {
	if chat.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	return chat.Options.S3.Delete(ctx, key)
}

// Storage represents a storage interface for chat sessions
type Storage struct {
	Options Options
}

// NewStorage creates a new chat storage
func NewStorage(options Options) *Storage {
	return &Storage{
		Options: options,
	}
}

// Load loads a chat from storage
func (s *Storage) Load(ctx context.Context, key string) (*Chat, error) {
	c := &Chat{Key: key, Options: s.Options}
	return c, c.Load(ctx, key)
}
