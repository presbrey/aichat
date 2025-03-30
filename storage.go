package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"
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
	ID          string         `json:"id,omitempty"`
	Messages    []*s3message   `json:"messages"`
	Meta        map[string]any `json:"meta,omitempty"`
	Created     time.Time      `json:"created"`
	LastUpdated time.Time      `json:"last_updated"`
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

	// Decode into a temporary structure first
	var s3payload s3chat
	if err := json.NewDecoder(reader).Decode(&s3payload); err != nil {
		return fmt.Errorf("failed to decode chat data: %w", err)
	}

	// Restore all fields
	chat.ID = s3payload.ID
	chat.Created = s3payload.Created
	chat.LastUpdated = s3payload.LastUpdated
	chat.Meta = s3payload.Meta

	// Reconstruct messages and restore metadata
	loadedMessages := make([]*Message, 0, len(s3payload.Messages))
	for _, s3msg := range s3payload.Messages {
		// Start with the base message decoded within s3message
		msg := s3msg.Message
		if msg == nil {
			// Handle cases where the embedded message might be nil, though unlikely if saved correctly
			continue
		}

		// Restore metadata from s3msg.Meta
		msg.meta = s3msg.Meta
		loadedMessages = append(loadedMessages, msg)
	}
	chat.Messages = loadedMessages // Assign the reconstructed messages

	return nil
}

// Save saves the session to S3 storage
func (chat *Chat) Save(ctx context.Context, key string) error {
	// Ensure S3 storage is configured in options.
	// We assume Options struct itself is not a pointer based on previous lint error.
	if chat.Options.S3 == nil {
		return fmt.Errorf("s3 storage not initialized in options")
	}

	// Convert Messages to s3message format, including metadata
	s3messages := make([]*s3message, 0, len(chat.Messages))
	for _, msg := range chat.Messages {
		s3msg := &s3message{Message: msg, Meta: msg.meta}
		s3messages = append(s3messages, s3msg)
	}

	// Prepare the payload including explicitly chosen chat fields and converted messages
	s3payload := s3chat{
		ID:          chat.ID,
		Meta:        chat.Meta,
		Messages:    s3messages,
		Created:     chat.Created,
		LastUpdated: chat.LastUpdated,
	}

	// Marshal the payload to JSON.
	data, err := json.Marshal(s3payload)
	if err != nil {
		return fmt.Errorf("failed to marshal chat data for S3: %w", err)
	}

	// Put the data into S3 storage
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
