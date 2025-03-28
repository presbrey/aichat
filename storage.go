package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
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

	// Decode into a temporary structure first
	var s3payload s3chat
	if err := json.NewDecoder(reader).Decode(&s3payload); err != nil {
		return fmt.Errorf("failed to decode chat data: %w", err)
	}

	// Now, restore the Chat fields (excluding Messages for now)
	// This assumes s3chat embeds *Chat and we want to copy non-message fields.
	// If s3chat only contains Messages, this part might need adjustment.
	*chat = *s3payload.Chat // Potential issue: overwrites original chat fields if s3payload.Chat is modified

	// Reconstruct messages and restore metadata
	loadedMessages := make([]*Message, 0, len(s3payload.Messages))
	for _, s3msg := range s3payload.Messages {
		// Start with the base message decoded within s3message
		msg := s3msg.Message
		if msg == nil {
			// Handle cases where the embedded message might be nil, though unlikely if saved correctly
			continue
		}

		// Restore metadata if it exists
		if len(s3msg.Meta) > 0 {
			msg.meta = &sync.Map{}
			for k, v := range s3msg.Meta {
				msg.meta.Store(k, v)
			}
		} else {
			// Ensure meta is nil if no metadata was saved
			msg.meta = nil
		}
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
		s3msg := &s3message{Message: msg}
		if msg.meta != nil {
			s3msg.Meta = make(map[string]any)
			msg.meta.Range(func(key, value any) bool {
				s3msg.Meta[key.(string)] = value // Assuming keys are strings
				return true
			})
		}
		s3messages = append(s3messages, s3msg)
	}

	// Prepare the payload including the chat itself and the converted messages
	s3payload := s3chat{
		Chat:     chat,       // Embed the current chat state
		Messages: s3messages, // Use the converted messages with metadata
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
