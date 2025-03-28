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

type s3message struct {
	*Message
	Meta map[string]any `json:"meta,omitempty"`
}

type s3chat struct {
	*Chat
	Messages []*s3message `json:"messages"`
}

// Load loads a chat from S3 storage
func (chat *Chat) Load(ctx context.Context, key string) error {
	if chat.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized")
	}

	reader, err := chat.Options.S3.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get session from storage: %w", err)
	}
	defer reader.Close()

	if err := json.NewDecoder(reader).Decode(&s3chat{Chat: chat}); err != nil {
		return fmt.Errorf("failed to decode chat data: %w", err)
	}

	return nil
}

// Save saves the session to S3 storage
func (chat *Chat) Save(ctx context.Context, key string) error {
	// Ensure S3 storage is configured in options.
	// We assume Options struct itself is not a pointer based on previous lint error.
	if chat.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized in options")
	}

	s3messages := []*s3message{}
	for _, msg := range chat.Messages {
		s3messages = append(s3messages, &s3message{Message: msg})
	}
	s3payload := s3chat{
		Chat:     chat,
		Messages: s3messages,
	}

	// Marshal the payload to JSON.
	data, err := json.Marshal(s3payload)
	if err != nil {
		return fmt.Errorf("failed to marshal chat data for S3: %w", err)
	}

	// Put the marshaled data into S3.
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
