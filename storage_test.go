package aichat_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

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
	shouldErrorOnGet bool
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
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Create a new session and load the data
	loadedSession := &aichat.Chat{ID: "test-id", Options: aichat.Options{S3: s3}}
	err = loadedSession.Load(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify loaded data
	if loadedSession.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, loadedSession.ID)
	}

	if len(loadedSession.Messages) != len(session.Messages) {
		t.Errorf("Expected %d messages, got %d", len(session.Messages), len(loadedSession.Messages))
	}

	if loadedSession.Meta["test"] != "value" {
		t.Errorf("Expected metadata value 'value', got %v", loadedSession.Meta["test"])
	}

	err = loadedSession.Delete(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}
}

func TestStorageErrors(t *testing.T) {
	ctx := context.Background()

	// Test with nil S3
	session := &aichat.Chat{ID: "test-id"}

	if err := session.Save(ctx, "test-key"); err == nil {
		t.Error("Expected error when saving with nil S3")
	}

	if err := session.Load(ctx, "test-key"); err == nil {
		t.Error("Expected error when loading with nil S3")
	}

	t.Run("get error", func(t *testing.T) {
		s3 := &mockS3WithErrors{shouldErrorOnGet: true}
		chat := &aichat.Chat{Options: aichat.Options{S3: s3}}
		
		err := chat.Load(ctx, "test-key")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to get session from storage") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		s3 := &mockS3WithErrors{returnInvalidJSON: true}
		chat := &aichat.Chat{Options: aichat.Options{S3: s3}}
		
		err := chat.Load(ctx, "test-key")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to decode chat data") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		s3 := newMockS3()
		chat := &aichat.Chat{
			Options: aichat.Options{S3: s3},
			Messages: []aichat.Message{
				{
					Role: "user",
					Content: make(chan int), // channels cannot be marshaled to JSON
				},
			},
		}

		err := chat.Save(ctx, "test-key")
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to marshal chat data") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestNewStorage(t *testing.T) {
	s3 := newMockS3()
	opts := aichat.Options{S3: s3}

	storage := aichat.NewStorage(opts)
	if storage == nil {
		t.Fatal("Expected non-nil storage")
	}

	if storage.Options.S3 != s3 {
		t.Error("Storage options not set correctly")
	}
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
	if err := originalChat.Save(ctx, "test-key"); err != nil {
		t.Fatalf("Failed to save chat: %v", err)
	}

	// Now test loading the chat
	loadedChat, err := storage.Load(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to load chat: %v", err)
	}
	if loadedChat == nil {
		t.Fatal("Expected non-nil chat")
	}
	if loadedChat.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got %s", loadedChat.Key)
	}
	if loadedChat.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", loadedChat.ID)
	}
	if len(loadedChat.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(loadedChat.Messages))
	}
	if loadedChat.Options.S3 != s3 {
		t.Error("Chat options not set correctly")
	}

	// Test loading non-existent chat
	_, err = storage.Load(ctx, "non-existent-key")
	if err == nil {
		t.Error("Expected error when loading non-existent chat")
	}
}
