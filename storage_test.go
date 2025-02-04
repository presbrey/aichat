package aichat_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/presbrey/aichat"
)

// mockS3 implements the S3 interface for testing
type mockS3 struct {
	data map[string][]byte
}

func newMockS3() *mockS3 {
	return &mockS3{
		data: make(map[string][]byte),
	}
}

func (m *mockS3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	data, ok := m.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return io.NopCloser(strings.NewReader(string(data))), nil
}

func (m *mockS3) Put(ctx context.Context, key string, data io.Reader) error {
	b, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	m.data[key] = b
	return nil
}

func (m *mockS3) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// mockS3WithErrors is a mock S3 implementation that returns errors
type mockS3WithErrors struct {
	mockS3
	shouldErrorOnGet  bool
	returnInvalidJSON bool
}

func (m *mockS3WithErrors) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.shouldErrorOnGet {
		return nil, fmt.Errorf("mock get error")
	}
	if m.returnInvalidJSON {
		return io.NopCloser(strings.NewReader("invalid json")), nil
	}
	return m.mockS3.Get(ctx, key)
}

func TestChatStorage(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	session := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}

	// Add some test data
	session.AddUserContent("Test message")
	session.Meta = make(map[string]any)
	session.Meta["test"] = "value"

	// Test saving
	err := session.Save(ctx, "test-key")
	assert.NoError(t, err, "Failed to save session")

	// Create a new session and load the data
	loadedSession := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}
	err = loadedSession.Load(ctx, "test-key")
	assert.NoError(t, err, "Failed to load session")

	// Verify loaded data
	assert.Equal(t, session.ID, loadedSession.ID, "Session ID mismatch")
	assert.Equal(t, len(session.Messages), len(loadedSession.Messages), "Message count mismatch")
	assert.Equal(t, "value", loadedSession.Meta["test"], "Metadata value mismatch")

	err = loadedSession.Delete(ctx, "test-key")
	assert.NoError(t, err, "Failed to delete session")
}

func TestStorageErrors(t *testing.T) {
	ctx := context.Background()

	// Test with nil S3
	session := &aichat.Chat{ID: "test-id"}

	assert.Error(t, session.Save(ctx, "test-key"), "Expected error when saving with nil S3")
	assert.Error(t, session.Load(ctx, "test-key"), "Expected error when loading with nil S3")

	t.Run("get error", func(t *testing.T) {
		s3 := &mockS3WithErrors{shouldErrorOnGet: true}
		chat := &aichat.Chat{Options: aichat.Options{S3: s3}}

		err := chat.Load(ctx, "test-key")
		assert.Error(t, err, "Expected error")
		assert.Contains(t, err.Error(), "failed to get session from storage", "Unexpected error message")
	})

	t.Run("decode error", func(t *testing.T) {
		s3 := &mockS3WithErrors{returnInvalidJSON: true}
		chat := &aichat.Chat{Options: aichat.Options{S3: s3}}

		err := chat.Load(ctx, "test-key")
		assert.Error(t, err, "Expected error")
		assert.Contains(t, err.Error(), "failed to decode chat data", "Unexpected error message")
	})

	t.Run("marshal error", func(t *testing.T) {
		s3 := newMockS3()
		chat := &aichat.Chat{
			Options: aichat.Options{S3: s3},
			Messages: []*aichat.Message{
				{
					Role:    "user",
					Content: make(chan int), // channels cannot be marshaled to JSON
				},
			},
		}

		err := chat.Save(ctx, "test-key")
		assert.Error(t, err, "Expected error")
		assert.Contains(t, err.Error(), "failed to marshal chat data", "Unexpected error message")
	})
}

func TestNewStorage(t *testing.T) {
	s3 := newMockS3()
	opts := aichat.Options{S3: s3}

	storage := aichat.NewStorage(opts)
	assert.NotNil(t, storage, "Expected non-nil storage")
	assert.Equal(t, s3, storage.Options.S3, "Storage options not set correctly")
}

func TestStorageLoad(t *testing.T) {
	ctx := context.Background()
	s3 := newMockS3()
	opts := aichat.Options{S3: s3}
	storage := aichat.NewStorage(opts)

	// First create and save a chat
	originalChat := &aichat.Chat{
		Key:     "test-key",
		ID:      "test-id",
		Options: opts,
	}
	originalChat.AddUserContent("Hello")
	assert.NoError(t, originalChat.Save(ctx, "test-key"), "Failed to save chat")

	// Now test loading the chat
	loadedChat, err := storage.Load(ctx, "test-key")
	assert.NoError(t, err, "Failed to load chat")
	assert.NotNil(t, loadedChat, "Expected non-nil chat")
	assert.Equal(t, "test-key", loadedChat.Key, "Chat key mismatch")
	assert.Equal(t, "test-id", loadedChat.ID, "Chat ID mismatch")
	assert.Equal(t, 1, len(loadedChat.Messages), "Message count mismatch")
	assert.Equal(t, s3, loadedChat.Options.S3, "Chat options not set correctly")

	// Test loading non-existent chat
	_, err = storage.Load(ctx, "non-existent-key")
	assert.Error(t, err, "Expected error when loading non-existent chat")
}
